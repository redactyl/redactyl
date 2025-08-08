package redactyl

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Run: func(_ *cobra.Command, _ []string) {
			rev := ""
			ts := ""
			if info, ok := debug.ReadBuildInfo(); ok {
				for _, s := range info.Settings {
					if s.Key == "vcs.revision" {
						rev = s.Value
					}
					if s.Key == "vcs.time" {
						ts = s.Value
					}
				}
			}
			if rev != "" || ts != "" {
				fmt.Printf("%s (commit %s, built %s)\n", version, short(rev), ts)
				return
			}
			fmt.Println(version)
		},
	}
	rootCmd.AddCommand(cmd)
}

func short(s string) string {
	if len(s) > 7 {
		return s[:7]
	}
	return s
}
