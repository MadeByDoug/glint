# RFC: Folder Name Validation

- Start Date: 2025-09-25
- RFC Status: Draft
- Owners: @maintainers

## Summary

Formalize the `folderName` linter check that enforces naming rules on directory basenames. The check supports casing predicates, allow/deny lists, and affix rules (required and prohibited) while using the existing selector pipeline. This RFC documents the configuration shape, evaluation order, and reporting behavior.

## Motivation

Consistent folder naming improves predictability, navigation, and tooling. Specifically, it enables:

- predictable casing for directories (kebab, snake, camel, pascal, lower, upper, etc.)
- enforced prefixes/suffixes (e.g., ban leading `_` for private dirs)
- simple allow/deny exceptions for special-purpose directories
- reuse of the familiar selector pipeline with `selector.kind: folder`

## Design

### Scope and Targeting
- Enforces naming rules on directories selected via the rule's `selector`.
- Only applies to directories below the repository root; the root itself is skipped.

### Evaluation Order

1. Allow list: if any `allow` regex matches the basename, the directory is accepted and no further checks run.
2. Deny list: if any `disallow` regex matches, emit one diagnostic and stop.
3. Predicates: evaluate in order; emit a diagnostic for each failing predicate. Continue to affixes.
4. Affixes: evaluate all affix requirements and prohibitions; each failure produces its own diagnostic.

### Configuration Shape

Each rule entry in `linter.rules` contains:

- `id` (string) — unique identifier for diagnostics.
- `selector` — identifies which folders the rule will inspect. (Selector semantics are documented separately.)
- `apply.checks` — list of concrete checks. For folder name validation, add a `folderName` item whose value configures the behaviour described below.

The `folderName` configuration supports:

- `predicates` (array<string>): supported values `kebab`, `snake`, `camel`, `pascal`, `lower`, `upper`.
- `allow` (array<string>): regex patterns that, when matched, short-circuit the check in favour of the folder.
- `disallow` (array<string>): regex patterns that immediately fail the folder.
- `prefix` (array<string>): required prefixes.
- `suffix` (array<string>): required suffixes.
- `prohibitPrefix` (array<string>): forbidden prefixes.
- `prohibitSuffix` (array<string>): forbidden suffixes.
- `message` (string, optional): prepended to emitted diagnostics.
- `severity` (enum: `off` | `error` | `warn` | `info`, optional): overrides the rule-level default (`none` is accepted as an alias for `off`).

Notes:

- Regexes are validated on load and then wrapped with `^…$` anchors by the planner.
- Empty configuration is a no-op.
- Unknown predicate names are configuration errors.

#### Predicate Evaluation

All configured predicates are evaluated; any that fail each emit a diagnostic. Emission order follows configuration order for stability.

### Reporting

- Diagnostics include the rule `id`, the directory’s relative path, and either a predicate reason or affix reason.
- Predicates: emit a diagnostic for each failing predicate.
- Affix failures may emit multiple diagnostics.
- If `message` is set, it prefixes the reason, e.g., `Custom (reason)`.

### Check Identifier

- Check type: `check.folderName`.

## Examples

Minimal policy enforcing kebab-case on all src subdirectories:

```yaml
linter:
  rules:
    - id: rfc-2025-001-kebab-dirs
      selector:
        kind: folder
        path:
          - src/[^/]+
      apply:
        checks:
          - folderName:
              predicates: [kebab]
```

Strict policy with exceptions and affixes:

```yaml
linter:
  rules:
    - id: rfc-2025-001-strict-dirs
      selector:
        kind: folder
        path:
          - .*
      apply:
        checks:
          - folderName:
              message: "Invalid folder name"
              severity: error
              allow: ["^\\.git$", "^node_modules$"]
              disallow: ["[A-Z]", "\\s"]
              predicates: [lower]
              prohibitPrefix: ["_"]
              suffix: ["-pkg"]
```

## Non-Goals

- Not a file name check.
- Does not enforce rules on the full path; only the final basename.
- No automatic renames or fixers are included in this RFC.
- OS-reserved names and filesystem portability are out of scope.

## Open Questions

- Should we expose additional casings such as title-case or locale-aware variants in a future release?
- Do we need an option to limit diagnostics to a single message per directory (including affixes), or is multiple affix reporting preferable?
