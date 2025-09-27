package util

import "strings"

// EnsureAnchors ensures a regex pattern is anchored at the beginning and end.
func EnsureAnchors(pat string) string {
	if !strings.HasPrefix(pat, "^") {
		pat = "^" + pat
	}
	if !strings.HasSuffix(pat, "$") {
		pat += "$"
	}
	return pat
}
