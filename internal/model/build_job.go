package model

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func init() {
	RegisterKind("Job", buildJobInto)
}

func buildJobInto(m *Model, o unstructured.Unstructured) {
	m.Jobs = append(m.Jobs, buildJob(o))
}

func buildJob(o unstructured.Unstructured) Job {
	ns := nsOrDefault(o.GetNamespace())
	j := Job{Name: o.GetName(), Namespace: ns}
	j.PodTemplate = buildPodTemplateRef(o.Object, "spec", "template")
	return j
}
