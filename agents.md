# Agents Guide for Glint

This page highlights agent-specific expectations. For the full workflow, read `CONTRIBUTING.md`.

Agent Basics

- Follow the RFC workflow in `CONTRIBUTING.md` before branching or opening PRs.
- Use the RFC ID format `YYYY-mon.rfc#` (e.g., `2025-sep.rfc1`).
- Implementation branches must be named `agent/<rfc-id>/<topic>` and target the RFC branch via PR.
- Keep RFC files in `rfcs/YYYY/Mon/` and use the filename pattern `YYYY-mon.rfc#(-DRAFT).md`.
- While an RFC is under discussion, keep the `-DRAFT` suffix in the filename.
- Check `rfcs/active-index.json` on `main` to understand other active RFCs and avoid conflicts.

Validation (run locally before PRs)

- `go test ./... -race -shuffle=on`
- `go vet ./...`
- `golangci-lint run`
- `govulncheck ./...`
- `go run ./cmd/glint --help`

Quality and Commits

- PR titles follow Conventional Commits (feat:, fix:, docs:, chore:, etc.).
- Every commit includes a DCO sign-off (`git commit -s`).
- CI requires vet, lint, tests (race), govulncheck, and ≥ 60% total coverage to pass.

Code & Tests

- Favor small, composable functions; keep complexity low (≤ 8).
- Propagate `context.Context` in public entry points.
- Write table-driven tests; add or update tests for new behavior.
