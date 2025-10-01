package reporting

import (
	"io"

	"github.com/rs/zerolog"
)

// BaseReporter provides a foundation for reporters, handling common tasks
// like filtering issues by severity.
type BaseReporter struct {
	writer io.Writer
	level  zerolog.Level
}

// filterIssues returns a new slice containing only the issues that meet the
// configured severity level.
func (r *BaseReporter) filterIssues(issues []Issue) []Issue {
	var filteredIssues []Issue
	for _, issue := range issues {
		if severityToLevel(issue.Severity) >= r.level {
			filteredIssues = append(filteredIssues, issue)
		}
	}
	return filteredIssues
}
