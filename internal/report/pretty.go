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
			fmt.Fprintf(w, "- %s/%s (%s) replicas=%d\n", wl.Namespace, wl.Name, wl.Kind, wl.Replicas)
			if len(wl.Containers) == 0 {
				fmt.Fprintln(w, "  - (no containers)")
				continue
			}
			for _, c := range wl.Containers {
				fmt.Fprintf(w, "  - %s image=%s cpu=%s/%s mem=%s/%s probes=R:%s L:%s\n",
					emptyIf(c.Name, "?"),
					emptyIf(c.Image, "?"),
					qStr(c.CPURequest), qStr(c.CPULimit),
					qStr(c.MemRequest), qStr(c.MemLimit),
					yn(c.HasReadiness), yn(c.HasLiveness),
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
			fmt.Fprintf(w, "- %s/%s type=%s ports=[%s]\n", s.Namespace, s.Name, s.Type, strings.Join(ports, ", "))
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
