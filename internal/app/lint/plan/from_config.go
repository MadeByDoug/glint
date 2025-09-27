// internal/app/lint/plan/from_config.go
package plan

import (
	"fmt"
	"regexp"

	model "github.com/MrBigCode/glint/internal/app/config/model"
	"github.com/MrBigCode/glint/internal/app/infra/output/reporting"
	"github.com/MrBigCode/glint/internal/app/lint"
	"github.com/MrBigCode/glint/internal/app/lint/util"
)

// FromConfig translates the linter config into a list of concrete checkers.
func FromConfig(cfg model.LinterConfig) ([]lint.Checker, error) {
	var out []lint.Checker

	for i := range cfg.Rules {
		rule := &cfg.Rules[i]
		selector := buildSelector(rule)
		selectedChecks, err := buildSelectedChecks(rule, selector)
		if err != nil {
			return nil, err
		}
		out = append(out, selectedChecks...)
	}

	return out, nil
}

func buildSelector(rule *model.Rule) lint.Selector {
	return lint.Selector{
		Kind:        lint.SelectorKind(rule.Selector.Kind),
		PathRegexes: compilePatterns(rule.Selector.Path),
		Meta:        rule.Selector.Meta,
	}
}

func compilePatterns(patterns []string) []*regexp.Regexp {
	compiled := make([]*regexp.Regexp, len(patterns))
	for i, pat := range patterns {
		compiled[i] = regexp.MustCompile(util.EnsureAnchors(pat))
	}
	return compiled
}

func buildSelectedChecks(rule *model.Rule, selector lint.Selector) ([]lint.Checker, error) {
	var checks []lint.Checker
	for _, chkConfig := range rule.Apply.Checks {
		checker, err := New(chkConfig.Type, rule.Id, reporting.SeverityError, chkConfig.Params)
		if err != nil {
			return nil, fmt.Errorf("build check for rule %q: %w", rule.Id, err)
		}
		checks = append(checks, lint.WithSelector(selector, checker))
	}
	return checks, nil
}

func messageDeref(m *model.Message) string {
	if m == nil {
		return ""
	}
	return string(*m)
}

func mapSeverity(s *reporting.Severity) reporting.Severity {
	if s == nil {
		// Default to 'error' if severity is not specified in the config.
		return reporting.SeverityError
	}
	if *s == "" {
		return reporting.SeverityError
	}
	return *s
}

func toStringSlice(pp []model.Predicate) []string {
	out := make([]string, len(pp))
	for i, p := range pp {
		out[i] = string(p)
	}
	return out
}
