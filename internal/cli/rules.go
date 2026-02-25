package cli

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"github.com/r3based/helm-guard/internal/rules/builtin"
)

var rulesCmd = &cobra.Command{
	Use:   "rules",
	Short: "Manage and inspect rule set",
}

var rulesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all built-in rules",
	RunE: func(cmd *cobra.Command, args []string) error {
		rs := builtin.All()
		sort.Slice(rs, func(i, j int) bool { return rs[i].ID() < rs[j].ID() })
		for _, r := range rs {
			fmt.Printf("%s [%s] %s\n", r.ID(), r.Severity().String(), r.Title())
		}
		return nil
	},
}

func init() {
	rulesCmd.AddCommand(rulesListCmd)
}
