package model

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

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
		switch o.GetKind() {
		case "Deployment", "StatefulSet":
			w := buildWorkload(o)
			m.Workloads = append(m.Workloads, w)
			accumulate(&m.Summary, w)
		case "Service":
			s := buildService(o)
			m.Services = append(m.Services, s)
		}
	}

	return m
}

func buildWorkload(o unstructured.Unstructured) Workload {
	ns := o.GetNamespace()
	if ns == "" {
		ns = "default"
	}

	replicas := int64(1)
	if v, ok, _ := unstructured.NestedInt64(o.Object, "spec", "replicas"); ok {
		replicas = v
	}

	containersAny, _, _ := unstructured.NestedSlice(o.Object, "spec", "template", "spec", "containers")
	containers := make([]Container, 0, len(containersAny))

	for _, cAny := range containersAny {
		cMap, ok := cAny.(map[string]any)
		if !ok {
			continue
		}

		name, _, _ := unstructured.NestedString(cMap, "name")
		image, _, _ := unstructured.NestedString(cMap, "image")

		c := Container{
			Name:  name,
			Image: image,
		}

		if _, ok, _ := unstructured.NestedMap(cMap, "readinessProbe"); ok {
			c.HasReadiness = true
		}
		if _, ok, _ := unstructured.NestedMap(cMap, "livenessProbe"); ok {
			c.HasLiveness = true
		}

		reqCPU := nestedQuantity(cMap, "resources", "requests", "cpu")
		reqMem := nestedQuantity(cMap, "resources", "requests", "memory")
		limCPU := nestedQuantity(cMap, "resources", "limits", "cpu")
		limMem := nestedQuantity(cMap, "resources", "limits", "memory")

		c.CPURequest = reqCPU
		c.MemRequest = reqMem
		c.CPULimit = limCPU
		c.MemLimit = limMem

		containers = append(containers, c)
	}

	return Workload{
		Kind:       o.GetKind(),
		Name:       o.GetName(),
		Namespace:  ns,
		Replicas:   replicas,
		Containers: containers,
	}
}

func buildService(o unstructured.Unstructured) Service {
	ns := o.GetNamespace()
	if ns == "" {
		ns = "default"
	}

	typ, ok, _ := unstructured.NestedString(o.Object, "spec", "type")
	if !ok || typ == "" {
		typ = "ClusterIP"
	}

	portsAny, _, _ := unstructured.NestedSlice(o.Object, "spec", "ports")
	ports := make([]ServicePort, 0, len(portsAny))

	for _, pAny := range portsAny {
		pMap, ok := pAny.(map[string]any)
		if !ok {
			continue
		}

		name, _, _ := unstructured.NestedString(pMap, "name")
		proto, _, _ := unstructured.NestedString(pMap, "protocol")
		if proto == "" {
			proto = "TCP"
		}

		port, _, _ := unstructured.NestedInt64(pMap, "port")

		targetPort := ""
		if tpStr, ok, _ := unstructured.NestedString(pMap, "targetPort"); ok && tpStr != "" {
			targetPort = tpStr
		} else if tpInt, ok, _ := unstructured.NestedInt64(pMap, "targetPort"); ok {
			targetPort = fmt.Sprintf("%d", tpInt)
		}

		ports = append(ports, ServicePort{
			Name:       name,
			Port:       port,
			TargetPort: targetPort,
			Protocol:   proto,
		})
	}

	return Service{
		Name:      o.GetName(),
		Namespace: ns,
		Type:      typ,
		Ports:     ports,
	}
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

func accumulate(sum *Summary, w Workload) {
	for _, c := range w.Containers {
		if c.CPURequest != nil {
			sum.CPURequests.Add(*c.CPURequest)
		}
		if c.CPULimit != nil {
			sum.CPULimits.Add(*c.CPULimit)
		}
		if c.MemRequest != nil {
			sum.MemRequests.Add(*c.MemRequest)
		}
		if c.MemLimit != nil {
			sum.MemLimits.Add(*c.MemLimit)
		}
	}
}
