// internal/app/lint/util/casing.go
package util

import (
	"regexp"
	"sort"
	"strings"
)

var (
	isSnake  = regexp.MustCompile(`^[a-z0-9]+(_[a-z0-9]+)*$`).MatchString
	isKebab  = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`).MatchString
	isLower  = regexp.MustCompile(`^[a-z0-9]+$`).MatchString
	isUpper  = regexp.MustCompile(`^[A-Z0-9]+$`).MatchString
	isCamel  = regexp.MustCompile(`^[a-z]+[a-zA-Z0-9]*$`).MatchString
	isPascal = regexp.MustCompile(`^[A-Z][a-zA-Z0-9]*$`).MatchString
)

var casingMatchers = map[string]func(string) bool{
	"snake-case":  isSnake,
	"kebab-case":  isKebab,
	"camel-case":  isCamel,
	"pascal-case": isPascal,
	"lower-case":  isLower,
	"upper-case":  isUpper,
}

var predicateCasings = map[string][]string{
	"camel":  {"camel-case"},
	"kebab":  {"kebab-case"},
	"lower":  {"lower-case"},
	"pascal": {"pascal-case"},
	"snake":  {"snake-case"},
	"upper":  {"upper-case"},
}

// NormalizeCasingPredicateName converts a user-provided predicate name into the
// canonical identifier understood by glint. It returns false when the value is
// not recognized.
func NormalizeCasingPredicateName(s string) (string, bool) {
	if s == "" {
		return "", false
	}
	if _, ok := predicateCasings[s]; ok {
		return s, true
	}
	return "", false
}

// SupportedCasingPredicates returns the list of canonical predicate names that
// glint can evaluate.
func SupportedCasingPredicates() []string {
	keys := make([]string, 0, len(predicateCasings))
	for k := range predicateCasings {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// PredicateCasings returns the list of casing keywords that should be supplied
// to CheckCasing for the given canonical predicate name.
func PredicateCasings(predicate string) ([]string, bool) {
	casings, ok := predicateCasings[predicate]
	if !ok {
		return nil, false
	}
	return append([]string(nil), casings...), true
}

// CheckCasing validates if a string matches a given casing style.
func CheckCasing(s string, casings []string) bool {
	if len(casings) == 0 {
		return true // No casing rule specified, pass vacuously.
	}

	for _, casing := range casings {
		matcher := casingMatchers[strings.ToLower(casing)]
		if matcher == nil {
			return false
		}
		if matcher(s) {
			return true // Matched at least one, so we can exit early.
		}
	}
	return false // Did not match any of the specified casings.
}
