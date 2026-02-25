package builtin

import (
	"github.com/r3based/helm-guard/internal/model"
	"github.com/r3based/helm-guard/internal/rules"
)

type SingleReplicaLoadBalancerRule struct{}

func (r SingleReplicaLoadBalancerRule) ID() string               { return "HG3002" }
func (r SingleReplicaLoadBalancerRule) Title() string            { return "Single replica exposed publicly" }
func (r SingleReplicaLoadBalancerRule) Severity() rules.Severity { return rules.High }
func (r SingleReplicaLoadBalancerRule) Rationale() string {
	return "A single replica exposed via LoadBalancer or Ingress is a single point of failure."
}
func (r SingleReplicaLoadBalancerRule) Remediation() string {
	return "Increase replicas to at least 2 and consider PodDisruptionBudget."
}
func (r SingleReplicaLoadBalancerRule) Links() []string { return nil }

func selectorMatches(selector, labels map[string]string) bool {
	if len(selector) == 0 {
		return false
	}
	for k, v := range selector {
		if labels[k] != v {
			return false
		}
	}
	return true
}

func (r SingleReplicaLoadBalancerRule) Check(m model.Model) []rules.Finding {
	var out []rules.Finding

	for _, w := range m.Workloads {
		if w.Replicas != 1 {
			continue
		}

		for _, s := range m.Services {
			if s.Namespace != w.Namespace {
				continue
			}

			if !selectorMatches(s.Selector, w.PodLabels) {
				continue
			}

			public := false

			// Case 1: LoadBalancer
			if s.Type == "LoadBalancer" {
				public = true
			}

			// Case 2: Ingress references this service
			if !public {
				for _, ing := range m.Ingresses {
					if ing.Namespace != s.Namespace {
						continue
					}
					for _, svcName := range ing.Services {
						if svcName == s.Name {
							public = true
							break
						}
					}
					if public {
						break
					}
				}
			}

			if public {
				out = append(out, rules.Finding{
					RuleID:   r.ID(),
					Severity: r.Severity(),
					Title:    r.Title(),
					Message:  "Workload has replicas=1 and is exposed via LoadBalancer or Ingress",
					Object: rules.ObjectRef{
						Kind:      w.Kind,
						Namespace: w.Namespace,
						Name:      w.Name,
					},
				})
			}
		}
	}

	return out
}
