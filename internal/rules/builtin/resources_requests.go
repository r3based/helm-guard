package builtin

import (
	"github.com/r3based/helm-guard/internal/model"
	"github.com/r3based/helm-guard/internal/rules"
)

type MissingRequestsRule struct{}

func (r MissingRequestsRule) ID() string               { return "HG2001" }
func (r MissingRequestsRule) Title() string            { return "Missing resource requests" }
func (r MissingRequestsRule) Severity() rules.Severity { return rules.Medium }
func (r MissingRequestsRule) Rationale() string {
	return "Without requests, scheduling and QoS become less predictable; pods are more prone to resource contention and eviction."
}
func (r MissingRequestsRule) Remediation() string {
	return "Define cpu/memory requests for each container (at least memory)."
}
func (r MissingRequestsRule) Links() []string {
	return []string{
		"https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/",
	}
}

func (r MissingRequestsRule) Check(m model.Model) []rules.Finding {
	var out []rules.Finding

	for _, w := range m.Workloads {
		for _, c := range w.Containers {
			missingCPU := c.CPURequest == nil
			missingMem := c.MemRequest == nil

			if missingCPU || missingMem {
				msg := "Missing resource requests:"
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
