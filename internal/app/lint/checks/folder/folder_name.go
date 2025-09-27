// internal/app/lint/checks/folder/folder_name.go
package folder

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/MrBigCode/glint/internal/app/infra/output/reporting"
	"github.com/MrBigCode/glint/internal/app/lint"
	"github.com/MrBigCode/glint/internal/app/lint/util"
)

// FolderNameConfig holds the validation rules for a folder's name.
// These rules are applied to nodes pre-filtered by a selector.
type FolderNameConfig struct {
	Predicates     []string
	Allow          []string
	Disallow       []string
	Prefix         []string
	Suffix         []string
	ProhibitPrefix []string
	ProhibitSuffix []string
	Message        string
}

// FolderNameCheck enforces naming rules on folder names.
type FolderNameCheck struct {
	config   FolderNameConfig
	code     string
	severity reporting.Severity

	allowRes    []*regexp.Regexp
	disallowRes []*regexp.Regexp
}

// NewFolderNameCheck is the constructor for the FolderName check.
func NewFolderNameCheck(config *FolderNameConfig, code string, sev reporting.Severity) lint.Checker {
	cfg := FolderNameConfig{}
	if config != nil {
		cfg = *config
	}

	fn := &FolderNameCheck{
		config:   cfg,
		code:     code,
		severity: sev,
	}
	for _, pat := range cfg.Allow {
		fn.allowRes = append(fn.allowRes, regexp.MustCompile(util.EnsureAnchors(pat)))
	}
	for _, pat := range cfg.Disallow {
		fn.disallowRes = append(fn.disallowRes, regexp.MustCompile(util.EnsureAnchors(pat)))
	}
	return fn
}

var predicateValidators = newPredicateValidators()

func newPredicateValidators() map[string]func(string) bool {
	validators := make(map[string]func(string) bool)
	for _, pred := range util.SupportedCasingPredicates() {
		casings, ok := util.PredicateCasings(pred)
		if !ok {
			continue
		}
		validators[pred] = newPredicateValidator(casings)
	}
	return validators
}

func newPredicateValidator(casings []string) func(string) bool {
	casingsCopy := append([]string(nil), casings...)
	return func(name string) bool {
		return util.CheckCasing(name, casingsCopy)
	}
}

func normalizePredicate(pred string) string {
	if canonical, ok := util.NormalizeCasingPredicateName(pred); ok {
		return canonical
	}
	return ""
}

func (c *FolderNameCheck) ID() string { return "check.folderName" }

func (c *FolderNameCheck) Apply(_ context.Context, t *lint.Tree) []reporting.Report {
	panic("internal error: FolderNameCheck is a node-specific check and must be wrapped in a selector")
}

func (c *FolderNameCheck) ApplyToNode(ctx context.Context, n *lint.Node) []reporting.Report {
	if err := ctx.Err(); err != nil {
		return nil
	}
	if !c.isTarget(n) {
		return nil
	}

	name := n.Name
	if c.allowed(name) {
		return nil
	}

	if report := c.disallowedReport(n, name); report != nil {
		return []reporting.Report{*report}
	}

	return c.collectViolations(n, name)
}

func (c *FolderNameCheck) newReport(n *lint.Node, msg string) reporting.Report {
	message := msg
	if c.config.Message != "" {
		message = fmt.Sprintf("%s (%s)", c.config.Message, msg)
	}
	return reporting.Report{
		Severity: c.severity,
		Code:     c.code,
		Msg:      fmt.Sprintf("%s: %s", n.Path(), message),
	}
}

func (c *FolderNameCheck) checkPredicates(name string) []string {
	var failed []string
	if len(c.config.Predicates) == 0 {
		return failed
	}

	for _, raw := range c.config.Predicates {
		pred := normalizePredicate(raw)
		validator, ok := predicateValidators[pred]
		if !ok {
			continue
		}
		if !validator(name) {
			failed = append(failed, pred)
		}
	}

	return failed
}

func (c *FolderNameCheck) allowed(name string) bool {
	for _, re := range c.allowRes {
		if re.MatchString(name) {
			return true
		}
	}
	return false
}

func (c *FolderNameCheck) disallowed(name string) (matched bool, pattern string) {
	for _, re := range c.disallowRes {
		if re.MatchString(name) {
			return true, re.String()
		}
	}
	return false, ""
}

func (c *FolderNameCheck) isTarget(n *lint.Node) bool {
	return n.Kind == lint.Dir && n.Parent != nil
}

func (c *FolderNameCheck) disallowedReport(n *lint.Node, name string) *reporting.Report {
	if bad, pat := c.disallowed(name); bad {
		report := c.newReport(n, fmt.Sprintf("name '%s' matches disallowed pattern %q", name, pat))
		return &report
	}
	return nil
}

func (c *FolderNameCheck) collectViolations(n *lint.Node, name string) []reporting.Report {
	var diags []reporting.Report

	for _, pred := range c.checkPredicates(name) {
		diags = append(diags, c.newReport(n, fmt.Sprintf("name '%s' does not satisfy predicate: %s", name, pred)))
	}

	for _, msg := range c.propertyViolations(name) {
		diags = append(diags, c.newReport(n, msg))
	}

	return diags
}

func (c *FolderNameCheck) propertyViolations(name string) []string {
	var msgs []string

	msgs = append(msgs, c.checkRequiredAffixes(name)...)
	msgs = append(msgs, c.checkProhibitedAffixes(name)...)

	return msgs
}

func (c *FolderNameCheck) checkRequiredAffixes(name string) []string {
	var msgs []string

	for _, prefix := range c.config.Prefix {
		if !strings.HasPrefix(name, prefix) {
			msgs = append(msgs, fmt.Sprintf("name '%s' does not have required prefix '%s'", name, prefix))
		}
	}
	for _, suffix := range c.config.Suffix {
		if !strings.HasSuffix(name, suffix) {
			msgs = append(msgs, fmt.Sprintf("name '%s' does not have required suffix '%s'", name, suffix))
		}
	}

	return msgs
}

func (c *FolderNameCheck) checkProhibitedAffixes(name string) []string {
	var msgs []string

	for _, prefix := range c.config.ProhibitPrefix {
		if strings.HasPrefix(name, prefix) {
			msgs = append(msgs, fmt.Sprintf("name '%s' has prohibited prefix '%s'", name, prefix))
		}
	}
	for _, suffix := range c.config.ProhibitSuffix {
		if strings.HasSuffix(name, suffix) {
			msgs = append(msgs, fmt.Sprintf("name '%s' has prohibited suffix '%s'", name, suffix))
		}
	}

	return msgs
}
