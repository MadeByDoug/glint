# RFC: Artifact Selector

- Start Date: 2025-09-25
- RFC Status: Draft
- Owners: @maintainers

## Summary
Define a consistent, composable selector that targets repository artifacts (folders or files) for linter checks. A selector specifies the artifact kind, a set of path patterns, and optional metadata filters. Selectors gate which nodes a check evaluates; they do not emit diagnostics themselves.

## Motivation

Checks should apply only where intended. A shared selector primitive provides:

- Predictable targeting across all checks (folders and files).
- Clear configuration for path- and metadata-based inclusion.
- A foundation for performance optimizations (e.g., selective file enrichment).

## Design

### Scope and Targeting
- Selectors target artifacts of a specific kind: `folder` or `file`.
- Path matching is performed against the repository-relative path using forward slashes, independent of OS.
- The repository root has path `/`. 

### Evaluation Order

For each node visited during a lint run, selection proceeds as:

1) Kind filter: the node must match `selector.kind` (`folder` maps to directories; `file` to files).
2) Metadata filter: if `selector.meta` is provided, the node must contain all keys with values equal to the configured strings.
3) Path filter: the node’s relative path (without leading `/`) must match at least one pattern from `selector.path`.

### Configuration Shape

Each rule provides a `selector` object with the following fields:

- `kind` (string, required): `folder` | `file`.
- `path` (array<string>, required): one or more regular expressions tested against the node’s relative path. Patterns are OR’ed (union) and are anchored (`^…$`) by the planner.
- `meta` (map<string,string>, optional): exact-match metadata constraints; all key/value pairs must match.

Notes:

- The engine trims the leading slash from node paths before applying the regexes.
- Regexes are validated on load; invalid expressions fail configuration.
- Known built-in metadata keys include `relPath` and `absPath` (attached during tree build). Additional keys may be populated by file parsers.

### Reporting

Selectors do not report diagnostics. They determine which nodes downstream checks evaluate. Any diagnostics are produced by the checks themselves.

## Examples

Select immediate child folders of `src`:

```yaml
selector:
  kind: folder
  path: ["src/[^/]+"]
```

Select any file under `docs/` with a `.md` extension:

```yaml
selector:
  kind: file
  path: ["docs/.*\\.md"]
```

Select service folders by metadata (e.g., tagged during tree construction):

```yaml
selector:
  kind: folder
  path: ["services/[^/]+"]
  meta:
    layer: "svc"
```

## Non-Goals

- A general-purpose query language; selectors expose only kind, path-regex, and exact-match metadata.
- Glob syntax; patterns are regular expressions (anchored by the engine).
- Emitting diagnostics; selection is purely a gating mechanism.

## Open Questions

- Should we support boolean composition across metadata (e.g., OR) or nested selectors?
- Do we need case-insensitive matching options for paths?
- Should path anchoring be configurable, or remain implicit for stability?
