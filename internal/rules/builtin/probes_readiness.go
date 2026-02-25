package builtin

import (
	"github.com/r3based/helm-guard/internal/model"
	"github.com/r3based/helm-guard/internal/rules"
)

type MissingReadinessProbeRule struct{}

func (r MissingReadinessProbeRule) ID() string               { return "HG3001" }
func (r MissingReadinessProbeRule) Title() string            { return "Missing readinessProbe" }
func (r MissingReadinessProbeRule) Severity() rules.Severity { return rules.High }
func (r MissingReadinessProbeRule) Rationale() string {
	return "Without readinessProbe, traffic may be routed to pods that are not ready during startup or rollouts."
}
func (r MissingReadinessProbeRule) Remediation() string {
	return "Add a readinessProbe (http/tcp/exec) to each container that serves traffic."
}
func (r MissingReadinessProbeRule) Links() []string {
	return []string{
		"https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/",
	}
}

func (r MissingReadinessProbeRule) Check(m model.Model) []rules.Finding {
	var out []rules.Finding

	for _, w := range m.Workloads {
		for _, c := range w.Containers {
			if c.HasReadiness {
				continue
			}
			out = append(out, rules.Finding{
				RuleID:   r.ID(),
				Severity: r.Severity(),
				Title:    r.Title(),
				Message:  "Container has no readinessProbe",
				Object:   rules.ObjectRef{Kind: w.Kind, Namespace: w.Namespace, Name: w.Name, Container: c.Name},
			})
		}
	}

	return out
}
