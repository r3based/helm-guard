package profile

import (
	"fmt"
	"io"

	"github.com/r3based/helm-guard/internal/rules"
	"sigs.k8s.io/yaml"
)

// yamlProfile mirrors the on-disk YAML representation and is converted
// into a typed Profile with rules.Severity values.
type yamlProfile struct {
	Name   string         `yaml:"name"`
	Base   string         `yaml:"base,omitempty"`
	FailOn string         `yaml:"failOn,omitempty"`
	Rules  []yamlRuleSpec `yaml:"rules,omitempty"`
}

type yamlRuleSpec struct {
	ID       string  `yaml:"id"`
	Enabled  *bool   `yaml:"enabled,omitempty"`
	Severity string  `yaml:"severity,omitempty"`
}

// LoadFromYAML reads a profile definition from a YAML reader and
// converts it into an in-memory Profile.
func LoadFromYAML(r io.Reader) (Profile, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return Profile{}, err
	}

	var yp yamlProfile
	if err := yaml.Unmarshal(data, &yp); err != nil {
		return Profile{}, err
	}

	var failOnPtr *rules.Severity
	if yp.FailOn != "" {
		s, ok := parseSeverity(yp.FailOn)
		if !ok {
			return Profile{}, fmt.Errorf("invalid failOn severity %q", yp.FailOn)
		}
		failOnPtr = &s
	}

	rulesOut := make([]RuleOverride, 0, len(yp.Rules))
	seen := map[string]struct{}{}

	for _, rr := range yp.Rules {
		if rr.ID == "" {
			return Profile{}, fmt.Errorf("rule entry with empty id")
		}
		if _, exists := seen[rr.ID]; exists {
			return Profile{}, fmt.Errorf("duplicate rule override for id %q", rr.ID)
		}
		seen[rr.ID] = struct{}{}

		var sevPtr *rules.Severity
		if rr.Severity != "" {
			s, ok := parseSeverity(rr.Severity)
			if !ok {
				return Profile{}, fmt.Errorf("invalid severity %q for rule %q", rr.Severity, rr.ID)
			}
			sevPtr = &s
		}

		rulesOut = append(rulesOut, RuleOverride{
			ID:       rr.ID,
			Enabled:  rr.Enabled,
			Severity: sevPtr,
		})
	}

	return Profile{
		Name:   yp.Name,
		Base:   yp.Base,
		FailOn: failOnPtr,
		Rules:  rulesOut,
	}, nil
}

// ResolveProfile combines a requested base profile name (typically
// coming from a CLI flag) and an optional file-based profile into a
// single logical Profile. The base precedence is:
//
//   1. fileProfile.Base, if non-empty
//   2. baseName, if non-empty
//
// If neither is provided, the resulting profile is just the fileProfile
// (or an empty Profile if fileProfile is nil).
func ResolveProfile(baseName string, fileProfile *Profile, builtin map[string]Profile) (Profile, error) {
	var chosenBase string
	if fileProfile != nil && fileProfile.Base != "" {
		chosenBase = fileProfile.Base
	} else if baseName != "" {
		chosenBase = baseName
	}

	var result Profile

	if chosenBase != "" {
		base, ok := builtin[chosenBase]
		if !ok {
			return Profile{}, fmt.Errorf("unknown base profile %q", chosenBase)
		}
		result = base
	}

	if fileProfile == nil {
		return result, nil
	}

	// Overlay simple fields.
	if fileProfile.Name != "" {
		result.Name = fileProfile.Name
	}
	// Allow explicit failOn override from file.
	if fileProfile.FailOn != nil {
		result.FailOn = fileProfile.FailOn
	}

	// Merge rule overrides by ID, with fileProfile taking precedence.
	byID := map[string]RuleOverride{}
	for _, r := range result.Rules {
		byID[r.ID] = r
	}
	for _, r := range fileProfile.Rules {
		prev, ok := byID[r.ID]
		if !ok {
			byID[r.ID] = r
			continue
		}
		merged := prev
		if r.Enabled != nil {
			merged.Enabled = r.Enabled
		}
		if r.Severity != nil {
			merged.Severity = r.Severity
		}
		byID[r.ID] = merged
	}

	result.Rules = result.Rules[:0]
	for _, r := range byID {
		result.Rules = append(result.Rules, r)
	}

	return result, nil
}

// parseSeverity parses a case-insensitive textual representation of
// severity into the rules.Severity enum.
func parseSeverity(s string) (rules.Severity, bool) {
	switch normalizeSeverity(s) {
	case "cosmetic":
		return rules.Cosmetic, true
	case "low":
		return rules.Low, true
	case "medium":
		return rules.Medium, true
	case "high":
		return rules.High, true
	case "critical":
		return rules.Critical, true
	default:
		return 0, false
	}
}

func normalizeSeverity(s string) string {
	var out []rune
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			out = append(out, r+'a'-'A')
		} else {
			out = append(out, r)
		}
	}
	return string(out)
}

