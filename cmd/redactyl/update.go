package redactyl

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update redactyl to the latest release",
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := selfUpdate(); err != nil {
				return fmt.Errorf("update failed: %w", err)
			}
			fmt.Println("Updated to latest. Re-run your command.")
			return nil
		},
	}
	rootCmd.AddCommand(cmd)
}
