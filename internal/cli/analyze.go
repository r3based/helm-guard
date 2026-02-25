package cli

import (
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/r3based/helm-guard/internal/helm"
	"github.com/r3based/helm-guard/internal/kube"
	"github.com/r3based/helm-guard/internal/model"
	"github.com/r3based/helm-guard/internal/report"
	"github.com/r3based/helm-guard/internal/rules"
	"github.com/r3based/helm-guard/internal/rules/builtin"
)

var (
	valuesFiles []string
	setValues   []string
	namespace   string
	releaseName string

	failOn  string
	disable []string
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze <chartPath>",
	Args:  cobra.ExactArgs(1),
	Short: "Render chart and analyze manifests",
	RunE: func(cmd *cobra.Command, args []string) error {
		chartPath := args[0]

		rendered, err := helm.Template(helm.TemplateParams{
			ChartPath: chartPath,
			Release:   releaseName,
			Namespace: namespace,
			Values:    valuesFiles,
			Set:       setValues,
		})
		if err != nil {
			return err
		}

		objs, err := kube.ParseManifests(rendered)
		if err != nil {
			return err
		}

		m := model.Build(objs)

		disableIDs := map[string]bool{}
		for _, x := range disable {
			for _, id := range strings.Split(x, ",") {
				id = strings.TrimSpace(id)
				if id != "" {
					disableIDs[id] = true
				}
			}
		}

		engine := rules.New(builtin.All(), rules.Options{DisableIDs: disableIDs})
		findings := engine.Run(m)

		report.Pretty(os.Stdout, len(objs), m, findings)

		failSeverity := parseFailOn(failOn)
		if failSeverity >= 0 {
			os.Exit(rules.ExitCode(findings, failSeverity))
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
