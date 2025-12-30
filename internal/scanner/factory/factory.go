package factory

import (
	"fmt"

	"github.com/varalys/redactyl/internal/config"
	"github.com/varalys/redactyl/internal/scanner"
	"github.com/varalys/redactyl/internal/scanner/gitleaks"
)

// Config is the subset of configuration needed to create a scanner.
type Config struct {
	Root           string
	GitleaksConfig config.GitleaksConfig
}

func New(cfg Config) (scanner.Scanner, error) {
	if cfg.GitleaksConfig.GetConfigPath() == "" {
		if detected := gitleaks.DetectConfigPath(cfg.Root); detected != "" {
			cfgPath := detected
			cfg.GitleaksConfig.ConfigPath = &cfgPath
		}
	}

	scnr, err := gitleaks.NewScanner(cfg.GitleaksConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create gitleaks scanner: %w", err)
	}

	return scnr, nil
}

func DefaultDetectors() []string {
	return []string{
		"github-pat", "github-fine-grained-pat", "github-oauth", "github-app-token",
		"aws-access-key", "aws-secret-key", "aws-mws-key",
		"stripe-access-token", "stripe-secret-key",
		"slack-webhook-url", "slack-bot-token", "slack-app-token",
		"google-api-key", "google-oauth", "gcp-service-account",
		"gitlab-pat", "gitlab-pipeline-token", "gitlab-runner-token",
		"sendgrid-api-key",
		"openai-api-key",
		"anthropic-api-key",
		"npm-access-token",
		"pypi-token",
		"docker-config-auth",
		"jwt",
		"private-key",
		"generic-api-key",
	}
}
