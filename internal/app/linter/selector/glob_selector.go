// internal/app/linter/markdown/selector/glob_selector.go
package selector

import (
	"fmt"
	"io/fs"
	"path/filepath"
)

// GlobSelector selects files and directories based on glob patterns.
type GlobSelector struct {
	Patterns []string
}

// NewGlobSelector is the constructor for GlobSelector.
func NewGlobSelector(patterns []string) *GlobSelector {
	return &GlobSelector{Patterns: patterns}
}

// Select walks the filesystem from the context's root and returns artifacts matching the glob patterns.
func (s *GlobSelector) Select(ctx Context) ([]Artifact, error) {
	artifacts := make(map[string]Artifact) // Use a map to avoid duplicates

	// The mapper and os.Expand logic is now removed, as variables are
	// expanded globally when the configuration is first loaded.
	for _, pattern := range s.Patterns {
		err := filepath.WalkDir(ctx.Root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			relPath, err := filepath.Rel(ctx.Root, path)
			if err != nil {
				return err
			}

			// The pattern is now used directly, as it has already been expanded.
			matched, err := filepath.Match(pattern, relPath)
			if err != nil {
				return fmt.Errorf("invalid glob pattern '%s': %w", pattern, err)
			}

			if matched {
				artifactType := "file"
				if d.IsDir() {
					artifactType = "dir"
				}
				artifacts[path] = Artifact{Path: path, Type: artifactType}
			}
			return nil
		})

		if err != nil {
			return nil, err
		}
	}

	// Convert map to slice
	result := make([]Artifact, 0, len(artifacts))
	for _, art := range artifacts {
		result = append(result, art)
	}
	return result, nil
}