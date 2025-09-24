## Phase 1: Scope & Goals
**Overview**
- Build markdown schema validator
- Document required RFC structure
- Align human and AI collaborators
- Support deterministic review workflows
- Ensure list validation works
- Maintain schema in YAML
- Enable future extensions
- Track implementation readiness
- Improve collaboration signals

Goals and non-goals:
- Goal: Define RFC baseline
- Goal: Catch structural drift
- Non-Goal: Replace human review
- Goal: Align release timelines

## Phase 2: Implementation
- Use goldmark for parsing
- Walk AST sections in order
- Validate sequence rules

- Risk: Parser mismatch with schema
- Risk: Missing section detection gaps

## Phase 3: Release & Review
**Release Intent**
- Preview: Yes
- Beta: No
- GA: No

Reviewers:
- maintainer@example.com
- reviewer@example.com
