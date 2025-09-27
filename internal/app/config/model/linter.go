// internal/app/config/model/linter.go
package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/MrBigCode/glint/internal/app/infra/output/reporting"
	util "github.com/MrBigCode/glint/internal/app/lint/util"
)

// ----------------------
// Core types
// ----------------------

type LinterConfig struct {
	Version int    `json:"version" yaml:"version" mapstructure:"version"`
	Rules   []Rule `json:"rules"   yaml:"rules"   mapstructure:"rules"`
}

func (c *LinterConfig) UnmarshalJSON(b []byte) error {
	type plain LinterConfig
	var p plain
	if err := strictDecode(b, &p); err != nil {
		return err
	}
	if p.Version < 1 {
		return fmt.Errorf("field version: must be >= 1")
	}
	if len(p.Rules) == 0 {
		return fmt.Errorf("field rules: must contain at least 1 rule")
	}
	*c = LinterConfig(p)
	return nil
}

type Rule struct {
	Id       string   `json:"id"       yaml:"id"       mapstructure:"id"`
	Selector Selector `json:"selector" yaml:"selector" mapstructure:"selector"`
	Apply    Apply    `json:"apply"    yaml:"apply"    mapstructure:"apply"`
}

func (r *Rule) UnmarshalJSON(b []byte) error {
	type plain Rule
	var p plain
	if err := strictDecode(b, &p); err != nil {
		return err
	}
	if p.Id == "" {
		return fmt.Errorf("field id in Rule: required")
	}
	*r = Rule(p)
	return nil
}

// ----------------------
// Selector (folders + glob path)
// ----------------------

type SelectorKind string

const (
	SelectorKindFolder SelectorKind = "folder"
	SelectorKindFile   SelectorKind = "file"
)

type Selector struct {
	Kind SelectorKind      `json:"kind" yaml:"kind" mapstructure:"kind"` // must be "folder"
	Path RegexList         `json:"path" yaml:"path" mapstructure:"path"` // REQUIRED: regex
	Meta map[string]string `json:"meta,omitempty" yaml:"meta,omitempty" mapstructure:"meta,omitempty"`
}

func (s *Selector) UnmarshalJSON(b []byte) error {
	type plain Selector
	var p plain
	if err := strictDecode(b, &p); err != nil {
		return err
	}

	// This switch is more scalable than the old 'if' statement.
	switch p.Kind {
	case SelectorKindFolder, SelectorKindFile:
		// Valid kind, proceed.
	case "":
		return fmt.Errorf("selector.kind: field is required")
	default:
		return fmt.Errorf("selector.kind: unsupported value %q (supported: 'folder', 'file')", p.Kind)
	}

	if len(p.Path) == 0 {
		return fmt.Errorf("selector.path: required")
	}
	*s = Selector(p)
	return nil
}

// ----------------------
// Apply: only folderName check supported
// ----------------------

type Apply struct {
	Checks []ApplyCheck `json:"checks" yaml:"checks" mapstructure:"checks"`
}

// ApplyCheck is now a dynamic struct that can support any check type.
type ApplyCheck struct {
	Type   string
	Params json.RawMessage
}

func (c *ApplyCheck) UnmarshalJSON(b []byte) error {
	type explicit struct {
		Type   string          `json:"type"`
		Params json.RawMessage `json:"params"`
	}

	var e explicit
	if err := strictDecode(b, &e); err == nil && strings.TrimSpace(e.Type) != "" {
		return fmt.Errorf("apply.checks item: type/params form is no longer supported; use compact syntax like {\"folderName\": {...}}")
	}

	var m map[string]json.RawMessage
	if err := strictDecode(b, &m); err != nil {
		return fmt.Errorf("apply.checks item: expected object with single check entry: %w", err)
	}
	if len(m) != 1 {
		return fmt.Errorf("apply.checks item: expected exactly one check entry, found %d", len(m))
	}
	for k, v := range m {
		if strings.TrimSpace(k) == "" {
			return fmt.Errorf("apply.checks item: empty check name is not allowed")
		}
		c.Type = k
		c.Params = v
	}
	return nil
}

// ----------------------
// FolderName check (single canonical shape)
// ----------------------

type Message string

// FolderNameCheckParams directly contains rule fields + optional reporting fields.
type FolderNameCheckParams struct {
	// generalized policy
	Predicates Predicates `json:"predicates,omitempty" yaml:"predicates,omitempty" mapstructure:"predicates,omitempty"`
	Allow      RegexList  `json:"allow,omitempty"      yaml:"allow,omitempty"      mapstructure:"allow,omitempty"`
	Disallow   RegexList  `json:"disallow,omitempty"   yaml:"disallow,omitempty"   mapstructure:"disallow,omitempty"`

	// optional constraints (kept for parity; still useful)
	Prefix         StringList `json:"prefix,omitempty"         yaml:"prefix,omitempty"`
	Suffix         StringList `json:"suffix,omitempty"         yaml:"suffix,omitempty"`
	ProhibitPrefix StringList `json:"prohibitPrefix,omitempty" yaml:"prohibitPrefix,omitempty"`
	ProhibitSuffix StringList `json:"prohibitSuffix,omitempty" yaml:"prohibitSuffix,omitempty"`

	// reporting
	Message  *Message            `json:"message,omitempty"  yaml:"message,omitempty"`
	Severity *reporting.Severity `json:"severity,omitempty" yaml:"severity,omitempty"`
}

func (j *FolderNameCheckParams) UnmarshalJSON(b []byte) error {
	type plain FolderNameCheckParams
	var p plain
	if err := strictDecode(b, &p); err != nil {
		return err
	}
	*j = FolderNameCheckParams(p)
	return nil
}

// ----------------------
// Utility leaf types
// ----------------------

type RegexList []string

func (rl *RegexList) UnmarshalJSON(b []byte) error {
	var v []string
	if err := strictDecode(b, &v); err != nil {
		return err
	}
	// Validate each regex at load time
	for i, s := range v {
		if s == "" {
			return fmt.Errorf("regex at index %d: empty pattern", i)
		}
		if _, err := regexp.Compile(s); err != nil {
			return fmt.Errorf("regex at index %d invalid: %v", i, err)
		}
	}
	*rl = v
	return nil
}

type StringList []string

func (sl *StringList) UnmarshalJSON(b []byte) error {
	var list []string
	if err := strictDecode(b, &list); err != nil {
		return err
	}
	for i, s := range list {
		if strings.TrimSpace(s) == "" {
			return fmt.Errorf("string list entry at index %d: empty strings are not allowed", i)
		}
	}
	*sl = list
	return nil
}

type Predicate string

const (
	PredicateCamel  Predicate = "camel"
	PredicateKebab  Predicate = "kebab"
	PredicateLower  Predicate = "lower"
	PredicatePascal Predicate = "pascal"
	PredicateSnake  Predicate = "snake"
	PredicateUpper  Predicate = "upper"
)

func (p *Predicate) UnmarshalJSON(b []byte) error {
	var s string
	if err := strictDecode(b, &s); err != nil {
		return err
	}
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return fmt.Errorf("predicate: empty strings are not allowed")
	}
	if trimmed != s {
		return fmt.Errorf("predicate: leading or trailing whitespace is not allowed")
	}
	if canonical, ok := util.NormalizeCasingPredicateName(trimmed); ok {
		*p = Predicate(canonical)
		return nil
	}
	return fmt.Errorf(
		"unknown predicate %q (valid: %s)",
		trimmed,
		strings.Join(util.SupportedCasingPredicates(), ", "),
	)
}

type Predicates []Predicate

func (pp *Predicates) UnmarshalJSON(b []byte) error {
	var lst []string
	if err := strictDecode(b, &lst); err != nil {
		return err
	}
	if len(lst) == 0 {
		*pp = nil
		return nil
	}
	out := make([]Predicate, len(lst))
	for i, raw := range lst {
		var predicate Predicate
		if err := (&predicate).UnmarshalJSON([]byte(`"` + raw + `"`)); err != nil {
			return fmt.Errorf("predicates[%d]: %w", i, err)
		}
		out[i] = predicate
	}
	*pp = out
	return nil
}

// ----------------------
// Strict decode helper (unknown fields => error)
// ----------------------

func strictDecode(b []byte, out any) error {
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	if err := dec.Decode(out); err != nil {
		return fmt.Errorf("decode JSON payload: %w", err)
	}
	return nil
}

// ----------------------
// Regex
// ----------------------
