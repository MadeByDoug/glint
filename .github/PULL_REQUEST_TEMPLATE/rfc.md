<!-- Keep this PR template minimal. The RFC in rfc/ is the source of truth. -->

## RFC
- File: `rfc/2025-sept.rfc1-feature-x.md`

## Feature Gates
- Primary gate: `2025-sept.rfc1`
- Additional gates (if any): <!-- comma-separated -->

## Scope & Impact
- Area(s): <!-- e.g., parser, cli, http -->
- Behavior with gates **OFF**: unchanged ✅
- Behavior with gates **ON**: <!-- 1–2 lines of WHAT changes (not how) -->

## Readiness Checklist
- [ ] All code paths are **behind the gates** listed above
- [ ] Default behavior (gates OFF) covered by tests and remains green
- [ ] Gate-ON tests added (unit/integration as appropriate)
- [ ] `go vet`, `golangci-lint`, `govulncheck` pass
- [ ] Coverage ≥ threshold (OFF mode)
- [ ] Functional test added for any bug fixed in this PR (if applicable)
- [ ] Docs/help updated (only user-visible behavior)

## RFC State
- [ ] Planning
- [ ] Implementation
- [ ] Beta
- [ ] Release Candidate
- [ ] Released

## Notes for Reviewers (optional)
<!-- Routing hints, screenshots, perf notes, rollout plan, etc. -->
