package profile

import "github.com/r3based/helm-guard/internal/rules"

// ToEngineOptions converts an EffectiveProfile into rules.Options that can be
// consumed by the rules.Engine. Additional disabled IDs supplied by the CLI
// are applied with highest precedence.
func ToEngineOptions(ep EffectiveProfile, cliDisabled map[string]bool) rules.Options {
	disable := map[string]bool{}
	severityOverride := map[string]rules.Severity{}

	for id, cfg := range ep.RuleConfigs {
		if !cfg.Enabled {
			disable[id] = true
		}
		severityOverride[id] = cfg.Severity
	}

	// CLI-level disables always win over profile configuration.
	for id := range cliDisabled {
		disable[id] = true
	}

	return rules.Options{
		DisableIDs:       disable,
		SeverityOverride: severityOverride,
	}
}

