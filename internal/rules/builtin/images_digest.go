package builtin

import (
	"strings"

	"github.com/r3based/helm-guard/internal/model"
	"github.com/r3based/helm-guard/internal/rules"
)

type ImageDigestRule struct{}

func (r ImageDigestRule) ID() string               { return "HG1002" }
func (r ImageDigestRule) Title() string            { return "Image without digest" }
func (r ImageDigestRule) Severity() rules.Severity { return rules.Low }
func (r ImageDigestRule) Rationale() string {
	return "Images referenced by tag only can change over time; using a digest pins the exact image and improves reproducibility and security."
}
func (r ImageDigestRule) Remediation() string {
	return "Use image with digest, e.g. image: myrepo/myimg@sha256:..."
}
func (r ImageDigestRule) Links() []string { return nil }

func (r ImageDigestRule) Check(m model.Model) []rules.Finding {
	var out []rules.Finding

	for _, w := range m.Workloads {
		for _, c := range w.Containers {
			img := strings.TrimSpace(c.Image)
			if img == "" {
				continue
			}
			if !strings.Contains(img, "@sha256:") {
				out = append(out, rules.Finding{
					RuleID:   r.ID(),
					Severity: r.Severity(),
					Title:    r.Title(),
					Message:  "Image is not pinned by digest",
					Object:   rules.ObjectRef{Kind: w.Kind, Namespace: w.Namespace, Name: w.Name, Container: c.Name},
				})
			}
		}
	}

	return out
}
