package builtin

import (
	"strings"

	"github.com/r3based/helm-guard/internal/model"
	"github.com/r3based/helm-guard/internal/rules"
)

type ImageLatestRule struct{}

func (r ImageLatestRule) ID() string               { return "HG1001" }
func (r ImageLatestRule) Title() string            { return "Image tag is 'latest' or missing" }
func (r ImageLatestRule) Severity() rules.Severity { return rules.Low }
func (r ImageLatestRule) Rationale() string {
	return "Using ':latest' (or no tag) reduces reproducibility and makes rollbacks/debugging harder."
}
func (r ImageLatestRule) Remediation() string {
	return "Pin image tags (e.g., vX.Y.Z) and/or use image digests."
}
func (r ImageLatestRule) Links() []string {
	return []string{
		"https://kubernetes.io/docs/concepts/containers/images/",
	}
}

func (r ImageLatestRule) Check(m model.Model) []rules.Finding {
	var out []rules.Finding

	for _, w := range m.Workloads {
		for _, c := range w.Containers {
			img := strings.TrimSpace(c.Image)
			if img == "" {
				continue
			}
			// if no ":" after last "/" => no tag (could still be digest, but we treat as missing tag)
			lastSlash := strings.LastIndex(img, "/")
			lastColon := strings.LastIndex(img, ":")

			hasTag := lastColon > lastSlash
			if hasTag && strings.HasSuffix(img, ":latest") {
				out = append(out, rules.Finding{
					RuleID:   r.ID(),
					Severity: r.Severity(),
					Title:    r.Title(),
					Message:  "Image uses ':latest'",
					Object:   rules.ObjectRef{Kind: w.Kind, Namespace: w.Namespace, Name: w.Name, Container: c.Name},
				})
				continue
			}
			if !hasTag && !strings.Contains(img, "@sha256:") {
				out = append(out, rules.Finding{
					RuleID:   r.ID(),
					Severity: r.Severity(),
					Title:    r.Title(),
					Message:  "Image tag is missing",
					Object:   rules.ObjectRef{Kind: w.Kind, Namespace: w.Namespace, Name: w.Name, Container: c.Name},
				})
			}
		}
	}

	return out
}
