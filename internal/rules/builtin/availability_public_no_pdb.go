package builtin

import (
	"github.com/r3based/helm-guard/internal/model"
	"github.com/r3based/helm-guard/internal/rules"
)

type PublicNoPDBRule struct{}

func (r PublicNoPDBRule) ID() string               { return "HG4002" }
func (r PublicNoPDBRule) Title() string            { return "Publicly exposed workload without PodDisruptionBudget" }
func (r PublicNoPDBRule) Severity() rules.Severity { return rules.Medium }
func (r PublicNoPDBRule) Rationale() string {
	return "Public workloads should have a PDB to limit voluntary disruption during node drains and upgrades, reducing risk of full outage."
}
func (r PublicNoPDBRule) Remediation() string {
	return "Add a PodDisruptionBudget that selects the workload's pods (e.g. minAvailable: 1 or maxUnavailable: 0)."
}
func (r PublicNoPDBRule) Links() []string { return nil }

func (r PublicNoPDBRule) Check(m model.Model) []rules.Finding {
	var out []rules.Finding

	publicSvcs := model.PublicServices(m, true)
	seen := map[struct{ ns, kind, name string }]bool{}

	for _, s := range publicSvcs {
		workloads := model.WorkloadsForService(m, s)
		for _, w := range workloads {
			key := struct{ ns, kind, name string }{w.Namespace, w.Kind, w.Name}
			if seen[key] {
				continue
			}
			seen[key] = true

			hasPDB := false
			for _, pdb := range m.PodDisruptionBudgets {
				if pdb.Namespace != w.Namespace {
					continue
				}
				if model.SelectorMatches(pdb.Selector, w.PodLabels) {
					hasPDB = true
					break
				}
			}
			if hasPDB {
				continue
			}

			out = append(out, rules.Finding{
				RuleID:   r.ID(),
				Severity: r.Severity(),
				Title:    r.Title(),
				Message:  "Workload is publicly exposed but has no PodDisruptionBudget",
				Object:   rules.ObjectRef{Kind: w.Kind, Namespace: w.Namespace, Name: w.Name},
			})
		}
	}

	return out
}
