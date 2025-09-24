package markdown

import (
	"bytes"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// Doc bundles the parsed AST and original buffer for convenience.
type Doc struct {
	Root ast.Node
	MD   goldmark.Markdown
	Buf  []byte
}

// ParseMarkdown converts raw Markdown bytes into a goldmark AST.
func ParseMarkdown(input []byte) (*Doc, error) {
	md := goldmark.New()
	root := md.Parser().Parse(text.NewReader(input))
	return &Doc{Root: root, MD: md, Buf: input}, nil
}

// HeadingText extracts the literal text from a heading node.
func HeadingText(n ast.Node, buf []byte) string {
	var b bytes.Buffer
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if lit, ok := c.(*ast.Text); ok {
			b.Write(lit.Segment.Value(buf))
		}
	}
	return b.String()
}

// SectionBlock groups a heading with the nodes up to the next peer heading.
type SectionBlock struct {
	Heading *ast.Heading
	Body    []ast.Node
}

// CollectSections walks the AST and returns ordered section blocks.
func CollectSections(doc *Doc) []SectionBlock {
	if doc == nil || doc.Root == nil {
		return nil
	}

	var headings []*ast.Heading
	ast.Walk(doc.Root, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if h, ok := n.(*ast.Heading); ok {
			headings = append(headings, h)
		}
		return ast.WalkContinue, nil
	})

	blocks := make([]SectionBlock, 0, len(headings))
	for _, h := range headings {
		block := SectionBlock{Heading: h}
		for n := h.NextSibling(); n != nil; n = n.NextSibling() {
			if nh, ok := n.(*ast.Heading); ok && nh.Level <= h.Level {
				break
			}
			block.Body = append(block.Body, n)
		}
		blocks = append(blocks, block)
	}

	return blocks
}
