package redactyl

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "sarif",
		Short: "SARIF utilities",
	}
	rootCmd.AddCommand(cmd)

	view := &cobra.Command{
		Use:   "view <file>",
		Short: "Pretty-print SARIF results",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			path := args[0]
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()
			var obj any
			dec := json.NewDecoder(f)
			if err := dec.Decode(&obj); err != nil {
				return err
			}
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			if err := enc.Encode(obj); err != nil {
				return err
			}
			return nil
		},
	}
	cmd.AddCommand(view)
}
