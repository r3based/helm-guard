package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var valuesFiles []string

var analyzeCmd = &cobra.Command{
	Use:   "analyze <chartPath>",
	Args:  cobra.ExactArgs(1),
	Short: "Render chart and analyze manifests",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("analyze:", args[0])
		fmt.Println("values:", valuesFiles)
		return nil
	},
}

func init() {
	analyzeCmd.Flags().StringArrayVarP(&valuesFiles, "values", "f", nil, "Values files")
}
