package model

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func init() {
	RegisterKind("CronJob", buildCronJobInto)
}

func buildCronJobInto(m *Model, o unstructured.Unstructured) {
	m.CronJobs = append(m.CronJobs, buildCronJob(o))
}

func buildCronJob(o unstructured.Unstructured) CronJob {
	ns := nsOrDefault(o.GetNamespace())
	schedule, _, _ := unstructured.NestedString(o.Object, "spec", "schedule")
	cj := CronJob{Name: o.GetName(), Namespace: ns, Schedule: schedule}
	cj.JobTemplate = buildPodTemplateRef(o.Object, "spec", "jobTemplate", "spec", "template")
	return cj
}

// buildPodTemplateRef builds PodTemplateRef from a nested path (e.g. spec.jobTemplate.spec.template or spec.template).
func buildPodTemplateRef(obj map[string]any, path ...string) *PodTemplateRef {
	template, _, _ := unstructured.NestedMap(obj, path...)
	if len(template) == 0 {
		return nil
	}
	labelMap, _, _ := unstructured.NestedStringMap(template, "metadata", "labels")
	labels := make(map[string]string)
	for k, v := range labelMap {
		labels[k] = v
	}
	containersAny, _, _ := unstructured.NestedSlice(template, "spec", "containers")
	containers := make([]Container, 0, len(containersAny))
	for _, cAny := range containersAny {
		cMap, ok := cAny.(map[string]any)
		if !ok {
			continue
		}
		containers = append(containers, buildContainer(cMap))
	}
	return &PodTemplateRef{Labels: labels, Containers: containers}
}
