package builtin

import "github.com/r3based/helm-guard/internal/rules"

func All() []rules.Rule {
	return []rules.Rule{
		ImageLatestRule{},
		ImageDigestRule{},
		MissingRequestsRule{},
		MissingLimitsRule{},
		MissingReadinessProbeRule{},
		MissingLivenessProbeRule{},
		SingleReplicaLoadBalancerRule{},
		PublicNoReadinessRule{},
		PublicNoPDBRule{},
		PrivilegedOrHostRule{},
	}
}
