// internal/app/linter/check/interfaces.go
package check

import (
	"github.com/MadeByDoug/glint/internal/app/infra/config"
	"github.com/MadeByDoug/glint/internal/app/infra/reporting"
	"github.com/MadeByDoug/glint/internal/app/linter/selector"
	"github.com/yuin/goldmark/ast"
)

// --- File-Based Check Interface & Context ---

// FileExecutionContext provides context for checks that run on a collection of files.
type FileExecutionContext struct {
	ProjectRoot string
	Config      *config.Config
	CheckConfig config.CheckConfig
	Artifacts   []selector.Artifact
}

// FileCheck defines the interface for checks that operate on files/directories.
type FileCheck interface {
	ExecuteOnFiles(ctx FileExecutionContext) ([]reporting.Issue, error)
}

// --- Policy Check Interface & Context (Node-Based) ---

// PolicyExecutionContext provides context for checks that run on a specific AST node.
type PolicyExecutionContext struct {
	ProjectRoot  string
	Config       *config.Config
	CheckConfig  config.CheckConfig
	TargetNode   ast.Node
	TargetSource []byte
}

// PolicyCheck defines the interface for checks that operate on an AST node.
type PolicyCheck interface {
	ExecuteOnNode(ctx PolicyExecutionContext) ([]reporting.Issue, error)
}