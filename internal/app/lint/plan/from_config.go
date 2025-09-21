// internal/app/lint/plan/from_config.go
package plan

import (
	"regexp"
	"strings"

	model "github.com/MrBigCode/glint/internal/app/config/model"
	"github.com/MrBigCode/glint/internal/app/infra/output/reporting"
	"github.com/MrBigCode/glint/internal/app/lint"
)

// FromConfig translates the linter config into a list of concrete checkers.
func FromConfig(cfg model.ConfigSchemaJson) []lint.Checker {
	var out []lint.Checker

	for _, rule := range cfg.Rules {
		selector := buildSelector(rule)
		out = append(out, buildSelectedChecks(rule, selector)...)
	}

	return out
}

func buildSelector(rule model.Rule) lint.Selector {
	return lint.Selector{
		Kind:        string(rule.Selector.Kind),
		PathRegexes: compilePatterns(rule.Selector.Path),
		Meta:        rule.Selector.Meta,
	}
}

func compilePatterns(patterns []string) []*regexp.Regexp {
	compiled := make([]*regexp.Regexp, len(patterns))
	for i, pat := range patterns {
		compiled[i] = regexp.MustCompile(ensureAnchors(pat))
	}
	return compiled
}

func ensureAnchors(pat string) string {
	anchored := pat
	if !strings.HasPrefix(anchored, "^") {
		anchored = "^" + anchored
	}
	if !strings.HasSuffix(anchored, "$") {
		anchored += "$"
	}
	return anchored
}

func buildSelectedChecks(rule model.Rule, selector lint.Selector) []lint.Checker {
	var checks []lint.Checker
	for _, chkConfig := range rule.Apply.Checks {
		checker, err := New(chkConfig.Type, rule.Id, reporting.SevError, chkConfig.Params)
		if err != nil {
			// Skip invalid checks but continue processing the rest.
			continue
		}
		checks = append(checks, lint.WithSelector(selector, checker))
	}
	return checks
}

func strDeref(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func messageDeref(m *model.Message) string {
	if m == nil {
		return ""
	}
	return string(*m)
}

func mapSeverity(s *model.Severity) reporting.Severity {
	if s == nil {
		// Default to 'error' if severity is not specified in the config.
		return reporting.SevError
	}
	switch *s {
	case model.SeverityError:
		return reporting.SevError
	case model.SeverityWarn:
		return reporting.SevWarning
	case model.SeverityInfo:
		return reporting.SevNote
	default:
		// Also default to 'error' for any invalid severity values.
		return reporting.SevError
	}
}

func toStringSlice(pp []model.Predicate) []string {
	out := make([]string, len(pp))
	for i, p := range pp {
		out[i] = string(p)
	}
	return out
}
