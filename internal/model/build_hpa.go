package model

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func init() {
	RegisterKind("HorizontalPodAutoscaler", buildHorizontalPodAutoscalerInto)
}

func buildHorizontalPodAutoscalerInto(m *Model, o unstructured.Unstructured) {
	m.HorizontalPodAutoscalers = append(m.HorizontalPodAutoscalers, buildHorizontalPodAutoscaler(o))
}

func buildHorizontalPodAutoscaler(o unstructured.Unstructured) HorizontalPodAutoscaler {
	ns := nsOrDefault(o.GetNamespace())
	targetKind, _, _ := unstructured.NestedString(o.Object, "spec", "scaleTargetRef", "kind")
	targetName, _, _ := unstructured.NestedString(o.Object, "spec", "scaleTargetRef", "name")
	if targetKind == "" {
		targetKind = "Deployment"
	}
	minReplicas, _, _ := unstructured.NestedInt64(o.Object, "spec", "minReplicas")
	maxReplicas, ok, _ := unstructured.NestedInt64(o.Object, "spec", "maxReplicas")
	if !ok {
		maxReplicas = 1
	}
	h := HorizontalPodAutoscaler{
		Name:        o.GetName(),
		Namespace:   ns,
		TargetKind:  targetKind,
		TargetName:  targetName,
		MinReplicas: int32(minReplicas),
		MaxReplicas: int32(maxReplicas),
	}
	if pct, ok, _ := unstructured.NestedInt64(o.Object, "spec", "targetCPUUtilizationPercentage"); ok && pct > 0 {
		p := int32(pct)
		h.TargetCPUUtilization = &p
	}
	if metrics, _, _ := unstructured.NestedSlice(o.Object, "spec", "metrics"); len(metrics) > 0 {
		for _, mAny := range metrics {
			mMap, ok := mAny.(map[string]any)
			if !ok {
				continue
			}
			t, _, _ := unstructured.NestedString(mMap, "type")
			if t != "Resource" {
				continue
			}
			resName, _, _ := unstructured.NestedString(mMap, "resource", "name")
			if resName != "cpu" {
				continue
			}
			if util, ok, _ := unstructured.NestedInt64(mMap, "resource", "target", "averageUtilization"); ok && util > 0 {
				p := int32(util)
				h.TargetCPUUtilization = &p
				break
			}
		}
	}
	return h
}
