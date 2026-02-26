package model

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func init() {
	RegisterKind("Service", buildServiceInto)
}

func buildServiceInto(m *Model, o unstructured.Unstructured) {
	m.Services = append(m.Services, buildService(o))
}

func buildService(o unstructured.Unstructured) Service {
	ns := nsOrDefault(o.GetNamespace())

	typ, ok, _ := unstructured.NestedString(o.Object, "spec", "type")
	if !ok || typ == "" {
		typ = "ClusterIP"
	}
	clusterIP, _, _ := unstructured.NestedString(o.Object, "spec", "clusterIP")

	selectorMap, _, _ := unstructured.NestedStringMap(o.Object, "spec", "selector")
	selector := map[string]string{}
	for k, v := range selectorMap {
		selector[k] = v
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
		ClusterIP: clusterIP,
		Selector:  selector,
		Ports:     ports,
	}
}
