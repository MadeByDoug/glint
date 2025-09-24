package markdown

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
)

// Diagnostic captures a validator finding.
type Diagnostic struct {
	Code     string
	Message  string
	Section  string
	Severity string
	Line     int
	Column   int
}

const (
	severityError = "error"
	severityWarn  = "warn"
	severityInfo  = "info"
)

// RuleContext provides helpers shared by rules during validation.
type RuleContext struct {
	CaseSensitive bool
	Normalize     Normalizer
	Buf           []byte
}

// ContentRule defines a validator for a piece of section content.
type ContentRule interface {
	Type() string
	Validate(rc RuleContext, secName string, body []NodeCursor) ([]Diagnostic, []NodeCursor)
}

// NodeCursor provides lightweight access to goldmark nodes.
type NodeCursor interface {
	Kind() string
	Text() string
	Children() []NodeCursor
	Position() int
}

var (
	ruleRegistry = make(map[string]func(ContentStep) (ContentRule, error))
	reCache      = struct {
		sync.RWMutex
		cache map[string]*regexp.Regexp
	}{cache: make(map[string]*regexp.Regexp)}

	emailRe = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
)

// RegisterRule makes a content rule available to schemas.
func RegisterRule(kind string, ctor func(ContentStep) (ContentRule, error)) {
	ruleRegistry[strings.TrimSpace(kind)] = ctor
}

// BuildRule instantiates a rule by kind.
func BuildRule(step ContentStep) (ContentRule, error) {
	ctor, ok := ruleRegistry[strings.TrimSpace(step.Type)]
	if !ok {
		return nil, fmt.Errorf("unknown rule type: %s", step.Type)
	}
	return ctor(step)
}

func compileCached(pattern string) (*regexp.Regexp, error) {
	reCache.RLock()
	re, ok := reCache.cache[pattern]
	reCache.RUnlock()
	if ok {
		return re, nil
	}

	reCache.Lock()
	defer reCache.Unlock()
	if re, ok := reCache.cache[pattern]; ok {
		return re, nil
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	reCache.cache[pattern] = re
	return re, nil
}

func isEmail(s string) bool {
	return emailRe.MatchString(s)
}

func lineNumber(buf []byte, offset int) int {
	if offset <= 0 {
		return 1
	}
	if offset > len(buf) {
		offset = len(buf)
	}
	line := 1
	for i := 0; i < offset; i++ {
		if buf[i] == '\n' {
			line++
		}
	}
	return line
}
