package builtin

import (
	"github.com/r3based/helm-guard/internal/model"
	"github.com/r3based/helm-guard/internal/rules"
)

type MissingLimitsRule struct{}

func (r MissingLimitsRule) ID() string               { return "HG2002" }
func (r MissingLimitsRule) Title() string            { return "Missing resource limits" }
func (r MissingLimitsRule) Severity() rules.Severity { return rules.Medium }
func (r MissingLimitsRule) Rationale() string {
	return "Without limits, containers may consume excessive resources and impact node stability; memory overuse can cause OOM kills."
}
func (r MissingLimitsRule) Remediation() string {
	return "Define memory limits (and CPU limits if required by your policy)."
}
func (r MissingLimitsRule) Links() []string {
	return []string{
		"https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/",
	}
}

func (r MissingLimitsRule) Check(m model.Model) []rules.Finding {
	var out []rules.Finding

	for _, w := range m.Workloads {
		for _, c := range w.Containers {
			missingCPU := c.CPULimit == nil
			missingMem := c.MemLimit == nil

			if missingCPU || missingMem {
				msg := "Missing resource limits:"
				if missingCPU {
					msg += " cpu"
				}
				if missingMem {
					msg += " memory"
				}
				out = append(out, rules.Finding{
					RuleID:   r.ID(),
					Severity: r.Severity(),
					Title:    r.Title(),
					Message:  msg,
					Object:   rules.ObjectRef{Kind: w.Kind, Namespace: w.Namespace, Name: w.Name, Container: c.Name},
				})
			}
		}
	}

	return out
}
