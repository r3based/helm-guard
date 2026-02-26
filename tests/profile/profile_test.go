package profile_test

import (
	"strings"
	"testing"

	"github.com/r3based/helm-guard/internal/model"
	"github.com/r3based/helm-guard/internal/profile"
	"github.com/r3based/helm-guard/internal/rules"
)

type stubRule struct {
	id       string
	severity rules.Severity
}

func (s stubRule) ID() string               { return s.id }
func (s stubRule) Title() string            { return "" }
func (s stubRule) Severity() rules.Severity { return s.severity }
func (s stubRule) Rationale() string        { return "" }
func (s stubRule) Remediation() string      { return "" }
func (s stubRule) Links() []string          { return nil }
func (s stubRule) Check(_ model.Model) []rules.Finding {
	return nil
}

func TestLoadFromYAMLAndResolveProfile(t *testing.T) {
	yamlData := `
name: my-prod-profile
base: prod
failOn: HIGH
rules:
  - id: HG3002
    enabled: true
    severity: CRITICAL
  - id: HG2003
    enabled: false
`

	pf, err := profile.LoadFromYAML(strings.NewReader(yamlData))
	if err != nil {
		t.Fatalf("LoadFromYAML error: %v", err)
	}

	if pf.Name != "my-prod-profile" {
		t.Fatalf("expected profile name %q, got %q", "my-prod-profile", pf.Name)
	}

	builtins := profile.Builtin()
	resolved, err := profile.ResolveProfile("", &pf, builtins)
	if err != nil {
		t.Fatalf("ResolveProfile error: %v", err)
	}

	if resolved.Base != "" {
		t.Fatalf("expected resolved.Base to be empty, got %q", resolved.Base)
	}
}

func TestBuildEffectiveProfileAndToEngineOptions(t *testing.T) {
	ruleset := []rules.Rule{
		stubRule{id: "HG2003", severity: rules.Medium},
		stubRule{id: "HG3002", severity: rules.High},
	}

	critical := rules.Critical
	enabled := true
	disabled := false

	prof := profile.Profile{
		Name: "test",
		Rules: []profile.RuleOverride{
			{ID: "HG3002", Enabled: &enabled, Severity: &critical},
			{ID: "HG2003", Enabled: &disabled},
		},
	}

	effective := profile.BuildEffectiveProfile(prof, ruleset)
	if len(effective.RuleConfigs) != 2 {
		t.Fatalf("expected 2 rule configs, got %d", len(effective.RuleConfigs))
	}

	cfg3002, ok := effective.RuleConfigs["HG3002"]
	if !ok {
		t.Fatalf("missing config for HG3002")
	}
	if !cfg3002.Enabled {
		t.Fatalf("expected HG3002 to be enabled")
	}
	if cfg3002.Severity != rules.Critical {
		t.Fatalf("expected HG3002 severity=CRITICAL, got %v", cfg3002.Severity)
	}

	cfg2003, ok := effective.RuleConfigs["HG2003"]
	if !ok {
		t.Fatalf("missing config for HG2003")
	}
	if cfg2003.Enabled {
		t.Fatalf("expected HG2003 to be disabled")
	}

	engineOpts := profile.ToEngineOptions(effective, map[string]bool{"HG3002": true})

	if !engineOpts.DisableIDs["HG2003"] {
		t.Fatalf("expected HG2003 to be disabled via profile")
	}
	if !engineOpts.DisableIDs["HG3002"] {
		t.Fatalf("expected HG3002 to be disabled via CLI override")
	}
	if sev, ok := engineOpts.SeverityOverride["HG3002"]; !ok || sev != rules.Critical {
		t.Fatalf("expected severity override for HG3002=CRITICAL, got %v, ok=%v", sev, ok)
	}
}
