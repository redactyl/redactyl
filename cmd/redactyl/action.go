package redactyl

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "action",
		Short: "GitHub Action helpers",
	}
	rootCmd.AddCommand(cmd)

	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Write a GitHub Action workflow for Redactyl",
		RunE: func(_ *cobra.Command, _ []string) error {
			dir := filepath.Join(".github", "workflows")
			if err := os.MkdirAll(dir, 0755); err != nil {
				return err
			}
			path := filepath.Join(dir, "redactyl.yml")
			content := `name: Redactyl Scan
on: [push, pull_request]
jobs:
  scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: 'stable'
      - run: go build -o bin/redactyl .
      - run: ./bin/redactyl scan --sarif > redactyl.sarif.json
      - uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: redactyl.sarif.json
`
			if err := os.WriteFile(path, []byte(content), 0644); err != nil {
				return err
			}
			fmt.Println("Wrote", path)
			return nil
		},
	}
	cmd.AddCommand(initCmd)
}
