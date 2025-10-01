// internal/app/linter/selector/interfaces.go
package selector

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/MadeByDoug/glint/internal/app/infra/config"
	"github.com/MadeByDoug/glint/internal/app/infra/logging"
)

// Context provides the selection environment with access to config, env vars, etc.
type Context struct {
	Root      string
	Consts    map[string]string
	EnvVars   map[string]string
	Artifacts map[string]Artifact // ADDED: Holds all selected artifacts.
}

// Artifact represents a selected item, like a file or directory.
type Artifact struct {
	Path string
	Type string // "file" or "dir"
}

// Selector defines the interface for any artifact selection strategy.
type Selector interface {
	Select(ctx Context) ([]Artifact, error)
}

// NewContext creates a selection context from the loaded configuration and root.
func NewContext(root string, cfg *config.Config) (Context, error) {
	devLog := logging.Get()

	absRoot, err := filepath.Abs(root)
	if err != nil {
		absRoot = root
	}

	// 1. Create the initial context environment for selectors to use.
	initialCtx := Context{
		Root:      absRoot,
		Consts:    make(map[string]string),
		EnvVars:   make(map[string]string),
		Artifacts: make(map[string]Artifact),
	}

	if cfg != nil {
		for k, v := range cfg.Consts {
			initialCtx.Consts[k] = v
		}
		for _, varName := range cfg.EnvVars {
			if val, ok := os.LookupEnv(varName); ok {
				initialCtx.EnvVars[varName] = val
			}
		}
	}

	// 2. Iterate through all rules to find and execute every selector.
	if cfg == nil {
		return initialCtx, nil
	}

	for _, rule := range cfg.Rules {
		// UPDATED: Look for selectors directly on the rule, not within a nested 'checks' object.
		for idx, selCfg := range rule.Selectors {
			selInstance, err := NewFromConfig(selCfg)
			if err != nil {
				return Context{}, fmt.Errorf("build selector %d for rule %s: %w", idx, rule.ID, err)
			}

			// ... (The rest of the selector execution and artifact collection logic is unchanged) ...
			results, err := selInstance.Select(initialCtx)
			if err != nil {
				return Context{}, fmt.Errorf("run selector %d for rule %s: %w", idx, rule.ID, err)
			}

			for _, art := range results {
				canonicalPath := art.Path
				if !filepath.IsAbs(canonicalPath) {
					canonicalPath = filepath.Join(initialCtx.Root, canonicalPath)
				}
				canonicalPath = filepath.Clean(canonicalPath)

				key := fmt.Sprintf("%s::%s", canonicalPath, art.Type)
				art.Path = canonicalPath
				initialCtx.Artifacts[key] = art
			}
		}
	}

	devLog.Info().
		Int("artifact_count", len(initialCtx.Artifacts)).
		Msg("selection context built with pre-loaded artifacts")

	return initialCtx, nil
}