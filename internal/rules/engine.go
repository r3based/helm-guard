package rules

import (
	"sort"

	"github.com/r3based/helm-guard/internal/model"
)

type Options struct {
	DisableIDs       map[string]bool
	SeverityOverride map[string]Severity
}

type Engine struct {
	rules []Rule
	opts  Options
}

func New(rs []Rule, opts Options) *Engine {
	return &Engine{rules: rs, opts: opts}
}

func (e *Engine) Run(m model.Model) []Finding {
	var out []Finding

	for _, r := range e.rules {
		if e.opts.DisableIDs != nil && e.opts.DisableIDs[r.ID()] {
			continue
		}
		findings := r.Check(m)
		for i := range findings {
			// normalize meta in case rule forgot
			if findings[i].RuleID == "" {
				findings[i].RuleID = r.ID()
			}
			if findings[i].Title == "" {
				findings[i].Title = r.Title()
			}
			if findings[i].Severity == 0 && r.Severity() != 0 {
				findings[i].Severity = r.Severity()
			}
			if e.opts.SeverityOverride != nil {
				if sev, ok := e.opts.SeverityOverride[r.ID()]; ok {
					findings[i].Severity = sev
				}
			}
			if findings[i].Remediation == "" {
				findings[i].Remediation = r.Remediation()
			}
			if len(findings[i].Links) == 0 {
				findings[i].Links = r.Links()
			}
		}
		out = append(out, findings...)
	}

	sort.Slice(out, func(i, j int) bool {
		// severity desc
		if out[i].Severity != out[j].Severity {
			return out[i].Severity > out[j].Severity
		}
		// rule id asc
		if out[i].RuleID != out[j].RuleID {
			return out[i].RuleID < out[j].RuleID
		}
		// object asc
		a, b := out[i].Object, out[j].Object
		if a.Kind != b.Kind {
			return a.Kind < b.Kind
		}
		if a.Namespace != b.Namespace {
			return a.Namespace < b.Namespace
		}
		if a.Name != b.Name {
			return a.Name < b.Name
		}
		return a.Container < b.Container
	})

	return out
}

func ExitCode(findings []Finding, failOn Severity) int {
	if failOn < Cosmetic {
		return 0
	}
	for _, f := range findings {
		if f.Severity >= failOn {
			return 1
		}
	}
	return 0
}
