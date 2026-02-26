package profile

import "github.com/r3based/helm-guard/internal/rules"

// RuleOverride describes per-rule customisation within a profile.
// Fields are pointers so that "unset" can be distinguished from
// explicit false/zero values when merging profiles.
type RuleOverride struct {
	ID       string          `yaml:"id" json:"id"`
	Enabled  *bool           `yaml:"enabled,omitempty" json:"enabled,omitempty"`
	Severity *rules.Severity `yaml:"severity,omitempty" json:"severity,omitempty"`
}

// Profile describes a logical configuration of rules.
// It can optionally be based on another profile (builtin or user-defined)
// and may provide a default failOn threshold.
type Profile struct {
	Name   string          `yaml:"name" json:"name"`
	Base   string          `yaml:"base,omitempty" json:"base,omitempty"`
	FailOn *rules.Severity `yaml:"failOn,omitempty" json:"failOn,omitempty"`
	Rules  []RuleOverride  `yaml:"rules,omitempty" json:"rules,omitempty"`
}

// ResolvedRuleConfig represents the effective configuration for a single rule
// after applying profiles, built-ins, and any overrides.
type ResolvedRuleConfig struct {
	ID       string         `json:"id"`
	Enabled  bool           `json:"enabled"`
	Severity rules.Severity `json:"severity"`
}

// EffectiveProfile is the fully materialised view of a profile after
// resolving inheritance and overrides.
type EffectiveProfile struct {
	Name        string                         `json:"name"`
	FailOn      *rules.Severity                `json:"failOn,omitempty"`
	RuleConfigs map[string]ResolvedRuleConfig  `json:"rules"`
}

// BuildEffectiveProfile constructs an EffectiveProfile from a logical
// Profile definition and the concrete set of registered rules.
func BuildEffectiveProfile(p Profile, rs []rules.Rule) EffectiveProfile {
	cfgs := make(map[string]ResolvedRuleConfig, len(rs))

	// Start with defaults from the rule implementations.
	for _, r := range rs {
		cfgs[r.ID()] = ResolvedRuleConfig{
			ID:       r.ID(),
			Enabled:  true,
			Severity: r.Severity(),
		}
	}

	// Apply profile overrides.
	for _, o := range p.Rules {
		c, ok := cfgs[o.ID]
		if !ok {
			// Unknown rule id in profile: ignore silently for now to
			// allow forward/backward compatibility with plugins.
			continue
		}
		if o.Enabled != nil {
			c.Enabled = *o.Enabled
		}
		if o.Severity != nil {
			c.Severity = *o.Severity
		}
		cfgs[o.ID] = c
	}

	return EffectiveProfile{
		Name:        p.Name,
		FailOn:      p.FailOn,
		RuleConfigs: cfgs,
	}
}

