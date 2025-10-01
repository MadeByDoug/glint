// internal/app/linter/check/prefix_policy.go
package check

import (
	"fmt"
	"strings"
	"github.com/MadeByDoug/glint/internal/app/infra/reporting"
)

// PrefixPolicyCheck validates that a node's text content starts with a given string.
type PrefixPolicyCheck struct{}

func (c *PrefixPolicyCheck) ExecuteOnNode(ctx PolicyExecutionContext) ([]reporting.Issue, error) {
	// 1. Parse parameters from the 'with:' block in the YAML.
	prefixVal, ok := ctx.CheckConfig.With["value"]
	if !ok {
		return nil, fmt.Errorf("prefix check requires a 'value' parameter")
	}
	prefix, ok := prefixVal.(string)
	if !ok {
		return nil, fmt.Errorf("'value' parameter for prefix check must be a string")
	}

	// 2. Get the node's text.
	nodeText := string(ctx.TargetNode.Text(ctx.TargetSource))

	// 3. Perform the check and return an issue on failure.
	if !strings.HasPrefix(nodeText, prefix) {
		issue := reporting.Issue{
			RuleID:   ctx.CheckConfig.ID,
			Severity: reporting.SeverityError,
			Message:  fmt.Sprintf("content must start with prefix '%s'", prefix),
		}
		return []reporting.Issue{issue}, nil
	}

	return nil, nil // Success
}