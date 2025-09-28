# Agents Guide for Glint

This page highlights agent-specific expectations. 

## Planning Guidelines

1. Documentation vs. Implementation: Identify mismatches between code comments, docstrings, READMEs, or API documentation and the actual implementation.

2. Naming Conventions: Check that identifiers (variables, methods, classes, constants, files, modules) follow a consistent style across the project. Highlight cases where similar entities are named differently without reason.

3. Design Patterns & Usage: Verify that patterns (factories, adapters, dependency injection, etc.) are applied consistently for similar use cases. Point out where equivalent problems are solved in different or contradictory ways.

4. Cross-Cutting Consistency: Look for differences in error handling, logging, configuration, or testing approaches across comparable modules.

## Coding Guidelines

1. Prefer small, easily verifiable and composable functions over large functions.

2. Always try to refactor and reuse existing code before adding new code. 

3. Always use SOLID principles, sound design patterns, and industry best practices. 

4. Always follow existing project conventions unless they're insecure, buggy, or fail to follow point #3. Raise your concern for discussion before proceeding.

## Validation (run locally before PRs)
- `go test ./... -race -shuffle=on`
- `go vet ./...`
- `golangci-lint run`
- `govulncheck ./...`
- `go run ./cmd/glint --help`
