package cli

import (
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/r3based/helm-guard/internal/analyzer"
	"github.com/r3based/helm-guard/internal/report"
	"github.com/r3based/helm-guard/internal/rules"
)

var (
	valuesFiles []string
	setValues   []string
	namespace   string
	releaseName string

	failOn      string
	disable     []string
	profileName string
	profileFile string
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze <chartPath>",
	Args:  cobra.ExactArgs(1),
	Short: "Render chart and analyze manifests",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := analyzer.Config{
			ChartPath:   args[0],
			Release:     releaseName,
			Namespace:   namespace,
			Values:      valuesFiles,
			Set:         setValues,
			FailOn:      failOn,
			Disable:     disable,
			ProfileName: profileName,
			ProfileFile: profileFile,
		}

		result, err := analyzer.Run(cfg)
		if err != nil {
			return err
		}

		report.Pretty(os.Stdout, result.RenderedCount, result.Model, result.Findings)

		failSeverity := parseFailOn(failOn)
		if failSeverity < 0 && result.EffectiveProfile.FailOn != nil {
			failSeverity = *result.EffectiveProfile.FailOn
		}

		if failSeverity >= 0 {
			os.Exit(rules.ExitCode(result.Findings, failSeverity))
		}
		return nil
	},
}

func init() {
	analyzeCmd.Flags().StringArrayVarP(&valuesFiles, "values", "f", nil, "Values files")
	analyzeCmd.Flags().StringArrayVar(&setValues, "set", nil, "Set values (key=val)")
	analyzeCmd.Flags().StringVar(&namespace, "namespace", "", "Namespace")
	analyzeCmd.Flags().StringVar(&releaseName, "release", "release", "Release name")

	analyzeCmd.Flags().StringVar(&failOn, "fail-on", "off", "Fail on severity: off|cosmetic|low|medium|high|critical")
	analyzeCmd.Flags().StringArrayVar(&disable, "disable", nil, "Disable rule IDs (comma-separated)")
	analyzeCmd.Flags().StringVar(&profileName, "profile", "", "Built-in profile name: dev|prod|strict")
	analyzeCmd.Flags().StringVar(&profileFile, "profile-file", "", "Path to YAML profile file with rule configuration")
}

func parseFailOn(s string) rules.Severity {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "off", "":
		return -1
	case "cosmetic":
		return rules.Cosmetic
	case "low":
		return rules.Low
	case "medium":
		return rules.Medium
	case "high":
		return rules.High
	case "critical":
		return rules.Critical
	default:
		return -1
	}
}
