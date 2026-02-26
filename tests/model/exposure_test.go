package model_test

import (
	"testing"

	"github.com/r3based/helm-guard/internal/model"
)

func TestPublicServicesAndWorkloadsForService(t *testing.T) {
	m := model.Model{
		Workloads: []model.Workload{
			{
				Kind:      "Deployment",
				Name:      "app",
				Namespace: "default",
				Replicas:  2,
				PodLabels: map[string]string{"app": "demo"},
			},
			{
				Kind:      "Deployment",
				Name:      "other",
				Namespace: "default",
				Replicas:  1,
				PodLabels: map[string]string{"app": "other"},
			},
		},
		Services: []model.Service{
			{
				Name:      "clusterip",
				Namespace: "default",
				Type:      "ClusterIP",
				Selector:  map[string]string{"app": "demo"},
			},
			{
				Name:      "lb",
				Namespace: "default",
				Type:      "LoadBalancer",
				Selector:  map[string]string{"app": "demo"},
			},
			{
				Name:      "nodeport",
				Namespace: "default",
				Type:      "NodePort",
				Selector:  map[string]string{"app": "other"},
			},
		},
		Ingresses: []model.Ingress{
			{
				Name:      "ing",
				Namespace: "default",
				Services:  []string{"clusterip"},
			},
		},
	}

	public := model.PublicServices(m, true)
	if len(public) != 3 {
		t.Fatalf("expected 3 public services (lb, nodeport, clusterip via Ingress), got %d", len(public))
	}

	var hasLB, hasClusterIP, hasNodePort bool
	for _, s := range public {
		switch s.Name {
		case "lb":
			hasLB = true
		case "clusterip":
			hasClusterIP = true
		case "nodeport":
			hasNodePort = true
		}
	}
	if !hasLB || !hasClusterIP || !hasNodePort {
		t.Fatalf("public services missing expected entries: lb=%v clusterip=%v nodeport=%v", hasLB, hasClusterIP, hasNodePort)
	}

	var lbService model.Service
	for _, s := range m.Services {
		if s.Name == "lb" {
			lbService = s
			break
		}
	}
	wlsForLB := model.WorkloadsForService(m, lbService)
	if len(wlsForLB) != 1 || wlsForLB[0].Name != "app" {
		t.Fatalf("expected lb to target workload 'app', got %+v", wlsForLB)
	}

	var clusterIPService model.Service
	for _, s := range m.Services {
		if s.Name == "clusterip" {
			clusterIPService = s
			break
		}
	}
	wlsForClusterIP := model.WorkloadsForService(m, clusterIPService)
	if len(wlsForClusterIP) != 1 || wlsForClusterIP[0].Name != "app" {
		t.Fatalf("expected clusterip to target workload 'app', got %+v", wlsForClusterIP)
	}
}
