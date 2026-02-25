package rules

import "github.com/r3based/helm-guard/internal/model"

type Severity int

const (
	Cosmetic Severity = iota // 0
	Low                      // 1
	Medium                   // 2
	High                     // 3
	Critical                 // 4
)

func (s Severity) String() string {
	switch s {
	case Cosmetic:
		return "COSMETIC"
	case Low:
		return "LOW"
	case Medium:
		return "MEDIUM"
	case High:
		return "HIGH"
	case Critical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

type Rule interface {
	ID() string
	Title() string
	Severity() Severity
	Rationale() string
	Remediation() string
	Links() []string
	Check(m model.Model) []Finding
}
