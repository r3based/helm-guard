package builtin

import "github.com/r3based/helm-guard/internal/rules"

func All() []rules.Rule {
	return []rules.Rule{
		ImageLatestRule{},
		MissingRequestsRule{},
		MissingLimitsRule{},
		MissingReadinessProbeRule{},
		MissingLivenessProbeRule{},
		SingleReplicaLoadBalancerRule{},
	}
}
