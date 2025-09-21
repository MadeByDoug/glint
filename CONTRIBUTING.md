Contributing to Glint

Thank you for contributing! This project uses an RFC-driven, trunk-based workflow with automation to keep main stable and make collaboration with AI agents smooth.

How We Work

- main is always production-ready and protected.
- RFC branches are medium-lived and host both the design document and implementation work.
- Agents contribute via PRs into RFC branches; humans review and guide.

Before You Start

- Use Go 1.24+.
- Install tools used by CI checks:
  - golangci-lint: `go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest`
  - govulncheck: `go install golang.org/x/vuln/cmd/govulncheck@latest`

Quick Local Checks

- `go test ./... -race -shuffle=on`
- `go vet ./...`
- `golangci-lint run`
- `govulncheck ./...`
- `go run ./cmd/glint --help`

RFC Process (Human-Friendly)

1) Propose the idea
- Open a new issue using “RFC Proposal”.
- Pick an ID with monthly rollover: `YYYY-mon.rfc#` (e.g., `2025-sep.rfc1`). `#` increments per month.

2) Create the RFC branch (choose one)
- Legacy (supported): `rfc/yyyy-mon-###` (lowercase month)
- Monthly (recommended): `YYYY-Mon.rfcN` (e.g., `2025-Sep.rfc2`)

3) Write the draft RFC
- Path: `rfcs/YYYY/Mon/`
- Filename: `YYYY-mon.rfc#-DRAFT.md` (e.g., `2025-sep.rfc1-DRAFT.md`)
- Keep `-DRAFT` while under discussion; remove it once accepted.
- Optional front matter (YAML) at the top helps coordination:
  ---
  owners:
    - @your-handle
  touched_paths:
    - cmd/glint/**
    - internal/app/**
  dependencies:
    - 2025-sep.rfc2
  ---

4) Open a PR (RFC branch → main)
- During draft, PRs should be docs-only changes to the RFC file. The bot labels these “rfc:doc-only”.
- Iterate until accepted; then remove `-DRAFT` in a follow-up PR.

5) Implement the RFC
- Create feature branches from the RFC branch and open PRs back into it.
- Agents should use: `agent/<rfc-id>/<topic>` (e.g., `agent/2025-sep.rfc1/parser`).
- Required checks for PRs to RFC branches:
  - Semantic PR title (Conventional Commits)
  - DCO sign-offs on every commit (`git commit -s`)
  - Vet, lint, tests (with race), govulncheck pass
  - Coverage ≥ 60% total
- Optional automerge for RFC PRs (maintainers): set repo var `AUTOMERGE_RFC=true` and label the PR `automerge` (or `agent:automerge`).

6) Keep the RFC branch current
- When `main` changes, a workflow opens PRs to merge `main` into RFC branches. Resolve conflicts there.

7) Merge to main when release-ready
- Once the implementation and tests are complete, merge the RFC branch to `main`.
- Release Please prepares the changelog; tagging `vX.Y.Z` triggers GoReleaser to publish binaries.

Cross-RFC Awareness

- On pushes to `main`, `rfc/**`, and `**.rfc*` branches, a workflow rebuilds `rfcs/active-index.json` (on `main`) with all active RFCs.
- RFC validation warns about potential overlaps (based on `touched_paths`) in PRs changing RFCs.

Conventions Summary

- RFC ID: `YYYY-mon.rfc#` (e.g., `2025-sep.rfc1`)
- RFC file: `rfcs/YYYY/Mon/YYYY-mon.rfc#(-DRAFT).md`
- RFC branch: `rfc/yyyy-mon-###` or `YYYY-Mon.rfcN`
- Draft status: while the `-DRAFT` suffix is present
- Commits: DCO sign-off required; PR titles follow Conventional Commits

Starter RFC (copy into `rfcs/YYYY/Mon/YYYY-mon.rfc#-DRAFT.md`)

---
owners:
  - @your-handle
touched_paths:
  - cmd/glint/**
  - internal/app/**
dependencies: []
---

# Glint RFC YYYY-mon.rfc#: Short Title

Summary

Brief overview of the proposal and expected outcome.

Motivation

Why this change is needed and what it solves.

Design Overview

High-level architecture, alternatives considered, trade-offs.

Implementation Plan

- Milestones
- Risks and mitigations
- Rollback strategy

Test Plan

- Unit/integration tests; coverage goal ≥ 60%
- Failure scenarios and edge cases

Backwards Compatibility

Breaking changes, migrations, and deprecations.

DCO and Conventional Commits

- DCO: each commit must include “Signed-off-by: Your Name <you@example.com>” (`git commit -s`).
- Conventional Commits: PR titles start with `feat:`, `fix:`, `docs:`, `chore:`, etc.

