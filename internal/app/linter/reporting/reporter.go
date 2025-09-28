// internal/app/linter/reporter.go
package reporting

import (
	"fmt"
	"io"
)

// Severity defines the level of a lint issue.
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
	SeverityInfo    Severity = "info"
)

// Issue represents a single finding by the linter.
type Issue struct {
	File     string   `json:"file"`
	Line     int      `json:"line"`
	Column   int      `json:"column"` // Optional, but good practice
	RuleID   string   `json:"ruleId"`
	Severity Severity `json:"severity"`
	Message  string   `json:"message"`
}

// Reporter defines the interface for any output formatter.
// It takes the destination writer and the list of issues to report.
type Reporter interface {
	Report(writer io.Writer, issues []Issue) error
}

// NewReporter is a factory that returns the requested reporter.
func NewReporter(format string) (Reporter, error) {
	switch format {
	case "text":
		return &TextReporter{}, nil
	case "json":
		return &JsonReporter{}, nil
	default:
		return nil, fmt.Errorf("unknown report format: %q", format)
	}
}