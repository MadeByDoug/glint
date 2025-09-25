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

	headings := collectHeadingNodes(doc.Root)
	return buildSectionBlocks(headings)
}

func collectHeadingNodes(root ast.Node) []*ast.Heading {
	if root == nil {
		return nil
	}
	var headings []*ast.Heading
	walker := func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if h, ok := n.(*ast.Heading); ok {
			headings = append(headings, h)
		}
		return ast.WalkContinue, nil
	}
	if err := ast.Walk(root, walker); err != nil {
		return nil
	}
	return headings
}

func buildSectionBlocks(headings []*ast.Heading) []SectionBlock {
	if len(headings) == 0 {
		return nil
	}
	blocks := make([]SectionBlock, 0, len(headings))
	for _, h := range headings {
		blocks = append(blocks, buildBlockForHeading(h))
	}
	return blocks
}

func buildBlockForHeading(h *ast.Heading) SectionBlock {
	block := SectionBlock{Heading: h}
	if h == nil {
		return block
	}
	for node := h.NextSibling(); node != nil; node = node.NextSibling() {
		if nh, ok := node.(*ast.Heading); ok && nh.Level <= h.Level {
			break
		}
		block.Body = append(block.Body, node)
	}
	return block
}
