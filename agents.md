# Agents Guide for glint

## Project Snapshot
- `glint` is a Go-based linter for aligning AI with project standards.  
- Project is linted using .golangci.yaml and a root-level .glint (or .glint.yaml) config file
- CLI entry point: `cmd/glint/main.go`; core packages live under `internal/app/...`.
  - `internal/app/config`:
  - `internal/app/infra`:
  - `internal/app/lint`:
  - `internal/app/runntime`:

## Environment Setup

The below tooling is a required part of change validation. They need to be installed if not present on the command line. 

- go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
- go install golang.org/x/vuln/cmd/govulncheck@latest

## Coding Conventions

### Architecture
- Focus on small, single-purpose, composable functions that are obviously correct.
- Apply SOLID principles and use established design patterns judiciously.
- Use existing conventions, patterns, and libraries where possible.
- Keep the cognitive and cyclomatic complexity at 8 or lower. 
- Propagate `context.Context` through public entry points; avoid hidden globals beyond registered singletons.
- Determinism matters: collections are sorted before returning (see `lint.Lint` and `fs.BuildTreeFromFS`).

### Communication
- All Go sources carry a repo-relative header comment (e.g., `// internal/app/lint/runner.go`). 
- Comments are purposeful and high-signal—prefer brief doc comments that explain why something exists over narrating what the code does. There's no need to narrate where your changes occured via comments.
- Wrap errors with context using `%w`, and favor explicit messages (`fmt.Errorf("build tree: %w", err)`).

### Style
- Code is formatted with `gofmt`; stick to idiomatic Go spacing and short inline `if err != nil { return err }` patterns.
- Put types and structs at the top of files, followed by public functions, then private functions. 
- Sort language constructs alphabetically.

## Testing Conventions

- When wiring checks, compose selectors with `lint.WithSelector` and keep check registration centralized in `internal/app/lint/plan/factory.go`.

### Architecture
- Group related test scenarios into test functions.
- Prefer table-driven tests with descriptive `name` fields and explicit expectations via `testutil.Expect`.
- Tests live beside the code under `internal/app/lint/checks/tests/...` and use `testutil.B` builders for synthetic trees.

## Validation
Run the below commands to validate your changes. Report and suspected false positives to the user. 

- `go test ./...`
- `go vet ./...`
- `golangci-lint run`
- `govulncheck ./...`
- `go run ./cmd/glint`
