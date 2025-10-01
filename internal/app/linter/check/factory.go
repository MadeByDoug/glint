// internal/app/linter/check/factory.go
package check

import (
	"fmt"
	"strings"
)

func New(uses string) (interface{}, error) {
	checkType := strings.ToLower(uses)
	switch checkType {
	case "markdown-schema":
		return &MarkdownSchemaCheck{}, nil
	case "starlark-policy":
		return &StarlarkPolicyCheck{}, nil
	case "prefix":
		return &PrefixPolicyCheck{}, nil
	default:
		return nil, fmt.Errorf("unsupported check type: %q", uses)
	}
}