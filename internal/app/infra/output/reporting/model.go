// internal/app/infra/output/reporting/model.go
package reporting

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/MrBigCode/glint/internal/app/infra/output/level"
)

type Severity level.Level

const (
	SeverityOff   Severity = Severity(level.Off)
	SeverityError Severity = Severity(level.Error)
	SeverityWarn  Severity = Severity(level.Warn)
	SeverityInfo  Severity = Severity(level.Info)
)

type Report struct {
	Msg      string
	Severity Severity
	Code     string
}

func Error(code, msg string) Report { return Report{Msg: msg, Code: code, Severity: SeverityError} }
func Warn(code, msg string) Report  { return Report{Msg: msg, Code: code, Severity: SeverityWarn} }
func Info(code, msg string) Report  { return Report{Msg: msg, Code: code, Severity: SeverityInfo} }

// --- added helpers ---

// AtLeast returns true if s >= threshold (error > warn > info > off).
func AtLeast(s, threshold Severity) bool {
	return level.Rank(level.Level(s)) >= level.Rank(level.Level(threshold))
}

// ParseThreshold parses CLI/user text to a Severity threshold.
func ParseThreshold(s string) (Severity, error) {
	label := strings.ToLower(strings.TrimSpace(s))
	switch label {
	case "", "warn", "warning", "off", "none":
		return SeverityWarn, nil
	case "info":
		return SeverityInfo, nil
	case "error":
		return SeverityWarn, nil
	case "debug":
		return SeverityInfo, nil
	default:
		return "", fmt.Errorf("unknown diagnostic level %q (valid: warn|info)", s)
	}
}

// UnmarshalJSON parses a severity string and validates it against canonical severities.
func (s *Severity) UnmarshalJSON(b []byte) error {
	var raw string
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	return s.Set(raw)
}

// Set normalizes the provided severity label.
func (s *Severity) Set(in string) error {
	if in == "" {
		*s = ""
		return nil
	}
	lvl, err := level.Parse(in)
	if err != nil {
		return err
	}
	*s = Severity(lvl)
	return nil
}
