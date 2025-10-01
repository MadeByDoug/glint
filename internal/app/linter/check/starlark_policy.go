// internal/app/linter/check/starlark_policy.go
package check

import (
	"fmt"

	"github.com/MadeByDoug/glint/internal/app/infra/reporting"
	"go.starlark.net/starlark"
)

// StarlarkPolicyCheck executes a starlark script against a specific AST node.
type StarlarkPolicyCheck struct{}

// starlarkPolicyParams holds the validated configuration for the check.
type starlarkPolicyParams struct {
	Script string
}

// parseParams validates the 'with' map from the check config.
func (c *StarlarkPolicyCheck) parseParams(with map[string]interface{}) (starlarkPolicyParams, error) {
	params := starlarkPolicyParams{}

	scriptVal, ok := with["script"]
	if !ok {
		return params, fmt.Errorf("missing required parameter 'script'")
	}
	scriptStr, ok := scriptVal.(string)
	if !ok {
		return params, fmt.Errorf("parameter 'script' must be a string")
	}
	params.Script = scriptStr

	return params, nil
}

// ExecuteOnNode runs the starlark policy check against a single AST node.
func (c *StarlarkPolicyCheck) ExecuteOnNode(ctx PolicyExecutionContext) ([]reporting.Issue, error) {
	if ctx.TargetNode == nil {
		return nil, fmt.Errorf("internal error: starlark-policy was called with a nil node")
	}

	params, err := c.parseParams(ctx.CheckConfig.With)
	if err != nil {
		return nil, fmt.Errorf("invalid params for check '%s': %w", ctx.CheckConfig.ID, err)
	}

	thread := &starlark.Thread{Name: fmt.Sprintf("policy-%s", ctx.CheckConfig.ID)}
	predeclared := buildStarlarkPredeclared(ctx)

	// 1. Execute the script as a program to define its functions.
	globals, err := starlark.ExecFile(thread, "<policy>", []byte(params.Script), predeclared)
	if err != nil {
		// This catches syntax errors in the user's script.
		return nil, fmt.Errorf("failed to parse starlark policy '%s': %w", ctx.CheckConfig.ID, err)
	}

	// 2. Look for the required 'validate' function in the script's globals.
	validateFunc, ok := globals["validate"]
	if !ok {
		return nil, fmt.Errorf("starlark policy '%s' must define a 'validate()' function", ctx.CheckConfig.ID)
	}

	// 3. Call the 'validate' function. It takes no arguments because the context
	//    is available via the globally declared helper functions (text(), level(), etc.).
	val, err := starlark.Call(thread, validateFunc, nil, nil)
	if err != nil {
		// This catches runtime errors inside the validate function.
		return nil, fmt.Errorf("error calling validate() in starlark policy '%s': %w", ctx.CheckConfig.ID, err)
	}

	// 4. Check the script's return value to see if a validation issue was raised.
	if msg, ok := starlark.AsString(val); ok && msg != "" {
		issue := reporting.Issue{
			RuleID:   ctx.CheckConfig.ID,
			Severity: reporting.SeverityError, // Future: This could be configurable.
			Message:  msg,
		}
		return []reporting.Issue{issue}, nil
	}

	return nil, nil // No issues found
}