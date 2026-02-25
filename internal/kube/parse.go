package kube

import (
	"bytes"
	"io"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func ParseManifests(rendered []byte) ([]unstructured.Unstructured, error) {
	dec := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(rendered), 4096)

	var objs []unstructured.Unstructured

	for {
		var u unstructured.Unstructured
		err := dec.Decode(&u)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		// skip empty docs
		if u.Object == nil || len(u.Object) == 0 {
			continue
		}
		// some docs may still lack kind/apiVersion; skip them
		if u.GetKind() == "" {
			continue
		}

		objs = append(objs, u)
	}

	return objs, nil
}
