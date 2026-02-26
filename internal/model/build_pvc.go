package model

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func init() {
	RegisterKind("PersistentVolumeClaim", buildPersistentVolumeClaimInto)
}

func buildPersistentVolumeClaimInto(m *Model, o unstructured.Unstructured) {
	m.PersistentVolumeClaims = append(m.PersistentVolumeClaims, buildPersistentVolumeClaim(o))
}

func buildPersistentVolumeClaim(o unstructured.Unstructured) PersistentVolumeClaim {
	ns := nsOrDefault(o.GetNamespace())
	sc, _, _ := unstructured.NestedString(o.Object, "spec", "storageClassName")
	modesAny, _, _ := unstructured.NestedSlice(o.Object, "spec", "accessModes")
	modes := make([]string, 0, len(modesAny))
	for _, m := range modesAny {
		if s, ok := m.(string); ok {
			modes = append(modes, s)
		}
	}
	capMap, _, _ := unstructured.NestedMap(o.Object, "spec", "resources", "requests")
	capacity := ""
	if req, ok := capMap["storage"]; ok {
		if q, ok := req.(string); ok {
			capacity = q
		}
	}
	return PersistentVolumeClaim{
		Name:         o.GetName(),
		Namespace:    ns,
		StorageClass: sc,
		AccessModes:  modes,
		Capacity:     capacity,
	}
}
