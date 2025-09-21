// internal/app/lint/checks/selector.go
package lint

import (
	"context"
	"strings"

	"github.com/MrBigCode/glint/internal/app/infra/output/reporting"
)

const folderKind = "folder"

// Selector defines the criteria for targeting nodes in the linting tree.

// Matches checks if a given node satisfies all conditions of the selector.
func (s *Selector) Matches(n *Node) bool {
	if !s.matchesNodeKind(n.Kind) {
		return false
	}

	return s.matchPath(n.Path())
}

// SelectedCheck prunes the tree based on the selector, then runs the inner check.

func (s *SelectedCheck) ID() string { return "check.selected(" + s.inner.ID() + ")" }

func (s *SelectedCheck) Apply(ctx context.Context, t *Tree) []reporting.Report {
	if nodeChecker, ok := s.inner.(NodeChecker); ok {
		return s.applyNodeChecker(ctx, t, nodeChecker)
	}

	return s.inner.Apply(ctx, t)
}

// WithSelector composes a checker with a selector.
func WithSelector(selector Selector, inner Checker) Checker {
	return &SelectedCheck{selector: selector, inner: inner}
}

func (s *Selector) MatchesRelPath(rel string, kind string) bool {
	if !s.matchesKindString(kind) {
		return false
	}
	return s.matchPath(rel)
}

func (s *Selector) matchesNodeKind(kind NodeKind) bool {
	switch s.Kind {
	case folderKind:
		return kind == Dir
	default:
		return false
	}
}

func (s *Selector) matchesKindString(kind string) bool {
	switch s.Kind {
	case folderKind:
		return kind == folderKind
	default:
		return false
	}
}

func (s *Selector) matchPath(path string) bool {
	if len(s.PathRegexes) == 0 {
		return false
	}

	rel := strings.TrimPrefix(path, "/")
	for _, re := range s.PathRegexes {
		if re.MatchString(rel) {
			return true
		}
	}
	return false
}

func (s *SelectedCheck) applyNodeChecker(ctx context.Context, t *Tree, checker NodeChecker) []reporting.Report {
	if t == nil || t.Root == nil {
		return nil
	}

	var out []reporting.Report
	s.walkSelectedNodes(ctx, t.Root, checker, &out)
	return out
}

func (s *SelectedCheck) walkSelectedNodes(ctx context.Context, node *Node, checker NodeChecker, out *[]reporting.Report) {
	if s.selector.Matches(node) {
		*out = append(*out, checker.ApplyToNode(ctx, node)...)
	}
	for _, child := range node.Children {
		s.walkSelectedNodes(ctx, child, checker, out)
	}
}
