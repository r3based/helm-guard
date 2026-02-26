package model

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func init() {
	RegisterKind("Ingress", buildIngressInto)
}

func buildIngressInto(m *Model, o unstructured.Unstructured) {
	m.Ingresses = append(m.Ingresses, buildIngress(o))
}

func buildIngress(o unstructured.Unstructured) Ingress {
	ns := nsOrDefault(o.GetNamespace())
	ingressClassName, _, _ := unstructured.NestedString(o.Object, "spec", "ingressClassName")

	var hosts []string
	var services []string
	rulesAny, _, _ := unstructured.NestedSlice(o.Object, "spec", "rules")
	for _, rAny := range rulesAny {
		rMap, ok := rAny.(map[string]any)
		if !ok {
			continue
		}
		host, _, _ := unstructured.NestedString(rMap, "host")
		if host != "" {
			hosts = append(hosts, host)
		}
		pathsAny, _, _ := unstructured.NestedSlice(rMap, "http", "paths")
		for _, pAny := range pathsAny {
			pMap, ok := pAny.(map[string]any)
			if !ok {
				continue
			}
			svcName, _, _ := unstructured.NestedString(pMap, "backend", "service", "name")
			if svcName == "" {
				svcName, _, _ = unstructured.NestedString(pMap, "backend", "serviceName")
			}
			if svcName != "" {
				services = append(services, svcName)
			}
		}
	}

	var tls []IngressTLS
	tlsAny, _, _ := unstructured.NestedSlice(o.Object, "spec", "tls")
	for _, tAny := range tlsAny {
		tMap, ok := tAny.(map[string]any)
		if !ok {
			continue
		}
		secretName, _, _ := unstructured.NestedString(tMap, "secretName")
		hostsTLS, _, _ := unstructured.NestedStringSlice(tMap, "hosts")
		tls = append(tls, IngressTLS{Hosts: hostsTLS, SecretName: secretName})
	}

	return Ingress{
		Name:             o.GetName(),
		Namespace:        ns,
		IngressClassName: ingressClassName,
		Hosts:            hosts,
		Services:         services,
		TLS:              tls,
	}
}
