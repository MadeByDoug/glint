// internal/app/lint/testutil/tree_builder.go
package testutil

import "github.com/MrBigCode/glint/internal/app/lint"

type B struct {
	cur  *lint.Node
	last *lint.Node // The most recently added node, for chaining.
}

func New() *B {
	root := &lint.Node{Name: "", Kind: lint.Dir, Meta: map[string]any{}}
	return &B{cur: root, last: root}
}

// Dir adds a directory under the current cursor and descends into it.
func (b *B) Dir(name string, kids ...func(*B)) *B {
	d := &lint.Node{Name: name, Kind: lint.Dir, Meta: map[string]any{}}
	d.Parent = b.cur
	b.cur.Children = append(b.cur.Children, d)
	b.last = d // Set as the last added node for chaining.

	// Descend for nested builders.
	prev := b.cur
	b.cur = d
	for _, k := range kids {
		k(b)
	}
	b.cur = prev // Ascend back to parent.
	return b
}

// File adds a file node under the current cursor.
func (b *B) File(name string) *B {
	f := &lint.Node{Name: name, Kind: lint.File, Meta: map[string]any{}}
	f.Parent = b.cur
	b.cur.Children = append(b.cur.Children, f)
	b.last = f
	return b
}

// WithMeta adds metadata to the most recently created node.
// This enables fluent chaining like `b.Dir("...").WithMeta(...)`.
func (b *B) WithMeta(key string, value any) *B {
	if b.last == nil {
		return b // Should not happen with proper usage.
	}
	if b.last.Meta == nil {
		b.last.Meta = make(map[string]any)
	}
	b.last.Meta[key] = value
	return b
}

func (b *B) Build() *lint.Tree { return &lint.Tree{Root: b.root()} }

func (b *B) root() *lint.Node {
	r := b.cur
	for r.Parent != nil {
		r = r.Parent
	}
	return r
}
