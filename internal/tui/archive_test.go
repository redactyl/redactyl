package tui

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

func createTestZip(t *testing.T, content string) string {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	f, err := w.Create("test.txt")
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.Write([]byte(content))
	if err != nil {
		t.Fatal(err)
	}
	err = w.Close()
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(t.TempDir(), "test.zip")
	err = os.WriteFile(path, buf.Bytes(), 0644)
	if err != nil {
		t.Fatal(err)
	}
	return path
}

func createTestTar(t *testing.T, content string) string {
	buf := new(bytes.Buffer)
	w := tar.NewWriter(buf)
	hdr := &tar.Header{
		Name: "test.txt",
		Mode: 0600,
		Size: int64(len(content)),
	}
	if err := w.WriteHeader(hdr); err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(t.TempDir(), "test.tar")
	err := os.WriteFile(path, buf.Bytes(), 0644)
	if err != nil {
		t.Fatal(err)
	}
	return path
}

func createTestTarGz(t *testing.T, content string) string {
	buf := new(bytes.Buffer)
	gw := gzip.NewWriter(buf)
	tw := tar.NewWriter(gw)

	hdr := &tar.Header{
		Name: "test.txt",
		Mode: 0600,
		Size: int64(len(content)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}

	tw.Close()
	gw.Close()

	path := filepath.Join(t.TempDir(), "test.tar.gz")
	err := os.WriteFile(path, buf.Bytes(), 0644)
	if err != nil {
		t.Fatal(err)
	}
	return path
}

func TestExtractFromZip(t *testing.T) {
	content := "secret data"
	path := createTestZip(t, content)

	data, err := extractFromArchive(path, "test.txt")
	if err != nil {
		t.Fatalf("Failed to extract zip: %v", err)
	}
	if string(data) != content {
		t.Errorf("Expected %q, got %q", content, string(data))
	}
}

func TestExtractFromTar(t *testing.T) {
	content := "secret data"
	path := createTestTar(t, content)

	data, err := extractFromArchive(path, "test.txt")
	if err != nil {
		t.Fatalf("Failed to extract tar: %v", err)
	}
	if string(data) != content {
		t.Errorf("Expected %q, got %q", content, string(data))
	}
}

func TestExtractFromTarGz(t *testing.T) {
	content := "secret data"
	path := createTestTarGz(t, content)

	data, err := extractFromArchive(path, "test.txt")
	if err != nil {
		t.Fatalf("Failed to extract tar.gz: %v", err)
	}
	if string(data) != content {
		t.Errorf("Expected %q, got %q", content, string(data))
	}
}

func TestExtractFromGz(t *testing.T) {
	content := "raw data"
	buf := new(bytes.Buffer)
	w := gzip.NewWriter(buf)
	w.Write([]byte(content))
	w.Close()

	path := filepath.Join(t.TempDir(), "test.gz")
	os.WriteFile(path, buf.Bytes(), 0644)

	data, err := extractFromArchive(path, "") // Empty internal path for direct gz
	if err != nil {
		t.Fatalf("Failed to extract gz: %v", err)
	}
	if string(data) != content {
		t.Errorf("Expected %q, got %q", content, string(data))
	}
}

func TestExtractNestedArchive(t *testing.T) {
	// Create inner zip
	innerBuf := new(bytes.Buffer)
	w := zip.NewWriter(innerBuf)
	f, _ := w.Create("inner.txt")
	f.Write([]byte("secret nested"))
	w.Close()

	// Create outer zip containing inner zip
	outerBuf := new(bytes.Buffer)
	w2 := zip.NewWriter(outerBuf)
	f2, _ := w2.Create("inner.zip")
	f2.Write(innerBuf.Bytes())
	w2.Close()

	path := filepath.Join(t.TempDir(), "outer.zip")
	os.WriteFile(path, outerBuf.Bytes(), 0644)

	// Test extraction: outer.zip -> inner.zip -> inner.txt
	// The virtual path logic splits by "::", so we pass the internal path
	data, err := extractFromArchive(path, "inner.zip::inner.txt")
	if err != nil {
		t.Fatalf("Failed to extract nested zip: %v", err)
	}
	if string(data) != "secret nested" {
		t.Errorf("Expected 'secret nested', got %q", string(data))
	}
}
