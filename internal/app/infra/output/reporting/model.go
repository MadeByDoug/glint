// internal/app/infra/output/reporting/model.go
package reporting

import (
	"fmt"
	"strings"
)

type Severity string

const (
	SevError   Severity = "error"
	SevWarning Severity = "warning"
	SevNote    Severity = "note"
)

type Report struct {
	Msg      string
	Severity Severity
	Code     string
}

func Error(code, msg string) Report   { return Report{Msg: msg, Code: code, Severity: SevError} }
func Warning(code, msg string) Report { return Report{Msg: msg, Code: code, Severity: SevWarning} }
func Note(code, msg string) Report    { return Report{Msg: msg, Code: code, Severity: SevNote} }

// --- added helpers ---

func severityRank(s Severity) int {
	switch s {
	case SevNote:
		return 0
	case SevWarning:
		return 1
	case SevError:
		return 2
	default:
		return -1
	}
}

// AtLeast returns true if s >= min (error > warning > note).
func AtLeast(s, min Severity) bool {
	return severityRank(s) >= severityRank(min)
}

// ParseThreshold parses CLI/user text to a Severity threshold.
// Accepts aliases: "warn" -> SevWarning, "info" -> SevNote.
func ParseThreshold(s string) (Severity, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "error":
		return SevError, nil
	case "warn", "warning":
		return SevWarning, nil
	case "note", "info":
		return SevNote, nil
	default:
		return "", fmt.Errorf("unknown severity %q (valid: error|warn|note)", s)
	}
}
