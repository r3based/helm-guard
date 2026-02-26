package model

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func init() {
	RegisterKind("Role", buildRoleInto)
	RegisterKind("ClusterRole", buildClusterRoleInto)
	RegisterKind("RoleBinding", buildRoleBindingInto)
	RegisterKind("ClusterRoleBinding", buildClusterRoleBindingInto)
}

func buildRoleInto(m *Model, o unstructured.Unstructured) {
	m.Roles = append(m.Roles, buildRole(o))
}

func buildClusterRoleInto(m *Model, o unstructured.Unstructured) {
	m.ClusterRoles = append(m.ClusterRoles, buildClusterRole(o))
}

func buildRoleBindingInto(m *Model, o unstructured.Unstructured) {
	m.RoleBindings = append(m.RoleBindings, buildRoleBinding(o))
}

func buildClusterRoleBindingInto(m *Model, o unstructured.Unstructured) {
	m.ClusterRoleBindings = append(m.ClusterRoleBindings, buildClusterRoleBinding(o))
}

func buildPolicyRules(rulesAny []interface{}) []PolicyRule {
	var out []PolicyRule
	for _, rAny := range rulesAny {
		rMap, ok := rAny.(map[string]any)
		if !ok {
			continue
		}
		groups, _, _ := unstructured.NestedStringSlice(rMap, "apiGroups")
		res, _, _ := unstructured.NestedStringSlice(rMap, "resources")
		verbs, _, _ := unstructured.NestedStringSlice(rMap, "verbs")
		out = append(out, PolicyRule{APIGroups: groups, Resources: res, Verbs: verbs})
	}
	return out
}

func buildSubjects(subjectsAny []interface{}) []Subject {
	var out []Subject
	for _, sAny := range subjectsAny {
		sMap, ok := sAny.(map[string]any)
		if !ok {
			continue
		}
		kind, _, _ := unstructured.NestedString(sMap, "kind")
		name, _, _ := unstructured.NestedString(sMap, "name")
		ns, _, _ := unstructured.NestedString(sMap, "namespace")
		out = append(out, Subject{Kind: kind, Name: name, Namespace: ns})
	}
	return out
}

func buildRole(o unstructured.Unstructured) Role {
	ns := nsOrDefault(o.GetNamespace())
	rulesAny, _, _ := unstructured.NestedSlice(o.Object, "rules")
	return Role{
		Name:      o.GetName(),
		Namespace: ns,
		Rules:     buildPolicyRules(rulesAny),
	}
}

func buildClusterRole(o unstructured.Unstructured) ClusterRole {
	rulesAny, _, _ := unstructured.NestedSlice(o.Object, "rules")
	return ClusterRole{
		Name:  o.GetName(),
		Rules: buildPolicyRules(rulesAny),
	}
}

func buildRoleBinding(o unstructured.Unstructured) RoleBinding {
	ns := nsOrDefault(o.GetNamespace())
	roleRefName, _, _ := unstructured.NestedString(o.Object, "roleRef", "name")
	subjectsAny, _, _ := unstructured.NestedSlice(o.Object, "subjects")
	return RoleBinding{
		Name:      o.GetName(),
		Namespace: ns,
		RoleRef:   roleRefName,
		Subjects:  buildSubjects(subjectsAny),
	}
}

func buildClusterRoleBinding(o unstructured.Unstructured) ClusterRoleBinding {
	roleRefName, _, _ := unstructured.NestedString(o.Object, "roleRef", "name")
	subjectsAny, _, _ := unstructured.NestedSlice(o.Object, "subjects")
	return ClusterRoleBinding{
		Name:     o.GetName(),
		RoleRef:  roleRefName,
		Subjects: buildSubjects(subjectsAny),
	}
}
