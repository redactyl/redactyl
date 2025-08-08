package redactyl

import (
	"fmt"

	"github.com/accrava/redactyl/internal/engine"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "detectors",
		Short: "List available detectors",
		Run: func(_ *cobra.Command, _ []string) {
			for _, id := range engine.DetectorIDs() {
				fmt.Println(id)
			}
		},
	}
	rootCmd.AddCommand(cmd)
}
