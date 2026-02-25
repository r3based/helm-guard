package kube

import (
	"bytes"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func ParseManifests(data []byte) ([]unstructured.Unstructured, error) {
	decoder := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(data), 4096)

	var objects []unstructured.Unstructured

	for {
		var obj unstructured.Unstructured
		err := decoder.Decode(&obj)
		if err != nil {
			break
		}

		if obj.Object == nil {
			continue
		}

		objects = append(objects, obj)
	}

	return objects, nil
}
