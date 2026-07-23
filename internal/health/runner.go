package health

import (
	"sort"
	"time"

	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/kube/inventory"
)

const DefaultEventLookback = 30 * time.Minute

type Clock func() time.Time

type Options struct {
	Now           Clock
	EventLookback time.Duration
}

func DefaultOptions() Options {
	return Options{
		Now:           time.Now,
		EventLookback: DefaultEventLookback,
	}
}

func (options Options) withDefaults() Options {
	defaults := DefaultOptions()
	if options.Now == nil {
		options.Now = defaults.Now
	}
	if options.EventLookback == 0 {
		options.EventLookback = defaults.EventLookback
	}
	return options
}

type Rule interface {
	ID() string
	Evaluate(snapshot inventory.Snapshot, options Options) []Finding
}

type RuleFunc struct {
	RuleID string
	Run    func(snapshot inventory.Snapshot, options Options) []Finding
}

func (rule RuleFunc) ID() string {
	return rule.RuleID
}

func (rule RuleFunc) Evaluate(snapshot inventory.Snapshot, options Options) []Finding {
	if rule.Run == nil {
		return nil
	}
	return rule.Run(snapshot, options)
}

type Runner struct {
	rules []Rule
}

func NewRunner(rules ...Rule) Runner {
	copied := append([]Rule(nil), rules...)
	return Runner{rules: copied}
}

func (runner Runner) Evaluate(snapshot inventory.Snapshot, options Options) []Finding {
	options = options.withDefaults()
	var findings []Finding
	for _, rule := range runner.rules {
		if rule == nil {
			continue
		}
		findings = append(findings, rule.Evaluate(snapshot, options)...)
	}
	SortFindings(findings)
	return findings
}

func SortFindings(findings []Finding) {
	sort.SliceStable(findings, func(i, j int) bool {
		left := findings[i]
		right := findings[j]

		if severityRank(left.Severity) != severityRank(right.Severity) {
			return severityRank(left.Severity) < severityRank(right.Severity)
		}
		if left.RuleID != right.RuleID {
			return left.RuleID < right.RuleID
		}
		if left.Resource.Namespace != right.Resource.Namespace {
			return left.Resource.Namespace < right.Resource.Namespace
		}
		if left.Resource.Kind != right.Resource.Kind {
			return left.Resource.Kind < right.Resource.Kind
		}
		if left.Resource.Name != right.Resource.Name {
			return left.Resource.Name < right.Resource.Name
		}
		return left.Summary < right.Summary
	})
}

func severityRank(severity Severity) int {
	switch severity {
	case SeverityBlocker:
		return 0
	case SeverityWarning:
		return 1
	case SeverityInfo:
		return 2
	default:
		return 3
	}
}
