package profile

import "github.com/r3based/helm-guard/internal/rules"

// Built-in profile identifiers.
const (
	ProfileDev    = "dev"
	ProfileProd   = "prod"
	ProfileStrict = "strict"
)

// Builtin returns a map of built-in profiles keyed by their logical name.
// These profiles only describe overrides and optional defaults; the full
// rule set and effective configuration are derived later when combined
// with the registered Rule implementations.
func Builtin() map[string]Profile {
	high := rules.High
	critical := rules.Critical

	return map[string]Profile{
		ProfileDev: {
			Name: ProfileDev,
			// dev: no opinionated failOn by default; user decides via CLI.
		},
		ProfileProd: {
			Name: ProfileProd,
			// prod: recommend failing on HIGH and above by default.
			FailOn: &high,
		},
		ProfileStrict: {
			Name: ProfileStrict,
			// strict: same default failOn as prod, but escalate a few
			// key availability rules to CRITICAL by default. The specific
			// rule IDs are stable string constants, so this is safe.
			FailOn: &high,
			Rules: []RuleOverride{
				{
					ID:       "HG3001", // Missing readinessProbe
					Severity: &critical,
				},
				{
					ID:       "HG3002", // Single replica exposed publicly
					Severity: &critical,
				},
			},
		},
	}
}

