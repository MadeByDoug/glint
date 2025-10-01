package reporting

import (
	"fmt"
	"strings"
)

// TextReporter formats issues as human-readable text.
type TextReporter struct {
	BaseReporter
}

func (r *TextReporter) Report(issues []Issue) error {
	filteredIssues := r.filterIssues(issues)

	for _, issue := range filteredIssues {
		level := formatSeverity(issue.Severity)
		caller := fmt.Sprintf("%s:%d", issue.File, issue.Line)
		message := fmt.Sprintf("%s rule=%s", issue.Message, issue.RuleID)

		// 2. Simplify the format string to remove the timestamp.
		// Pad the level to a fixed width (5) so columns align.
		line := fmt.Sprintf("%-5s %s > %s\n",
			level,
			caller,
			message,
		)

		if _, err := r.writer.Write([]byte(line)); err != nil {
			return fmt.Errorf("write report line: %w", err)
		}
	}
	return nil
}

// formatSeverity maps Severity to console-style level tokens used by zerolog ConsoleWriter.
// Examples: ERROR, WARN, INFO, DEBUG.
func formatSeverity(s Severity) string {
	switch strings.ToLower(string(s)) {
	case "error":
		return "ERROR"
	case "warning", "warn":
		return "WARN"
	case "info", "information":
		return "INFO"
	case "debug":
		return "DEBUG"
	default:
		return strings.ToUpper(string(s))
	}
}
