package builtin_test

import (
	"testing"

	"github.com/r3based/helm-guard/internal/model"
	"github.com/r3based/helm-guard/internal/rules/builtin"
)

func TestPublicNoReadinessRule_BasicScenarios(t *testing.T) {
	rule := builtin.PublicNoReadinessRule{}

	m := model.Model{
		Workloads: []model.Workload{
			{
				Kind:      "Deployment",
				Name:      "public-no-readiness",
				Namespace: "default",
				Replicas:  2,
				PodLabels: map[string]string{"app": "public-no-readiness"},
				Containers: []model.Container{
					{Name: "c1", HasReadiness: false},
				},
			},
			{
				Kind:      "Deployment",
				Name:      "public-with-readiness",
				Namespace: "default",
				Replicas:  2,
				PodLabels: map[string]string{"app": "public-with-readiness"},
				Containers: []model.Container{
					{Name: "c1", HasReadiness: true},
				},
			},
			{
				Kind:      "Deployment",
				Name:      "private-no-readiness",
				Namespace: "default",
				Replicas:  2,
				PodLabels: map[string]string{"app": "private-no-readiness"},
				Containers: []model.Container{
					{Name: "c1", HasReadiness: false},
				},
			},
		},
		Services: []model.Service{
			{
				Name:      "public-svc-no-readiness",
				Namespace: "default",
				Type:      "LoadBalancer",
				Selector:  map[string]string{"app": "public-no-readiness"},
			},
			{
				Name:      "public-svc-with-readiness",
				Namespace: "default",
				Type:      "LoadBalancer",
				Selector:  map[string]string{"app": "public-with-readiness"},
			},
		},
	}

	findings := rule.Check(m)

	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}

	f := findings[0]
	if f.Object.Name != "public-no-readiness" {
		t.Fatalf("expected finding for workload 'public-no-readiness', got %q", f.Object.Name)
	}
}
