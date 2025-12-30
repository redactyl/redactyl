package redactyl

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/varalys/redactyl/internal/engine"
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
