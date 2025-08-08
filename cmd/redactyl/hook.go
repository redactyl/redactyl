package redactyl

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "hook",
		Short: "Manage git hooks",
	}
	rootCmd.AddCommand(cmd)

	install := &cobra.Command{
		Use:   "install",
		Short: "Install git hooks",
	}
	cmd.AddCommand(install)

	preCommit := &cobra.Command{
		Use:   "--pre-commit",
		Short: "Install pre-commit hook for staged scan",
		RunE: func(_ *cobra.Command, _ []string) error {
			hookDir := filepath.Join(".git", "hooks")
			if _, err := os.Stat(hookDir); os.IsNotExist(err) {
				return fmt.Errorf("not a git repository (missing .git/hooks)")
			}
			hookPath := filepath.Join(hookDir, "pre-commit")
			content := "#!/bin/sh\n\nredactyl scan --staged\n"
			if err := os.WriteFile(hookPath, []byte(content), 0755); err != nil {
				return err
			}
			fmt.Println("Installed pre-commit hook -> .git/hooks/pre-commit")
			return nil
		},
	}
	// Allow calling as: redactyl hook install --pre-commit
	install.Flags().Bool("pre-commit", false, "install pre-commit hook")
	install.RunE = func(cmd *cobra.Command, _ []string) error {
		if ok, _ := cmd.Flags().GetBool("pre-commit"); ok {
			return preCommit.RunE(nil, nil)
		}
		return fmt.Errorf("specify --pre-commit")
	}

	// backwards compatible alias: redactyl hook install pre-commit
	installPre := &cobra.Command{
		Use:   "pre-commit",
		Short: "Install pre-commit hook for staged scan",
		RunE:  preCommit.RunE,
	}
	install.AddCommand(installPre)
}
