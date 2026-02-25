package builtin

import (
	"github.com/r3based/helm-guard/internal/model"
	"github.com/r3based/helm-guard/internal/rules"
)

type MissingLivenessProbeRule struct{}

func (r MissingLivenessProbeRule) ID() string               { return "HG2003" }
func (r MissingLivenessProbeRule) Title() string            { return "Missing livenessProbe" }
func (r MissingLivenessProbeRule) Severity() rules.Severity { return rules.Medium }
func (r MissingLivenessProbeRule) Rationale() string {
	return "Liveness probes can help Kubernetes detect and restart stuck containers, but must be configured carefully."
}
func (r MissingLivenessProbeRule) Remediation() string {
	return "Consider adding livenessProbe (or startupProbe for slow-starting apps) if appropriate."
}
func (r MissingLivenessProbeRule) Links() []string {
	return []string{
		"https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/",
	}
}

func (r MissingLivenessProbeRule) Check(m model.Model) []rules.Finding {
	var out []rules.Finding

	for _, w := range m.Workloads {
		for _, c := range w.Containers {
			if c.HasLiveness {
				continue
			}
			out = append(out, rules.Finding{
				RuleID:   r.ID(),
				Severity: r.Severity(),
				Title:    r.Title(),
				Message:  "Container has no livenessProbe",
				Object:   rules.ObjectRef{Kind: w.Kind, Namespace: w.Namespace, Name: w.Name, Container: c.Name},
			})
		}
	}

	return out
}
