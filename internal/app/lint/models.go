package lint

import (
	"context"
	"regexp"

	"github.com/MrBigCode/glint/internal/app/infra/output/reporting"
)


type Checker interface {
	// ID should be stable across versions for diagnostics & suppressions.
	ID() string
	// Apply analyzes the tree and returns zero or more diagnostics.
	Apply(ctx context.Context, t *Tree) []reporting.Report
}

const (
	Dir NodeKind = iota
	File
)

type Node struct {
	Name     string
	Kind     NodeKind
	Children []*Node        // only for Dir
	Meta     map[string]any // optional: tags, layer, language, etc.
	Parent   *Node          // set by builder for convenience
}

type NodeChecker interface {
	ID() string
	ApplyToNode(ctx context.Context, n *Node) []reporting.Report
}

type NodeKind int

type Selector struct {
	PathRegexes []*regexp.Regexp
	Kind        string `yaml:"kind"` // "folder"
	Meta        map[string]string
}

type SelectedCheck struct {
	selector Selector
	inner    Checker
}

type Tree struct {
	Root *Node
}

