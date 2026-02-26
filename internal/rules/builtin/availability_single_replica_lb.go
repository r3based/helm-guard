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
	return "A single replica exposed via LoadBalancer/NodePort/Ingress is a single point of failure; restarts and rollouts may cause downtime."
}
func (r SingleReplicaLoadBalancerRule) Remediation() string {
	return "Increase replicas to at least 2 and consider PodDisruptionBudget."
}
func (r SingleReplicaLoadBalancerRule) Links() []string { return nil }

func (r SingleReplicaLoadBalancerRule) Check(m model.Model) []rules.Finding {
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
			if w.Kind == "DaemonSet" || w.Replicas != 1 {
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
			seen[key] = true

			out = append(out, rules.Finding{
				RuleID:   r.ID(),
				Severity: r.Severity(),
				Title:    r.Title(),
				Message:  "Workload has replicas=1 and is exposed via LoadBalancer, NodePort, or Ingress",
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
