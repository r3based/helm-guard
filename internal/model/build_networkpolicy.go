package model

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func init() {
	RegisterKind("NetworkPolicy", buildNetworkPolicyInto)
}

func buildNetworkPolicyInto(m *Model, o unstructured.Unstructured) {
	m.NetworkPolicies = append(m.NetworkPolicies, buildNetworkPolicy(o))
}

func buildNetworkPolicy(o unstructured.Unstructured) NetworkPolicy {
	ns := nsOrDefault(o.GetNamespace())
	selectorMap, _, _ := unstructured.NestedStringMap(o.Object, "spec", "podSelector", "matchLabels")
	selector := map[string]string{}
	for k, v := range selectorMap {
		selector[k] = v
	}
	policyTypes, _, _ := unstructured.NestedStringSlice(o.Object, "spec", "policyTypes")

	ingressAny, _, _ := unstructured.NestedSlice(o.Object, "spec", "ingress")
	egressAny, _, _ := unstructured.NestedSlice(o.Object, "spec", "egress")

	return NetworkPolicy{
		Name:         o.GetName(),
		Namespace:    ns,
		PodSelector:  selector,
		PolicyTypes:  policyTypes,
		IngressRules: buildNPIngressRules(ingressAny),
		EgressRules:  buildNPEgressRules(egressAny),
	}
}

func buildNPIngressRules(rulesAny []interface{}) []NetworkPolicyIngressRule {
	var out []NetworkPolicyIngressRule
	for _, rAny := range rulesAny {
		rMap, ok := rAny.(map[string]any)
		if !ok {
			continue
		}
		out = append(out, NetworkPolicyIngressRule{
			Ports: buildNPPorts(rMap, "ports"),
			From:  buildNPPeers(rMap, "from"),
		})
	}
	return out
}

func buildNPEgressRules(rulesAny []interface{}) []NetworkPolicyEgressRule {
	var out []NetworkPolicyEgressRule
	for _, rAny := range rulesAny {
		rMap, ok := rAny.(map[string]any)
		if !ok {
			continue
		}
		out = append(out, NetworkPolicyEgressRule{
			Ports: buildNPPorts(rMap, "ports"),
			To:    buildNPPeers(rMap, "to"),
		})
	}
	return out
}

func buildNPPorts(rMap map[string]any, key string) []NPPort {
	portsAny, _, _ := unstructured.NestedSlice(rMap, key)
	var out []NPPort
	for _, pAny := range portsAny {
		pMap, ok := pAny.(map[string]any)
		if !ok {
			continue
		}
		proto, _, _ := unstructured.NestedString(pMap, "protocol")
		if proto == "" {
			proto = "TCP"
		}
		port := ""
		if s, ok, _ := unstructured.NestedString(pMap, "port"); ok && s != "" {
			port = s
		} else if i, ok, _ := unstructured.NestedInt64(pMap, "port"); ok {
			port = fmt.Sprintf("%d", i)
		}
		out = append(out, NPPort{Protocol: proto, Port: port})
	}
	return out
}

func buildNPPeers(rMap map[string]any, key string) []NPPeer {
	peersAny, _, _ := unstructured.NestedSlice(rMap, key)
	var out []NPPeer
	for _, peerAny := range peersAny {
		peerMap, ok := peerAny.(map[string]any)
		if !ok {
			continue
		}
		p := NPPeer{}
		if sel, _, _ := unstructured.NestedStringMap(peerMap, "podSelector", "matchLabels"); len(sel) > 0 {
			p.PodSelector = sel
		}
		if sel, _, _ := unstructured.NestedStringMap(peerMap, "namespaceSelector", "matchLabels"); len(sel) > 0 {
			p.NamespaceSelector = sel
		}
		if block, _, _ := unstructured.NestedMap(peerMap, "ipBlock"); len(block) > 0 {
			if cidr, _, _ := unstructured.NestedString(block, "cidr"); cidr != "" {
				p.IPBlockCIDR = cidr
			}
		}
		out = append(out, p)
	}
	return out
}
