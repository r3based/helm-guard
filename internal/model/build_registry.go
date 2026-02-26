package model

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// BuildFunc adds a single object into the model. Used by the builder registry.
type BuildFunc func(m *Model, o unstructured.Unstructured)

var registry = make(map[string]BuildFunc)

// RegisterKind registers a builder for the given kind. Called from init() of build_*.go files.
func RegisterKind(kind string, f BuildFunc) {
	registry[kind] = f
}

// Build constructs the model from a list of unstructured objects.
func Build(objs []unstructured.Unstructured) Model {
	m := Model{
		Summary: Summary{
			CPURequests: resource.MustParse("0"),
			CPULimits:   resource.MustParse("0"),
			MemRequests: resource.MustParse("0"),
			MemLimits:   resource.MustParse("0"),
		},
	}

	for _, o := range objs {
		kind := o.GetKind()
		if f, ok := registry[kind]; ok {
			f(&m, o)
		} else {
			m.Other = append(m.Other, ObjectRef{
				APIVersion: o.GetAPIVersion(),
				Kind:       kind,
				Name:       o.GetName(),
				Namespace:  nsOrDefault(o.GetNamespace()),
			})
		}
	}

	return m
}

func nsOrDefault(ns string) string {
	if ns == "" {
		return "default"
	}
	return ns
}

func intOrString(obj map[string]any, path ...string) string {
	s, ok, _ := unstructured.NestedString(obj, path...)
	if ok && s != "" {
		return s
	}
	i, ok, _ := unstructured.NestedInt64(obj, path...)
	if ok {
		return fmt.Sprintf("%d", i)
	}
	return ""
}

func nestedQuantity(obj map[string]any, path ...string) *resource.Quantity {
	s, ok, _ := unstructured.NestedString(obj, path...)
	if !ok || s == "" {
		return nil
	}
	q, err := resource.ParseQuantity(s)
	if err != nil {
		return nil
	}
	return &q
}
