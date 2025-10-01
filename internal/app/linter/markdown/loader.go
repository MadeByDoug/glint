// Package markdown provides helpers to parse Markdown files into
// a Goldmark AST for downstream linting rules.
package markdown

import (
    "fmt"
    "os"

    "github.com/yuin/goldmark"
    "github.com/yuin/goldmark/ast"
    "github.com/yuin/goldmark/extension"
    "github.com/yuin/goldmark/parser"
    "github.com/yuin/goldmark/text"
)

// Document holds the parsed AST root and the original source bytes.
type Document struct {
    // Root is the Goldmark AST root node for the document.
    Root ast.Node
    // Src is the original Markdown source used to build the AST.
    Src  []byte
}

// LoadFile reads a Markdown file from disk and parses it into a Goldmark AST.
// The returned Document contains both the root AST node and the original source bytes.
func LoadFile(path string) (*Document, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("read markdown file %q: %w", path, err)
    }

    md := goldmark.New(
        goldmark.WithExtensions(
            extension.GFM, // GitHub Flavored Markdown: tables, strikethrough, task lists, etc.
        ),
        goldmark.WithParserOptions(
            parser.WithAutoHeadingID(),
        ),
    )

    reader := text.NewReader(data)
    root := md.Parser().Parse(reader)

    return &Document{Root: root, Src: data}, nil
}

