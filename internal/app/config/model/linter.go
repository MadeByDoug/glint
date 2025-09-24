// internal/app/config/model/linter.go
package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// ----------------------
// Core types
// ----------------------

type ConfigSchemaJson struct {
	Version int    `json:"version" yaml:"version" mapstructure:"version"`
	Rules   []Rule `json:"rules"   yaml:"rules"   mapstructure:"rules"`
}

func (c *ConfigSchemaJson) UnmarshalJSON(b []byte) error {
	type plain ConfigSchemaJson
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
	*c = ConfigSchemaJson(p)
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
		c.Type = e.Type
		c.Params = e.Params
		return nil
	}

	var m map[string]json.RawMessage
	if err := strictDecode(b, &m); err != nil {
		return fmt.Errorf("apply.checks item: expected object with either type/params or single key: %w", err)
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

type Severity string

const (
	SeverityError Severity = "error"
	SeverityWarn  Severity = "warn"
	SeverityInfo  Severity = "info"
)

type Message string

// ChFolderName directly contains rule fields + optional reporting fields.
type ChFolderName struct {
	// generalized policy
	Predicates Predicates `json:"predicates,omitempty" yaml:"predicates,omitempty" mapstructure:"predicates,omitempty"`
	Allow      RegexList  `json:"allow,omitempty"      yaml:"allow,omitempty"      mapstructure:"allow,omitempty"`
	Disallow   RegexList  `json:"disallow,omitempty"   yaml:"disallow,omitempty"   mapstructure:"disallow,omitempty"`

	// optional constraints (kept for parity; still useful)
	Prefix         *string `json:"prefix,omitempty"         yaml:"prefix,omitempty"`
	Suffix         *string `json:"suffix,omitempty"         yaml:"suffix,omitempty"`
	ProhibitPrefix *string `json:"prohibitPrefix,omitempty" yaml:"prohibitPrefix,omitempty"`
	ProhibitSuffix *string `json:"prohibitSuffix,omitempty" yaml:"prohibitSuffix,omitempty"`

	// reporting
	Message  *Message  `json:"message,omitempty"  yaml:"message,omitempty"`
	Severity *Severity `json:"severity,omitempty" yaml:"severity,omitempty"`
}

func (j *ChFolderName) UnmarshalJSON(b []byte) error {
	type plain ChFolderName
	var p plain
	if err := strictDecode(b, &p); err != nil {
		return err
	}
	*j = ChFolderName(p)
	return nil
}

// ChMarkdownSchema defines the parameters for the markdown schema check.
type ChMarkdownSchema struct {
	Schema   string    `json:"schema" yaml:"schema" mapstructure:"schema"`
	Severity *Severity `json:"severity,omitempty" yaml:"severity,omitempty"`
}

func (c *ChMarkdownSchema) UnmarshalJSON(b []byte) error {
	type plain ChMarkdownSchema
	var p plain
	if err := strictDecode(b, &p); err != nil {
		return err
	}
	if strings.TrimSpace(p.Schema) == "" {
		return fmt.Errorf("field schema in markdownSchema params: required")
	}
	*c = ChMarkdownSchema(p)
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

type Predicate string

const (
	PredicateKebab Predicate = "kebab"
	PredicateLower Predicate = "lowercase"
	// future: PredicateSnake, PredicatePascal, etc.
)

func (p *Predicate) UnmarshalJSON(b []byte) error {
	var s string
	if err := strictDecode(b, &s); err != nil {
		return err
	}
	switch Predicate(s) {
	case PredicateKebab, PredicateLower, "":
		*p = Predicate(s)
		return nil
	default:
		return fmt.Errorf("unknown predicate %q (valid: kebab, lowercase)", s)
	}
}

type Predicates []Predicate

func (pp *Predicates) UnmarshalJSON(b []byte) error {
	// accept single string or list
	var s string
	if err := strictDecode(b, &s); err == nil {
		if s == "" {
			*pp = nil
			return nil
		}
		var p Predicate
		if err := (&p).UnmarshalJSON([]byte(`"` + s + `"`)); err != nil {
			return err
		}
		*pp = Predicates{p}
		return nil
	}
	var lst []Predicate
	if err := strictDecode(b, &lst); err != nil {
		return err
	}
	*pp = lst
	return nil
}

func (s *Severity) UnmarshalJSON(b []byte) error {
	var v string
	// Use strictDecode for consistency; this expects a JSON string, not an object.
	if err := strictDecode(b, &v); err != nil {
		return err
	}
	switch Severity(v) {
	case SeverityError, SeverityWarn, SeverityInfo, "":
		*s = Severity(v)
		return nil
	default:
		return fmt.Errorf("invalid severity: %q", v)
	}
}

// ----------------------
// Strict decode helper (unknown fields => error)
// ----------------------

func strictDecode(b []byte, out any) error {
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	return dec.Decode(out)
}

// ----------------------
// Regex
// ----------------------
