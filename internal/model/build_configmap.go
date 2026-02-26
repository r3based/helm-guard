package model

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func init() {
	RegisterKind("ConfigMap", buildConfigMapInto)
}

func buildConfigMapInto(m *Model, o unstructured.Unstructured) {
	m.ConfigMaps = append(m.ConfigMaps, buildConfigMap(o))
}

func buildConfigMap(o unstructured.Unstructured) ConfigMap {
	ns := nsOrDefault(o.GetNamespace())
	dataMap, _, _ := unstructured.NestedStringMap(o.Object, "data")
	keys := make([]string, 0, len(dataMap))
	for k := range dataMap {
		keys = append(keys, k)
	}
	binaryDataMap, _, _ := unstructured.NestedMap(o.Object, "binaryData")
	for k := range binaryDataMap {
		keys = append(keys, k)
	}
	return ConfigMap{Name: o.GetName(), Namespace: ns, DataKeys: keys}
}
