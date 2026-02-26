package builtin

import (
	"github.com/r3based/helm-guard/internal/model"
	"github.com/r3based/helm-guard/internal/rules"
)

type PublicNoReadinessRule struct{}

func (r PublicNoReadinessRule) ID() string               { return "HG3003" }
func (r PublicNoReadinessRule) Title() string            { return "Publicly exposed workload without readinessProbe" }
func (r PublicNoReadinessRule) Severity() rules.Severity { return rules.High }
func (r PublicNoReadinessRule) Rationale() string {
	return "Workloads that are publicly exposed without readiness probes may receive traffic before they are ready, causing failed requests and user-visible errors."
}
func (r PublicNoReadinessRule) Remediation() string {
	return "Configure a readinessProbe for containers in publicly exposed workloads so that traffic is only routed to healthy pods."
}
func (r PublicNoReadinessRule) Links() []string { return nil }

func (r PublicNoReadinessRule) Check(m model.Model) []rules.Finding {
	var out []rules.Finding

	publicSvcs := model.PublicServices(m, true)

	seen := map[struct {
		ns   string
		kind string
		name string
	}]bool{}

	for _, s := range publicSvcs {
		workloads := model.WorkloadsForService(m, s)
		for _, w := range workloads {
			if len(w.Containers) == 0 {
				continue
			}

			key := struct {
				ns   string
				kind string
				name string
			}{ns: w.Namespace, kind: w.Kind, name: w.Name}

			if seen[key] {
				continue
			}

			hasReadiness := false
			for _, c := range w.Containers {
				if c.HasReadiness {
					hasReadiness = true
					break
				}
			}

			if hasReadiness {
				continue
			}

			seen[key] = true

			out = append(out, rules.Finding{
				RuleID:   r.ID(),
				Severity: r.Severity(),
				Title:    r.Title(),
				Message:  "Workload is publicly exposed but has no readinessProbe configured",
				Object: rules.ObjectRef{
					Kind:      w.Kind,
					Namespace: w.Namespace,
					Name:      w.Name,
				},
			})
		}
	}

	return out
}

