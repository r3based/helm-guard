package rules

type ObjectRef struct {
	Kind      string
	Namespace string
	Name      string
	Container string
}

type Finding struct {
	RuleID      string
	Severity    Severity
	Title       string
	Message     string
	Object      ObjectRef
	Remediation string
	Links       []string
}
