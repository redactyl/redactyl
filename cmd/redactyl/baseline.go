package redactyl

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/redactyl/redactyl/internal/config"
	"github.com/redactyl/redactyl/internal/engine"
	"github.com/redactyl/redactyl/internal/report"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "baseline",
		Short: "Manage baselines",
	}

	update := &cobra.Command{
		Use:   "update",
		Short: "Update baseline from current scan",
		RunE: func(cmd *cobra.Command, _ []string) error {
			abs, _ := filepath.Abs(".")
			var gcfg, lcfg config.FileConfig
			if c, err := config.LoadGlobal(); err == nil {
				gcfg = c
			}
			if c, err := config.LoadLocal(abs); err == nil {
				lcfg = c
			}

			budget, globalBudget := resolveBudgets(0, lcfg, gcfg, 0)

			cfg := engine.Config{
				Root:                 abs,
				Threads:              pickInt(flagThreads, lcfg.Threads, gcfg.Threads),
				IncludeGlobs:         pickString("", lcfg.Include, gcfg.Include),
				ExcludeGlobs:         pickString("", lcfg.Exclude, gcfg.Exclude),
				MaxBytes:             pickInt64(flagMaxBytes, lcfg.MaxBytes, gcfg.MaxBytes),
				EnableDetectors:      pickString("", lcfg.Enable, gcfg.Enable),
				DisableDetectors:     pickString("", lcfg.Disable, gcfg.Disable),
				MinConfidence:        pickFloat(0, lcfg.MinConfidence, gcfg.MinConfidence),
				DefaultExcludes:      pickBool(flagDefaultExcludes, lcfg.DefaultExcludes, gcfg.DefaultExcludes),
				NoColor:              pickBool(false, lcfg.NoColor, gcfg.NoColor),
				ScanArchives:         pickBool(false, lcfg.Archives, gcfg.Archives),
				ScanContainers:       pickBool(false, lcfg.Containers, gcfg.Containers),
				ScanIaC:              pickBool(false, lcfg.IaC, gcfg.IaC),
				ScanHelm:             pickBool(false, lcfg.Helm, gcfg.Helm),
				ScanK8s:              pickBool(false, lcfg.K8s, gcfg.K8s),
				MaxArchiveBytes:      pickInt64(0, lcfg.MaxArchiveBytes, gcfg.MaxArchiveBytes),
				MaxEntries:           pickInt(0, lcfg.MaxEntries, gcfg.MaxEntries),
				MaxDepth:             pickInt(0, lcfg.MaxDepth, gcfg.MaxDepth),
				ScanTimeBudget:       budget,
				GlobalArtifactBudget: globalBudget,
				GitleaksConfig:       mergeGitleaksConfig(gcfg, lcfg),
			}
			results, err := engine.Scan(cfg)
			if err != nil {
				return err
			}
			if err := report.SaveBaseline("redactyl.baseline.json", results); err != nil {
				return err
			}
			fmt.Fprintln(os.Stdout, "Baseline updated.")
			return nil
		},
	}

	rootCmd.AddCommand(cmd)
	cmd.AddCommand(update)
}
