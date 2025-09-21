// internal/app/lint/util/casing.go
package util

import (
	"regexp"
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
	"lowercase":   isLower,
	"lower":       isLower,
	"upper-case":  isUpper,
}

// CheckCasing validates if a string matches a given casing style.
func CheckCasing(s string, casings []string) bool {
	if len(casings) == 0 {
		return true // No casing rule specified, pass vacuously.
	}

	for _, casing := range casings {
		matcher := casingMatchers[strings.ToLower(casing)]
		if matcher == nil {
			return true
		}
		if matcher(s) {
			return true // Matched at least one, so we can exit early.
		}
	}
	return false // Did not match any of the specified casings.
}
