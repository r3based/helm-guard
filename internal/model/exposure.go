package model

// PublicServices returns services that are considered publicly exposed.
// A service is treated as public if:
//   - its Type is LoadBalancer, or
//   - its Type is NodePort and treatNodePortPublic is true, or
//   - it is referenced as a backend by an Ingress in the same namespace.
//
// The result is de-duplicated and preserves a stable order based on the
// original Services slice.
func PublicServices(m Model, treatNodePortPublic bool) []Service {
	type key struct {
		ns   string
		name string
	}

	marked := make(map[key]bool)
	var out []Service

	mark := func(s Service) {
		k := key{ns: s.Namespace, name: s.Name}
		if !marked[k] {
			marked[k] = true
			out = append(out, s)
		}
	}

	// Direct exposure via Service type.
	for _, s := range m.Services {
		if s.Type == "LoadBalancer" || (treatNodePortPublic && s.Type == "NodePort") {
			mark(s)
		}
	}

	// Exposure via Ingress backends.
	for _, ing := range m.Ingresses {
		for _, svcName := range ing.Services {
			for _, s := range m.Services {
				if s.Namespace == ing.Namespace && s.Name == svcName {
					mark(s)
				}
			}
		}
	}

	return out
}

// WorkloadsForService returns workloads that are targeted by the given Service.
// Matching is strict: namespaces must be equal and the Service selector must be
// a subset of the workload's PodLabels.
func WorkloadsForService(m Model, svc Service) []Workload {
	var out []Workload
	for _, w := range m.Workloads {
		if w.Namespace != svc.Namespace {
			continue
		}
		if SelectorMatches(svc.Selector, w.PodLabels) {
			out = append(out, w)
		}
	}
	return out
}

// SelectorMatches reports whether all key/value pairs from selector are
// present in labels. An empty selector never matches.
func SelectorMatches(selector, labels map[string]string) bool {
	if len(selector) == 0 {
		return false
	}
	for k, v := range selector {
		if labels[k] != v {
			return false
		}
	}
	return true
}

