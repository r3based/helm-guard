package model

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func init() {
	RegisterKind("PodDisruptionBudget", buildPodDisruptionBudgetInto)
}

func buildPodDisruptionBudgetInto(m *Model, o unstructured.Unstructured) {
	m.PodDisruptionBudgets = append(m.PodDisruptionBudgets, buildPodDisruptionBudget(o))
}

func buildPodDisruptionBudget(o unstructured.Unstructured) PodDisruptionBudget {
	ns := nsOrDefault(o.GetNamespace())
	selectorMap, _, _ := unstructured.NestedStringMap(o.Object, "spec", "selector", "matchLabels")
	selector := map[string]string{}
	for k, v := range selectorMap {
		selector[k] = v
	}
	minAvail := intOrString(o.Object, "spec", "minAvailable")
	maxUnavail := intOrString(o.Object, "spec", "maxUnavailable")
	return PodDisruptionBudget{
		Name:           o.GetName(),
		Namespace:      ns,
		Selector:       selector,
		MinAvailable:   minAvail,
		MaxUnavailable: maxUnavail,
	}
}
