package model_test

import (
	"testing"

	"github.com/r3based/helm-guard/internal/model"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestBuild_ConfigMapSecretPDBPVCAndOther(t *testing.T) {
	objs := []unstructured.Unstructured{
		*configMapUnstructured("cm1", "default", map[string]string{"key1": "v1", "key2": "v2"}),
		*secretUnstructured("sec1", "default", "Opaque"),
		*pdbUnstructured("pdb1", "default", map[string]string{"app": "api"}, "1", ""),
		*pvcUnstructured("pvc1", "default", "standard", "10Gi"),
		otherUnstructured("v1", "LimitRange", "my-limitrange", "default"),
	}
	m := model.Build(objs)

	if len(m.ConfigMaps) != 1 {
		t.Fatalf("expected 1 ConfigMap, got %d", len(m.ConfigMaps))
	}
	if m.ConfigMaps[0].Name != "cm1" || len(m.ConfigMaps[0].DataKeys) != 2 {
		t.Fatalf("ConfigMap: name=%s keys=%d", m.ConfigMaps[0].Name, len(m.ConfigMaps[0].DataKeys))
	}

	if len(m.Secrets) != 1 {
		t.Fatalf("expected 1 Secret, got %d", len(m.Secrets))
	}
	if m.Secrets[0].Name != "sec1" || m.Secrets[0].Type != "Opaque" {
		t.Fatalf("Secret: name=%s type=%s", m.Secrets[0].Name, m.Secrets[0].Type)
	}

	if len(m.PodDisruptionBudgets) != 1 {
		t.Fatalf("expected 1 PDB, got %d", len(m.PodDisruptionBudgets))
	}
	if m.PodDisruptionBudgets[0].Name != "pdb1" || m.PodDisruptionBudgets[0].MinAvailable != "1" {
		t.Fatalf("PDB: name=%s minAvailable=%s", m.PodDisruptionBudgets[0].Name, m.PodDisruptionBudgets[0].MinAvailable)
	}

	if len(m.PersistentVolumeClaims) != 1 {
		t.Fatalf("expected 1 PVC, got %d", len(m.PersistentVolumeClaims))
	}
	if m.PersistentVolumeClaims[0].Name != "pvc1" || m.PersistentVolumeClaims[0].Capacity != "10Gi" {
		t.Fatalf("PVC: name=%s capacity=%s", m.PersistentVolumeClaims[0].Name, m.PersistentVolumeClaims[0].Capacity)
	}

	if len(m.Other) != 1 {
		t.Fatalf("expected 1 Other, got %d", len(m.Other))
	}
	if m.Other[0].Kind != "LimitRange" || m.Other[0].Name != "my-limitrange" {
		t.Fatalf("Other: kind=%s name=%s", m.Other[0].Kind, m.Other[0].Name)
	}
}

func configMapUnstructured(name, ns string, data map[string]string) *unstructured.Unstructured {
	u := &unstructured.Unstructured{}
	u.SetNamespace(ns)
	u.SetName(name)
	u.SetAPIVersion("v1")
	u.SetKind("ConfigMap")
	_ = unstructured.SetNestedStringMap(u.Object, data, "data")
	return u
}

func secretUnstructured(name, ns, typ string) *unstructured.Unstructured {
	u := &unstructured.Unstructured{}
	u.SetNamespace(ns)
	u.SetName(name)
	u.SetAPIVersion("v1")
	u.SetKind("Secret")
	_ = unstructured.SetNestedField(u.Object, typ, "type")
	return u
}

func pdbUnstructured(name, ns string, selector map[string]string, minAvail, maxUnavail string) *unstructured.Unstructured {
	u := &unstructured.Unstructured{}
	u.SetNamespace(ns)
	u.SetName(name)
	u.SetAPIVersion("policy/v1")
	u.SetKind("PodDisruptionBudget")
	_ = unstructured.SetNestedStringMap(u.Object, selector, "spec", "selector", "matchLabels")
	if minAvail != "" {
		_ = unstructured.SetNestedField(u.Object, minAvail, "spec", "minAvailable")
	}
	if maxUnavail != "" {
		_ = unstructured.SetNestedField(u.Object, maxUnavail, "spec", "maxUnavailable")
	}
	return u
}

func pvcUnstructured(name, ns, storageClass, capacity string) *unstructured.Unstructured {
	u := &unstructured.Unstructured{}
	u.SetNamespace(ns)
	u.SetName(name)
	u.SetAPIVersion("v1")
	u.SetKind("PersistentVolumeClaim")
	_ = unstructured.SetNestedField(u.Object, storageClass, "spec", "storageClassName")
	_ = unstructured.SetNestedMap(u.Object, map[string]interface{}{"storage": capacity}, "spec", "resources", "requests")
	return u
}

func otherUnstructured(apiVersion, kind, name, ns string) unstructured.Unstructured {
	u := unstructured.Unstructured{}
	u.SetAPIVersion(apiVersion)
	u.SetKind(kind)
	u.SetName(name)
	u.SetNamespace(ns)
	return u
}
