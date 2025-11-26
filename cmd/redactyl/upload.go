package redactyl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/redactyl/redactyl/internal/git"
	"github.com/redactyl/redactyl/internal/types"
	"github.com/redactyl/redactyl/pkg/core"
)

const uploadSchemaVersion = "1"

type uploadEnvelope struct {
	Tool     string         `json:"tool"`
	Version  string         `json:"version"`
	Schema   string         `json:"schema_version"`
	Repo     string         `json:"repo,omitempty"`
	Commit   string         `json:"commit,omitempty"`
	Branch   string         `json:"branch,omitempty"`
	Findings []core.Finding `json:"findings"`
}

func uploadFindings(rootPath, url, token string, noMeta bool, findings []core.Finding) error {
	if len(findings) == 0 {
		return nil
	}
	env := uploadEnvelope{Tool: "redactyl", Version: version, Schema: uploadSchemaVersion, Findings: findings}
	if !noMeta {
		repo, commit, branch := git.RepoMetadata(rootPath)
		env.Repo, env.Commit, env.Branch = repo, commit, branch
	}
	body, err := json.Marshal(env)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	httpClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("upload status %d", resp.StatusCode)
	}
	return nil
}

func convertFindings(in []types.Finding) []core.Finding {
	out := make([]core.Finding, len(in))
	for i := range in {
		out[i] = core.Finding(in[i])
	}
	return out
}
