Contributing to Glint

Thank you for contributing! This project uses an RFC-driven, trunk-based workflow to keep main stable and make collaboration with AI agents smooth, without governance overhead.

How We Work

- main is always production-ready and protected.
- RFC branches are medium-lived and host both the design document and implementation work.
- Agents contribute via PRs into RFC branches; humans review and guide.

Before You Start

- Use Go 1.22+.
- Install tools used by CI checks:
  - golangci-lint: `go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest`
  - govulncheck: `go install golang.org/x/vuln/cmd/govulncheck@latest`

Quick Local Checks

- `go test ./... -race -shuffle=on`
- `go vet ./...`
- `golangci-lint run`
- `govulncheck ./...`
- `go run ./cmd/glint --help`

RFC Process (Simple)

1) Propose the idea
- Open a new issue using “RFC Proposal”.
- Pick the next sequential ID using `RFC-YYYY-###` (e.g., `RFC-2025-002`). The counter resets each calendar year and is zero-padded to three digits.

2) Create the RFC branch
- Create a branch that starts with `rfc/` and includes the RFC ID (e.g., `rfc/RFC-2025-002` or `rfc/RFC-2025-002/initial-draft`).

3) Write the draft RFC (two phases)
- Path: `rfcs/`
- Filename: `RFC-YYYY-###-<short-slug>.md` (e.g., `RFC-2025-002-folder-name-validation.md`). The slug is a short, lowercase, hyphenated summary.
- Start each RFC with a short metadata preface:
  - Start Date: `{YYYY-MM-DD}`
  - RFC Status: `Draft` | `Accepted` | `Rejected`
  - Owners: `@handle`
  - Optional fields such as `Reviewers`, `Dependencies`, or `Touched Paths` may be listed as additional bullet points.

  Draft phases
  - Phase 1 — Use Case & Design Goals
    - Complete: Summary, Motivation, Design Goals, Backwards Compatibility, Open Questions
    - Goal: Align on problem, scope, and desired properties.
  - Phase 2 — Design Overview, Implementation & Test Plan
    - Complete: Design Overview (architecture, alternatives, trade-offs), Implementation Plan (types/structs/functions, file locations, config changes), Test Plan (unit/e2e, coverage), Risks/Mitigations, Rollback Strategy
    - Goal: Align on concrete design and how it will be verified.

  Review etiquette
  - Keep RFC changes focused and easy to review.
  - If a phase changes substantially, update the “Status” accordingly in the RFC.

  Clean commit history for RFCs
  - Iterate & refine: free-form commits on the RFC branch while evolving content.
  - When the phase text is stable, optionally squash/rebase locally to produce a focused content commit (e.g., `docs(rfc): phase 1 complete for RFC-2025-002`).

4) Open a PR (RFC branch → main)
- During draft, PRs should be docs-only changes to the RFC file. The bot labels these “rfc:doc-only”.
- Iterate until accepted; when the design is settled, update the metadata preface (e.g., set `RFC Status: Accepted`).

5) Implement the RFC
- Create feature branches from the RFC branch and open PRs back into it.
- Agents should use: `agent/<rfc-id>/<topic>` (e.g., `agent/RFC-2025-002/parser`).
- Required checks for PRs to RFC branches:
  - Semantic PR title (Conventional Commits)
  - Vet, lint, tests (with race), govulncheck pass
  - Coverage ≥ 60% total
  - No DCO required
  - Optional automerge for RFC PRs (maintainers): set repo var `AUTOMERGE_RFC=true` and label the PR `automerge` (or `agent:automerge`).

6) Keep the RFC branch current
- When `main` changes, a workflow opens PRs to merge `main` into RFC branches. Resolve conflicts there.

7) Merge to main when release-ready
- Once the implementation and tests are complete, merge the RFC branch to `main`.

Notes
- Release Please prepares the changelog; tagging `vX.Y.Z` triggers GoReleaser to publish binaries.

Cross-RFC Awareness

- On pushes to `main` and `**.rfc*` branches, a workflow rebuilds `rfcs/active-index.json` (on `main`) with all active RFCs, including each RFC's title and metadata preface.
- RFC validation warns about potential overlaps (based on `touched_paths`) in PRs changing RFCs.

Conventions Summary

- RFC ID: `RFC-YYYY-###` (e.g., `RFC-2025-002`)
- RFC file: `rfcs/RFC-YYYY-###-<short-slug>.md`
- RFC branch: `rfc/<RFC-id>` (optionally append a topic suffix)
- Draft status: managed via the `RFC Status:` line in the metadata preface
- Commits: PR titles follow Conventional Commits (no DCO requirement)
- Go source files: multi-word filenames use lowercase snake_case (e.g., `folder_name.go`, `folder_name_test.go`).

Starter RFC (copy into `rfcs/RFC-YYYY-###-<short-slug>.md`)

# RFC: Short Title

- Start Date: YYYY-MM-DD
- RFC Status: Draft
- Owners: @your-handle
- Touched Paths:
  - cmd/glint/**
  - internal/app/**
- Dependencies: []

Summary

Brief overview of the proposal and expected outcome.

Motivation

Why this change is needed and what it solves.

Design Goals

Desired properties and constraints; leave detailed design overview to phase 2.

Implementation Plan

- Milestones
- Risks and mitigations
- Rollback strategy

Test Plan

- Unit/integration tests; coverage goal ≥ 60%
- Failure scenarios and edge cases

Backwards Compatibility

Breaking changes, migrations, and deprecations.
