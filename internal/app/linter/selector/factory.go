package selector

import (
	"fmt"
	"strings"

	"github.com/MadeByDoug/glint/internal/app/infra/config"
)

// NewFromConfig creates a Selector implementation based on configuration.
func NewFromConfig(sel config.SelectorConfig) (Selector, error) {
	typeName := strings.ToLower(sel.Type)
	switch typeName {
	case "glob":
		patterns, err := sel.SelectorStrings()
		if err != nil {
			return nil, err
		}
		return NewGlobSelector(patterns), nil
	case "script":
		scriptPaths, err := sel.SelectorStrings()
		if err != nil {
			return nil, err
		}
		return NewScriptSelector(scriptPaths), nil
	default:
		return nil, fmt.Errorf("unsupported selector type %q", sel.Type)
	}
}
