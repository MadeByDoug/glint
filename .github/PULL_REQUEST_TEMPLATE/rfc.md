---
name: "RFC: Design & Plan"
about: "Propose or update an RFC document"
title: "docs(rfc): <short title>"
---

Checklist (Draft phase)

- [ ] This PR targets `main` and only changes a single DRAFT RFC markdown file.
- [ ] RFC filename follows: `YYYY-mon.rfc#-<short-slug>(-DRAFT).md` under `rfcs/YYYY/Mon/` (slug is kebab-case).
- [ ] RFC branch name is either `rfc/yyyy-mon-###` or `YYYY-Mon.rfcN`.
- [ ] Status is clear (keep `-DRAFT` suffix until accepted).
- [ ] Front matter includes `summary` (≤ 140 chars), and optionally `touched_paths`, `dependencies`, `owners`.

Links

- RFC ID: <!-- e.g., 2025-sep.rfc1 -->
- Related Issues: <!-- #123 -->

Summary (≤ 140 chars)

<!-- One-line short description. Also add to front matter as `summary:`. -->

Motivation

<!-- Why this change is needed. -->

Design Overview

<!-- High-level design with diagrams or pseudo as needed. -->

Implementation Plan

<!-- Phased plan; milestones, risks, rollbacks. -->

Test Plan

<!-- Unit, integration, e2e strategy; coverage areas. -->

Backwards Compatibility

<!-- Breaking changes and migration steps, if any. -->
