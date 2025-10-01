package reporting

import (
	"encoding/json"
	"fmt"
)

// JsonReporter formats issues as a JSON array.
type JsonReporter struct {
	BaseReporter
}

func (r *JsonReporter) Report(issues []Issue) error {
	filteredIssues := r.filterIssues(issues)

	if len(filteredIssues) == 0 {
		return nil
	}

	encoder := json.NewEncoder(r.writer)
	if err := encoder.Encode(filteredIssues); err != nil {
		return fmt.Errorf("encode issues: %w", err)
	}
	return nil
}
