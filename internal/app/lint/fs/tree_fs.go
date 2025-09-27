// internal/app/lint/fs/tree_fs.go
package fs

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"github.com/MrBigCode/glint/internal/app/lint"
)

type BuildOptions struct {
	DirsOnly     bool // when true, skip creating File nodes
	IncludeFiles bool // when true, include files (default if not DirsOnly)
}

const (
	metaKeyRelPath = "relPath"
	metaKeyAbsPath = "absPath"
)

// BuildTreeFromFS walks rootDir and constructs a lint.Tree according to options.
func BuildTreeFromFS(rootDir string, opts BuildOptions) (*lint.Tree, error) {
	rootAbs, err := filepath.Abs(rootDir)
	if err != nil {
		return nil, fmt.Errorf("resolve absolute path for %q: %w", rootDir, err)
	}

	root := &lint.Node{
		Name: "",
		Kind: lint.Dir,
		Meta: map[string]any{
			metaKeyRelPath: "",
			metaKeyAbsPath: rootAbs,
		},
	}

	index := map[string]*lint.Node{"": root}

	if err := filepath.WalkDir(rootAbs, buildWalker(rootAbs, opts, index)); err != nil {
		return nil, fmt.Errorf("walk directory %q: %w", rootAbs, err)
	}

	sortTree(root)
	return &lint.Tree{Root: root}, nil
}

func buildWalker(rootDir string, opts BuildOptions, index map[string]*lint.Node) func(string, fs.DirEntry, error) error {
	return func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("walk entry %q: %w", path, walkErr)
		}

		rel, err := relativePath(rootDir, path)
		if err != nil {
			return err
		}

		if !shouldProcess(rel, entry, opts) {
			return nil
		}

		addNode(index, rel, entry, rootDir)
		return nil
	}
}

func shouldProcess(rel string, entry fs.DirEntry, opts BuildOptions) bool {
	if rel == "" {
		return false
	}
	if shouldSkipEntry(entry, opts) {
		return false
	}
	return true
}

func relativePath(rootDir, path string) (string, error) {
	rel, err := filepath.Rel(rootDir, path)
	if err != nil {
		return "", fmt.Errorf("relative path from %q to %q: %w", rootDir, path, err)
	}
	rel = filepath.ToSlash(rel)
	if rel == "." {
		return "", nil
	}
	return rel, nil
}

func addNode(index map[string]*lint.Node, rel string, entry fs.DirEntry, rootDir string) {
	parent := ensureParent(index, parentKey(rel), rootDir)
	abs := filepath.Join(rootDir, filepath.FromSlash(rel))
	node := newTreeNode(rel, abs, entry, parent)
	parent.Children = append(parent.Children, node)
	index[rel] = node
}

func shouldSkipEntry(entry fs.DirEntry, opts BuildOptions) bool {
	if opts.DirsOnly && !entry.IsDir() {
		return true
	}
	if !opts.IncludeFiles && !entry.IsDir() {
		return true
	}
	return false
}

func parentKey(rel string) string {
	parent := filepath.ToSlash(filepath.Dir(rel))
	if parent == "." {
		return ""
	}
	return parent
}

func ensureParent(index map[string]*lint.Node, key, rootDir string) *lint.Node {
	if parent, ok := index[key]; ok {
		return parent
	}
	abs := filepath.Join(rootDir, filepath.FromSlash(key))
	parent := &lint.Node{
		Name: "",
		Kind: lint.Dir,
		Meta: map[string]any{
			metaKeyRelPath: key,
			metaKeyAbsPath: abs,
		},
	}
	index[key] = parent
	return parent
}

func newTreeNode(rel, abs string, entry fs.DirEntry, parent *lint.Node) *lint.Node {
	kind := lint.File
	if entry.IsDir() {
		kind = lint.Dir
	}

	return &lint.Node{
		Name: strings.TrimPrefix(filepath.Base(rel), "/"),
		Kind: kind,
		Meta: map[string]any{
			metaKeyRelPath: rel,
			metaKeyAbsPath: abs,
		},
		Parent: parent,
	}
}

func sortTree(n *lint.Node) {
	sort.SliceStable(n.Children, func(i, j int) bool {
		if n.Children[i].Kind != n.Children[j].Kind {
			return n.Children[i].Kind < n.Children[j].Kind
		}
		return n.Children[i].Name < n.Children[j].Name
	})

	for _, child := range n.Children {
		if child.Kind == lint.Dir {
			sortTree(child)
		}
	}
}
