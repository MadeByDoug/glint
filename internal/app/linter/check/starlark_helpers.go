// internal/app/linter/check/starlark_helpers.go
package check

import (
	"github.com/yuin/goldmark/ast"
	"go.starlark.net/starlark"
)

// buildStarlarkPredeclared creates the dictionary of globally available functions and
// constants for a Starlark policy script. This defines the API that users can
// script against.
func buildStarlarkPredeclared(ctx PolicyExecutionContext) starlark.StringDict {
	predeclared := starlark.StringDict{}

	// text() -> string
	// Returns the text content of the current node.
	predeclared["text"] = starlark.NewBuiltin("text", func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		if err := starlark.UnpackArgs(b.Name(), args, kwargs); err != nil {
			return nil, err
		}
		nodeText := string(ctx.TargetNode.Text(ctx.TargetSource))
		return starlark.String(nodeText), nil
	})

	// level() -> int | None
	// For Heading nodes, returns the heading level (1-6). For other nodes, returns None.
	predeclared["level"] = starlark.NewBuiltin("level", func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		if err := starlark.UnpackArgs(b.Name(), args, kwargs); err != nil {
			return nil, err
		}
		if h, ok := ctx.TargetNode.(*ast.Heading); ok {
			return starlark.MakeInt(h.Level), nil
		}
		return starlark.None, nil
	})

	// kind() -> string
	// Returns the type of the node as a string (e.g., "Heading", "Paragraph").
	predeclared["kind"] = starlark.NewBuiltin("kind", func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		if err := starlark.UnpackArgs(b.Name(), args, kwargs); err != nil {
			return nil, err
		}
		return starlark.String(ctx.TargetNode.Kind().String()), nil
	})

	return predeclared
}