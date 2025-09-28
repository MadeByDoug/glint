// internal/app/linter/json_reporter.go
package reporting

import (
	"encoding/json"
	"fmt"
	"io"
)

// JsonReporter formats issues as a JSON array.
type JsonReporter struct{}

func (r *JsonReporter) Report(writer io.Writer, issues []Issue) error {
	encoder := json.NewEncoder(writer)
	if err := encoder.Encode(issues); err != nil {
		return fmt.Errorf("encode issues: %w", err)
	}
	return nil
}
