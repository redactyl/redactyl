package gitleaks

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	githubLatestReleaseURL = "https://api.github.com/repos/gitleaks/gitleaks/releases/latest"
	httpClientTimeout      = 10 * time.Second
)

var (
	errBinaryNotFound  = errors.New("gitleaks binary not found")
	errVersionMismatch = errors.New("gitleaks binary version mismatch")
)

// BinaryManager handles detection and installation of the Gitleaks binary.
type BinaryManager struct {
	customPath string
	cachePath  string
}

// NewBinaryManager creates a new binary manager.
func NewBinaryManager(customPath string) *BinaryManager {
	homeDir, _ := os.UserHomeDir()
	cachePath := filepath.Join(homeDir, ".redactyl", "bin")
	return &BinaryManager{
		customPath: customPath,
		cachePath:  cachePath,
	}
}

// Find locates a Gitleaks binary honoring an expected version (when provided).
// Search order:
//  1. Custom path (if configured)
//  2. $PATH
//  3. Versioned cache (~/.redactyl/bin/gitleaks-<version>)
//  4. Legacy cache (~/.redactyl/bin/gitleaks)
func (bm *BinaryManager) Find(expectedVersion string) (string, error) {
	normalized := normalizeVersion(expectedVersion)

	if bm.customPath != "" {
		if _, err := os.Stat(bm.customPath); err == nil {
			if normalized != "" {
				actual, err := bm.Version(bm.customPath)
				if err != nil {
					return "", fmt.Errorf("failed to check version for custom gitleaks binary %s: %w", bm.customPath, err)
				}
				if normalizeVersion(actual) != normalized {
					return "", fmt.Errorf("%w: custom gitleaks binary %s reports %s (expected %s)", errVersionMismatch, bm.customPath, actual, normalized)
				}
			}
			return bm.customPath, nil
		}
		return "", fmt.Errorf("custom gitleaks path not found: %s", bm.customPath)
	}

	if path, err := exec.LookPath("gitleaks"); err == nil {
		if normalized != "" {
			actual, err := bm.Version(path)
			if err != nil {
				return "", fmt.Errorf("failed to determine gitleaks version from PATH (%s): %w", path, err)
			}
			if normalizeVersion(actual) != normalized {
				return "", fmt.Errorf("%w: gitleaks on PATH (%s) reports %s (expected %s)", errVersionMismatch, path, actual, normalized)
			}
		}
		return path, nil
	}

	if normalized != "" {
		versioned := bm.versionedPath(normalized)
		if fileExists(versioned) {
			actual, err := bm.Version(versioned)
			if err != nil {
				return "", fmt.Errorf("failed to determine cached gitleaks version (%s): %w", versioned, err)
			}
			if normalizeVersion(actual) != normalized {
				return "", fmt.Errorf("%w: cached gitleaks %s reports %s (expected %s)", errVersionMismatch, versioned, actual, normalized)
			}
			return versioned, nil
		}
	}

	legacy := bm.legacyCachePath()
	if fileExists(legacy) {
		if normalized != "" {
			actual, err := bm.Version(legacy)
			if err != nil {
				return "", fmt.Errorf("failed to determine legacy cached gitleaks version (%s): %w", legacy, err)
			}
			if normalizeVersion(actual) != normalized {
				return "", fmt.Errorf("%w: cached gitleaks %s reports %s (expected %s)", errVersionMismatch, legacy, actual, normalized)
			}
		}
		return legacy, nil
	}

	searchPath := legacy
	if normalized != "" {
		searchPath = bm.versionedPath(normalized)
	}
	return "", fmt.Errorf("%w: searched cache path %s", errBinaryNotFound, searchPath)
}

// Version runs gitleaks --version and parses the output.
func (bm *BinaryManager) Version(binaryPath string) (string, error) {
	cmd := exec.Command(binaryPath, "version") // #nosec G204
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get gitleaks version: %w", err)
	}

	version := strings.TrimSpace(string(output))
	version = strings.TrimPrefix(version, "version ")
	if len(version) > 0 && (version[0] == 'v' || version[0] == 'V') {
		version = version[1:]
	}
	if lines := strings.Split(version, "\n"); len(lines) > 0 {
		version = strings.TrimSpace(lines[0])
	}
	return version, nil
}

// Download fetches the requested version (or latest) and stores it in the cache.
func (bm *BinaryManager) Download(version string) (string, error) {
	client := &http.Client{Timeout: httpClientTimeout}
	normalized, withPrefix, err := bm.resolveVersion(version, client)
	if err != nil {
		return "", err
	}

	platform := GetPlatform()
	versionNoPrefix := strings.TrimPrefix(withPrefix, "v")

	var downloadURL string
	var isZip bool
	if runtime.GOOS == "windows" {
		downloadURL = fmt.Sprintf("https://github.com/gitleaks/gitleaks/releases/download/%s/gitleaks_%s_%s.zip",
			withPrefix, versionNoPrefix, platform)
		isZip = true
	} else {
		downloadURL = fmt.Sprintf("https://github.com/gitleaks/gitleaks/releases/download/%s/gitleaks_%s_%s.tar.gz",
			withPrefix, versionNoPrefix, platform)
		isZip = false
	}

	if err := os.MkdirAll(bm.cachePath, 0o755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}

	resp, err := client.Get(downloadURL) // #nosec G107
	if err != nil {
		return "", fmt.Errorf("failed to download gitleaks from %s: %w", downloadURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download gitleaks: HTTP %d from %s", resp.StatusCode, downloadURL)
	}

	destPath := bm.versionedPath(normalized)
	tmpFile, err := os.CreateTemp(bm.cachePath, "gitleaks-download-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file for gitleaks download: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer func() { _ = os.Remove(tmpPath) }()

	if isZip {
		if err := extractFromZip(resp.Body, executableName(), tmpPath); err != nil {
			return "", fmt.Errorf("failed to extract gitleaks: %w", err)
		}
	} else {
		if err := extractFromTarGz(resp.Body, executableName(), tmpPath); err != nil {
			return "", fmt.Errorf("failed to extract gitleaks: %w", err)
		}
	}

	if runtime.GOOS != "windows" {
		if err := os.Chmod(tmpPath, 0o755); err != nil {
			return "", fmt.Errorf("failed to make gitleaks executable: %w", err)
		}
	}

	if err := os.Rename(tmpPath, destPath); err != nil {
		return "", fmt.Errorf("failed to move gitleaks into cache: %w", err)
	}

	actual, err := bm.Version(destPath)
	if err != nil {
		if removeErr := os.Remove(destPath); removeErr != nil {
			return "", fmt.Errorf("failed to verify downloaded gitleaks version: %w (cleanup error: %v)", err, removeErr)
		}
		return "", fmt.Errorf("failed to verify downloaded gitleaks version: %w", err)
	}
	if normalizeVersion(actual) != normalized {
		if removeErr := os.Remove(destPath); removeErr != nil {
			return "", fmt.Errorf("%w: downloaded binary reports %s (expected %s); cleanup error: %v", errVersionMismatch, actual, normalized, removeErr)
		}
		return "", fmt.Errorf("%w: downloaded binary reports %s (expected %s)", errVersionMismatch, actual, normalized)
	}

	if err := bm.copyToLegacy(destPath); err != nil {
		return "", err
	}

	return destPath, nil
}

// GetPlatform returns the platform identifier for Gitleaks releases.
func GetPlatform() string {
	osName := runtime.GOOS
	arch := runtime.GOARCH
	switch arch {
	case "amd64":
		arch = "x64"
	case "386":
		arch = "x32"
	}
	return fmt.Sprintf("%s_%s", osName, arch)
}

func (bm *BinaryManager) resolveVersion(requested string, client *http.Client) (normalized string, withPrefix string, err error) {
	req := strings.TrimSpace(requested)
	if req == "" || strings.EqualFold(req, "latest") {
		tag, err := getLatestVersion(client)
		if err != nil {
			return "", "", fmt.Errorf("failed to resolve latest gitleaks version: %w", err)
		}
		normalized = normalizeVersion(tag)
		withPrefix = "v" + normalized
		return normalized, withPrefix, nil
	}

	normalized = normalizeVersion(req)
	if normalized == "" {
		return "", "", errors.New("invalid gitleaks version requested")
	}
	withPrefix = "v" + normalized
	return normalized, withPrefix, nil
}

func (bm *BinaryManager) versionedPath(normalizedVersion string) string {
	name := "gitleaks"
	if normalizedVersion != "" {
		name = fmt.Sprintf("gitleaks-%s", normalizedVersion)
	}
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return filepath.Join(bm.cachePath, name)
}

func (bm *BinaryManager) legacyCachePath() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(bm.cachePath, "gitleaks.exe")
	}
	return filepath.Join(bm.cachePath, "gitleaks")
}

func (bm *BinaryManager) copyToLegacy(source string) error {
	legacy := bm.legacyCachePath()
	if legacy == source {
		return nil
	}
	input, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("failed to open gitleaks binary for copying: %w", err)
	}
	defer func() {
		_ = input.Close()
	}()

	output, err := os.Create(legacy)
	if err != nil {
		return fmt.Errorf("failed to create legacy gitleaks binary: %w", err)
	}
	defer func() {
		_ = output.Close()
	}()

	if _, err := io.Copy(output, input); err != nil {
		return fmt.Errorf("failed to copy gitleaks binary: %w", err)
	}

	if runtime.GOOS != "windows" {
		if err := os.Chmod(legacy, 0o755); err != nil {
			return fmt.Errorf("failed to set permissions on legacy gitleaks: %w", err)
		}
	}
	return nil
}

func executableName() string {
	if runtime.GOOS == "windows" {
		return "gitleaks.exe"
	}
	return "gitleaks"
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func normalizeVersion(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return ""
	}
	if v[0] == 'v' || v[0] == 'V' {
		v = v[1:]
	}
	return v
}

// getLatestVersion fetches the latest release version from GitHub API.
func getLatestVersion(client *http.Client) (string, error) {
	req, err := http.NewRequest(http.MethodGet, githubLatestReleaseURL, nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	const tagPrefix = `"tag_name":"`
	start := strings.Index(string(body), tagPrefix)
	if start == -1 {
		return "", fmt.Errorf("could not find tag_name in GitHub response")
	}
	start += len(tagPrefix)
	end := strings.Index(string(body[start:]), `"`)
	if end == -1 {
		return "", fmt.Errorf("malformed tag_name in GitHub response")
	}

	return string(body[start : start+end]), nil
}

// extractFromTarGz extracts a single file from a tar.gz archive.
func extractFromTarGz(r io.Reader, filename, destPath string) error {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gzr.Close() //nolint:errcheck

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if strings.HasSuffix(header.Name, filename) {
			outFile, err := os.Create(destPath)
			if err != nil {
				return err
			}
			defer outFile.Close() //nolint:errcheck

			if _, err := io.Copy(outFile, tr); err != nil {
				return err
			}
			return nil
		}
	}

	return fmt.Errorf("file %s not found in archive", filename)
}

// extractFromZip extracts a single file from a zip archive.
func extractFromZip(r io.Reader, filename, destPath string) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	zr, err := zip.NewReader(&bytesReaderAt{data: data}, int64(len(data)))
	if err != nil {
		return err
	}

	for _, f := range zr.File {
		if strings.HasSuffix(f.Name, filename) {
			rc, err := f.Open()
			if err != nil {
				return err
			}
			defer rc.Close() //nolint:errcheck

			outFile, err := os.Create(destPath)
			if err != nil {
				return err
			}
			defer outFile.Close() //nolint:errcheck

			if _, err := io.Copy(outFile, rc); err != nil {
				return err
			}
			return nil
		}
	}

	return fmt.Errorf("file %s not found in archive", filename)
}

type bytesReaderAt struct {
	data []byte
}

func (b *bytesReaderAt) ReadAt(p []byte, off int64) (n int, err error) {
	if off < 0 {
		return 0, fmt.Errorf("negative offset")
	}
	if off >= int64(len(b.data)) {
		return 0, io.EOF
	}
	n = copy(p, b.data[off:])
	if n < len(p) {
		err = io.EOF
	}
	return
}
