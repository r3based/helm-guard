package model

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func init() {
	RegisterKind("Secret", buildSecretInto)
}

func buildSecretInto(m *Model, o unstructured.Unstructured) {
	m.Secrets = append(m.Secrets, buildSecret(o))
}

func buildSecret(o unstructured.Unstructured) Secret {
	ns := nsOrDefault(o.GetNamespace())
	typ, _, _ := unstructured.NestedString(o.Object, "type")
	if typ == "" {
		typ = "Opaque"
	}
	return Secret{Name: o.GetName(), Namespace: ns, Type: typ}
}
