// internal/app/lint/enrich.go
package lint

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/MrBigCode/glint/internal/app/lint/parsers"
)

// EnrichTree walks the node tree and populates the Meta field for file nodes
// by calling the appropriate shallow parsers.
func EnrichTree(tree *Tree, rootDir string) error {
	if tree == nil || tree.Root == nil {
		return nil
	}
	return enrichNode(tree.Root, rootDir)
}

func enrichNode(n *Node, rootDir string) error {
	if n.Kind == File {
		if err := enrichFileNode(n, rootDir); err != nil {
			return err
		}
	}

	for _, child := range n.Children {
		if err := enrichNode(child, rootDir); err != nil {
			return err
		}
	}
	return nil
}

func enrichFileNode(n *Node, rootDir string) error {
	parser := parsers.GetParser(n.Name)
	if parser == nil {
		return nil
	}

	filePath, err := resolveFilePath(rootDir, n.Path())
	if err != nil {
		return err
	}
	content, err := os.ReadFile(filePath) // #nosec G304 -- resolveFilePath confines access to rootDir
	if err != nil {
		return &fs.PathError{Op: "read", Path: n.Path(), Err: err}
	}

	meta, err := parser(content)
	if err != nil {
		return &fs.PathError{Op: "parse", Path: n.Path(), Err: err}
	}

	if n.Meta == nil {
		n.Meta = make(map[string]any, len(meta))
	}
	for k, v := range meta {
		n.Meta[k] = v
	}

	return nil
}

func EnrichTreeSelected(tree *Tree, rootDir string, fileSelectors []Selector) error {
	if !shouldEnrichSelected(tree, fileSelectors) {
		return nil
	}

	return walkSelected(tree.Root, rootDir, fileSelectors)
}

func matchesAny(sels []Selector, rel string) bool {
	for _, s := range sels {
		if s.Kind == SelectorKindFile && s.MatchesRelPath(rel, SelectorKindFile) {
			return true
		}
	}
	return false
}

func shouldEnrichSelected(tree *Tree, selectors []Selector) bool {
	return tree != nil && tree.Root != nil && len(selectors) > 0
}

func walkSelected(n *Node, rootDir string, selectors []Selector) error {
	if n.Kind == File {
		if err := enrichSelectedFile(n, rootDir, selectors); err != nil {
			return err
		}
	}

	for _, child := range n.Children {
		if err := walkSelected(child, rootDir, selectors); err != nil {
			return err
		}
	}

	return nil
}

func enrichSelectedFile(n *Node, rootDir string, selectors []Selector) error {
	rel := strings.TrimPrefix(n.Path(), "/")
	if !matchesAny(selectors, rel) {
		return nil
	}

	parser := parsers.GetParser(n.Name)
	if parser == nil {
		return nil
	}

	filePath, err := resolveFilePath(rootDir, n.Path())
	if err != nil {
		return err
	}
	content, err := os.ReadFile(filePath) // #nosec G304 -- resolveFilePath confines access to rootDir
	if err != nil {
		return &fs.PathError{Op: "read", Path: n.Path(), Err: err}
	}

	meta, err := parser(content)
	if err != nil {
		return &fs.PathError{Op: "parse", Path: n.Path(), Err: err}
	}

	if n.Meta == nil {
		n.Meta = make(map[string]any, len(meta))
	}
	for k, v := range meta {
		n.Meta[k] = v
	}

	return nil
}

func resolveFilePath(rootDir, rel string) (string, error) {
	original := rel
	rootAbs, err := filepath.Abs(rootDir)
	if err != nil {
		return "", fmt.Errorf("resolve root %q: %w", rootDir, err)
	}
	rootAbs = filepath.Clean(rootAbs)

	rel = strings.TrimLeft(rel, "/\\")
	relPath := filepath.FromSlash(rel)
	relClean := filepath.Clean(relPath)
	if relClean == "." {
		relClean = ""
	}
	fullPath := filepath.Join(rootAbs, relClean)
	if fullPath != rootAbs {
		prefix := rootAbs + string(os.PathSeparator)
		if !strings.HasPrefix(fullPath, prefix) {
			return "", fmt.Errorf("path %q escapes root %q", original, rootDir)
		}
	}
	return fullPath, nil
}
