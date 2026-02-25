package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/r3based/helm-guard/internal/helm"
	"github.com/r3based/helm-guard/internal/kube"
	"github.com/r3based/helm-guard/internal/model"
)

var (
	valuesFiles []string
	setValues   []string
	namespace   string
	releaseName string
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

		fmt.Printf("Rendered objects: %d\n\n", len(objs))

		sort.Slice(m.Workloads, func(i, j int) bool {
			a, b := m.Workloads[i], m.Workloads[j]
			if a.Kind != b.Kind {
				return a.Kind < b.Kind
			}
			if a.Namespace != b.Namespace {
				return a.Namespace < b.Namespace
			}
			return a.Name < b.Name
		})

		fmt.Println("Workloads:")
		if len(m.Workloads) == 0 {
			fmt.Println("- (none)")
		}
		for _, w := range m.Workloads {
			fmt.Printf("- %s/%s (%s) replicas=%d\n", w.Namespace, w.Name, w.Kind, w.Replicas)
			if len(w.Containers) == 0 {
				fmt.Println("  - (no containers)")
				continue
			}
			for _, c := range w.Containers {
				fmt.Printf(
					"  - %s image=%s cpu=%s/%s mem=%s/%s probes=R:%s L:%s\n",
					emptyIf(c.Name, "?"),
					emptyIf(c.Image, "?"),
					qStrQ(c.CPURequest), qStrQ(c.CPULimit),
					qStrQ(c.MemRequest), qStrQ(c.MemLimit),
					yn(c.HasReadiness), yn(c.HasLiveness),
				)
			}
		}

		fmt.Println()

		sort.Slice(m.Services, func(i, j int) bool {
			a, b := m.Services[i], m.Services[j]
			if a.Namespace != b.Namespace {
				return a.Namespace < b.Namespace
			}
			return a.Name < b.Name
		})

		fmt.Println("Services:")
		if len(m.Services) == 0 {
			fmt.Println("- (none)")
		}
		for _, s := range m.Services {
			ps := make([]string, 0, len(s.Ports))
			for _, p := range s.Ports {
				tp := p.TargetPort
				if tp == "" {
					tp = "-"
				}
				prefix := ""
				if p.Name != "" {
					prefix = p.Name + ":"
				}
				ps = append(ps, fmt.Sprintf("%s%d->%s/%s", prefix, p.Port, tp, p.Protocol))
			}
			fmt.Printf("- %s/%s type=%s ports=[%s]\n", s.Namespace, s.Name, s.Type, strings.Join(ps, ", "))
		}

		fmt.Println()
		fmt.Println("Summary:")
		fmt.Printf("- total cpu requests: %s\n", m.Summary.CPURequests.String())
		fmt.Printf("- total cpu limits:   %s\n", m.Summary.CPULimits.String())
		fmt.Printf("- total mem requests: %s\n", m.Summary.MemRequests.String())
		fmt.Printf("- total mem limits:   %s\n", m.Summary.MemLimits.String())

		return nil
	},
}

func init() {
	analyzeCmd.Flags().StringArrayVarP(&valuesFiles, "values", "f", nil, "Values files")
	analyzeCmd.Flags().StringArrayVar(&setValues, "set", nil, "Set values (key=val)")
	analyzeCmd.Flags().StringVar(&namespace, "namespace", "", "Namespace")
	analyzeCmd.Flags().StringVar(&releaseName, "release", "release", "Release name")
}

func qStrQ(q *resource.Quantity) string {
	if q == nil {
		return "-"
	}
	return q.String()
}

func yn(b bool) string {
	if b {
		return "Y"
	}
	return "N"
}

func emptyIf(s, fallback string) string {
	if strings.TrimSpace(s) == "" {
		return fallback
	}
	return s
}
