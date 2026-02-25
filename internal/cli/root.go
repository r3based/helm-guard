package cli

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "helm-guard",
	Short: "Analyze rendered Helm manifests",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(analyzeCmd)
	rootCmd.AddCommand(rulesCmd)
}
