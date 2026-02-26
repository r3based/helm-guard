package builtin

import (
	"github.com/r3based/helm-guard/internal/model"
	"github.com/r3based/helm-guard/internal/rules"
)

type PrivilegedOrHostRule struct{}

func (r PrivilegedOrHostRule) ID() string               { return "HG4001" }
func (r PrivilegedOrHostRule) Title() string            { return "Privileged container or host namespace sharing" }
func (r PrivilegedOrHostRule) Severity() rules.Severity { return rules.Critical }
func (r PrivilegedOrHostRule) Rationale() string {
	return "Privileged containers and hostNetwork/hostPID/hostIPC weaken isolation and can lead to host compromise or cross-pod interference."
}
func (r PrivilegedOrHostRule) Remediation() string {
	return "Avoid privileged containers; set securityContext.privileged to false. Prefer hostNetwork/hostPID/hostIPC false unless strictly required."
}
func (r PrivilegedOrHostRule) Links() []string {
	return []string{"https://kubernetes.io/docs/concepts/security/pod-security-policy/"}
}

func (r PrivilegedOrHostRule) Check(m model.Model) []rules.Finding {
	var out []rules.Finding

	for _, w := range m.Workloads {
		if w.HostNetwork || w.HostPID || w.HostIPC {
			out = append(out, rules.Finding{
				RuleID:   r.ID(),
				Severity: r.Severity(),
				Title:    r.Title(),
				Message:  "Workload uses host network or host PID/IPC",
				Object:   rules.ObjectRef{Kind: w.Kind, Namespace: w.Namespace, Name: w.Name},
			})
		}

		for _, c := range w.Containers {
			if c.Privileged {
				out = append(out, rules.Finding{
					RuleID:   r.ID(),
					Severity: r.Severity(),
					Title:    r.Title(),
					Message:  "Container runs privileged",
					Object:   rules.ObjectRef{Kind: w.Kind, Namespace: w.Namespace, Name: w.Name, Container: c.Name},
				})
			}
		}
	}

	return out
}
