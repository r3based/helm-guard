package report

import (
	"fmt"
	"io"
	"strings"

	"github.com/r3based/helm-guard/internal/model"
	"github.com/r3based/helm-guard/internal/rules"
)

func Pretty(w io.Writer, renderedCount int, m model.Model, findings []rules.Finding) {
	fmt.Fprintf(w, "Rendered objects: %d\n\n", renderedCount)

	fmt.Fprintln(w, "Workloads:")
	if len(m.Workloads) == 0 {
		fmt.Fprintln(w, "- (none)")
	} else {
		for _, wl := range m.Workloads {
			repl := fmt.Sprintf("%d", wl.Replicas)
			if wl.Kind == "DaemonSet" {
				repl = "daemon"
			}
			hostExtra := ""
			if wl.HostNetwork || wl.HostPID || wl.HostIPC {
				var h []string
				if wl.HostNetwork {
					h = append(h, "hostNet")
				}
				if wl.HostPID {
					h = append(h, "hostPID")
				}
				if wl.HostIPC {
					h = append(h, "hostIPC")
				}
				hostExtra = " " + strings.Join(h, ",")
			}
			sa := emptyIf(wl.ServiceAccountName, "-")
			fmt.Fprintf(w, "- %s/%s (%s) replicas=%s sa=%s%s\n", wl.Namespace, wl.Name, wl.Kind, repl, sa, hostExtra)
			if len(wl.InitContainers) > 0 {
				for _, c := range wl.InitContainers {
					fmt.Fprintf(w, "  init: %s image=%s cpu=%s/%s mem=%s/%s\n",
						emptyIf(c.Name, "?"), emptyIf(c.Image, "?"),
						qStr(c.CPURequest), qStr(c.CPULimit), qStr(c.MemRequest), qStr(c.MemLimit))
				}
			}
			if len(wl.Containers) == 0 {
				fmt.Fprintln(w, "  - (no containers)")
				continue
			}
			for _, c := range wl.Containers {
				sec := ""
				if c.Privileged || c.ReadOnlyRootFilesystem {
					var s []string
					if c.Privileged {
						s = append(s, "priv")
					}
					if c.ReadOnlyRootFilesystem {
						s = append(s, "roRoot")
					}
					sec = " " + strings.Join(s, ",")
				}
				fmt.Fprintf(w, "  - %s image=%s cpu=%s/%s mem=%s/%s probes=R:%s L:%s%s\n",
					emptyIf(c.Name, "?"),
					emptyIf(c.Image, "?"),
					qStr(c.CPURequest), qStr(c.CPULimit),
					qStr(c.MemRequest), qStr(c.MemLimit),
					yn(c.HasReadiness), yn(c.HasLiveness),
					sec,
				)
			}
		}
	}

	fmt.Fprintln(w)
	fmt.Fprintln(w, "Services:")
	if len(m.Services) == 0 {
		fmt.Fprintln(w, "- (none)")
	} else {
		for _, s := range m.Services {
			ports := make([]string, 0, len(s.Ports))
			for _, p := range s.Ports {
				tp := p.TargetPort
				if tp == "" {
					tp = "-"
				}
				prefix := ""
				if p.Name != "" {
					prefix = p.Name + ":"
				}
				ports = append(ports, fmt.Sprintf("%s%d->%s/%s", prefix, p.Port, tp, p.Protocol))
			}
			clusterIP := s.ClusterIP
			if clusterIP == "" {
				clusterIP = "-"
			}
			fmt.Fprintf(w, "- %s/%s type=%s clusterIP=%s ports=[%s]\n", s.Namespace, s.Name, s.Type, clusterIP, strings.Join(ports, ", "))
		}
	}

	fmt.Fprintln(w)
	fmt.Fprintln(w, "Ingresses:")
	if len(m.Ingresses) == 0 {
		fmt.Fprintln(w, "- (none)")
	} else {
		for _, ing := range m.Ingresses {
			hosts := strings.Join(ing.Hosts, ", ")
			if hosts == "" {
				hosts = "-"
			}
			svcs := strings.Join(ing.Services, ", ")
			if svcs == "" {
				svcs = "-"
			}
			class := ing.IngressClassName
			if class == "" {
				class = "-"
			}
			tlsInfo := ""
			if len(ing.TLS) > 0 {
				tlsInfo = fmt.Sprintf(" tls=%d", len(ing.TLS))
			}
			fmt.Fprintf(w, "- %s/%s class=%s hosts=[%s] services=[%s]%s\n", emptyIf(ing.Namespace, "default"), ing.Name, class, hosts, svcs, tlsInfo)
		}
	}

	fmt.Fprintln(w)
	fmt.Fprintln(w, "ConfigMaps:")
	if len(m.ConfigMaps) == 0 {
		fmt.Fprintln(w, "- (none)")
	} else {
		for _, cm := range m.ConfigMaps {
			fmt.Fprintf(w, "- %s/%s keys=%d\n", cm.Namespace, cm.Name, len(cm.DataKeys))
		}
	}

	fmt.Fprintln(w)
	fmt.Fprintln(w, "Secrets:")
	if len(m.Secrets) == 0 {
		fmt.Fprintln(w, "- (none)")
	} else {
		for _, sec := range m.Secrets {
			fmt.Fprintf(w, "- %s/%s type=%s\n", sec.Namespace, sec.Name, emptyIf(sec.Type, "Opaque"))
		}
	}

	fmt.Fprintln(w)
	fmt.Fprintln(w, "PodDisruptionBudgets:")
	if len(m.PodDisruptionBudgets) == 0 {
		fmt.Fprintln(w, "- (none)")
	} else {
		for _, pdb := range m.PodDisruptionBudgets {
			minMax := fmt.Sprintf("minAvailable=%s maxUnavailable=%s", pdb.MinAvailable, pdb.MaxUnavailable)
			fmt.Fprintf(w, "- %s/%s %s\n", pdb.Namespace, pdb.Name, minMax)
		}
	}

	fmt.Fprintln(w)
	fmt.Fprintln(w, "PersistentVolumeClaims:")
	if len(m.PersistentVolumeClaims) == 0 {
		fmt.Fprintln(w, "- (none)")
	} else {
		for _, pvc := range m.PersistentVolumeClaims {
			fmt.Fprintf(w, "- %s/%s storageClass=%s capacity=%s\n", pvc.Namespace, pvc.Name, emptyIf(pvc.StorageClass, "-"), pvc.Capacity)
		}
	}

	fmt.Fprintln(w)
	fmt.Fprintln(w, "HorizontalPodAutoscalers:")
	if len(m.HorizontalPodAutoscalers) == 0 {
		fmt.Fprintln(w, "- (none)")
	} else {
		for _, h := range m.HorizontalPodAutoscalers {
			cpu := "-"
			if h.TargetCPUUtilization != nil {
				cpu = fmt.Sprintf("%d%%", *h.TargetCPUUtilization)
			}
			fmt.Fprintf(w, "- %s/%s target=%s/%s min=%d max=%d cpu=%s\n", h.Namespace, h.Name, h.TargetKind, h.TargetName, h.MinReplicas, h.MaxReplicas, cpu)
		}
	}

	if len(m.Roles)+len(m.ClusterRoles)+len(m.RoleBindings)+len(m.ClusterRoleBindings) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "RBAC:")
		for _, r := range m.Roles {
			fmt.Fprintf(w, "- Role %s/%s rules=%d\n", r.Namespace, r.Name, len(r.Rules))
		}
		for _, r := range m.ClusterRoles {
			fmt.Fprintf(w, "- ClusterRole %s rules=%d\n", r.Name, len(r.Rules))
		}
		for _, rb := range m.RoleBindings {
			fmt.Fprintf(w, "- RoleBinding %s/%s -> Role %s subjects=%d\n", rb.Namespace, rb.Name, rb.RoleRef, len(rb.Subjects))
		}
		for _, crb := range m.ClusterRoleBindings {
			fmt.Fprintf(w, "- ClusterRoleBinding %s -> ClusterRole %s subjects=%d\n", crb.Name, crb.RoleRef, len(crb.Subjects))
		}
	}

	if len(m.CronJobs)+len(m.Jobs) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "CronJobs/Jobs:")
		for _, cj := range m.CronJobs {
			fmt.Fprintf(w, "- CronJob %s/%s schedule=%s\n", cj.Namespace, cj.Name, cj.Schedule)
		}
		for _, j := range m.Jobs {
			fmt.Fprintf(w, "- Job %s/%s\n", j.Namespace, j.Name)
		}
	}

	if len(m.NetworkPolicies) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "NetworkPolicies:")
		for _, np := range m.NetworkPolicies {
			types := strings.Join(np.PolicyTypes, ",")
			if types == "" {
				types = "-"
			}
			fmt.Fprintf(w, "- %s/%s policyTypes=[%s]\n", np.Namespace, np.Name, types)
		}
	}

	if len(m.Other) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintf(w, "Other objects: %d\n", len(m.Other))
		kinds := make(map[string]int)
		for _, o := range m.Other {
			kinds[o.Kind]++
		}
		for k, n := range kinds {
			fmt.Fprintf(w, "  - %s: %d\n", k, n)
		}
	}

	fmt.Fprintln(w)
	fmt.Fprintln(w, "Summary:")
	fmt.Fprintf(w, "- total cpu requests: %s\n", m.Summary.CPURequests.String())
	fmt.Fprintf(w, "- total cpu limits:   %s\n", m.Summary.CPULimits.String())
	fmt.Fprintf(w, "- total mem requests: %s\n", m.Summary.MemRequests.String())
	fmt.Fprintf(w, "- total mem limits:   %s\n", m.Summary.MemLimits.String())

	fmt.Fprintln(w)
	fmt.Fprintln(w, "Warnings:")
	if len(findings) == 0 {
		fmt.Fprintln(w, "- (none)")
	} else {
		for _, f := range findings {
			ref := fmt.Sprintf("%s/%s", emptyIf(f.Object.Namespace, "default"), f.Object.Name)
			if f.Object.Container != "" {
				ref = ref + " (container " + f.Object.Container + ")"
			}
			fmt.Fprintf(w, "[%s][%s] %s %s: %s\n", f.RuleID, f.Severity.String(), f.Object.Kind, ref, f.Message)
		}
	}
}

func yn(b bool) string {
	if b {
		return "Y"
	}
	return "N"
}

func emptyIf(s, fallback string) string {
	if strings.TrimSpace(s) == "" {
		return fallback
	}
	return s
}

func qStr(q interface{ String() string }) string {
	if q == nil {
		return "-"
	}
	return q.String()
}
