package builtin_test

import (
	"testing"

	"github.com/r3based/helm-guard/internal/model"
	"github.com/r3based/helm-guard/internal/rules/builtin"
)

func TestSingleReplicaLoadBalancerRule_Smoke(t *testing.T) {
	rule := builtin.SingleReplicaLoadBalancerRule{}

	m := model.Model{
		Workloads: []model.Workload{
			{
				Kind:      "Deployment",
				Name:      "single-public",
				Namespace: "default",
				Replicas:  1,
				PodLabels: map[string]string{"app": "single-public"},
			},
			{
				Kind:      "Deployment",
				Name:      "multi-public",
				Namespace: "default",
				Replicas:  2,
				PodLabels: map[string]string{"app": "multi-public"},
			},
		},
		Services: []model.Service{
			{
				Name:      "lb-single",
				Namespace: "default",
				Type:      "LoadBalancer",
				Selector:  map[string]string{"app": "single-public"},
			},
			{
				Name:      "lb-multi",
				Namespace: "default",
				Type:      "LoadBalancer",
				Selector:  map[string]string{"app": "multi-public"},
			},
		},
	}

	findings := rule.Check(m)

	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}

	if findings[0].Object.Name != "single-public" {
		t.Fatalf("expected finding for workload 'single-public', got %q", findings[0].Object.Name)
	}
}
