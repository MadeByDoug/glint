# RFC: File Name Validation

- Start Date: 2025-09-25
- RFC Status: Draft
- Owners: @maintainers

## Summary

Introduce a first-class `fileName` linter check that enforces naming rules for files, analogous to the existing `folderName` check. The feature plugs into the current selector + check registry, allowing users to target files via regex selectors and apply consistent predicates (e.g., kebab, lowercase), allow/disallow lists, and affix rules.

## Motivation

Projects often require consistent naming for source files, configs, and documents (e.g., RFCs, markdown docs). Glint already supports folder naming; extending parity to files reduces friction and enables:

- Consistent casing of file basenames (kebab, lowercase, etc.)
- Policy on prefixes/suffixes (e.g., `*.spec`, `_`-prefixed files forbidden)
- Allow/deny lists by pattern (e.g., ban temporary files, constrain RFC names)
- Targeted enforcement with existing `selector.kind: file`

This brings the file layer under the same policy surface as directories without introducing new concepts.

## Guide-Level Explanation

Add a new check type `fileName`. Users select files with a selector and apply rules just like `folderName`.

Example: enforce kebab-case RFC filenames under `rfcs/` and allow common dotfiles elsewhere.

```yaml
linter:
  version: 1
  rules:
    - id: RFC-FILE-NAMES
      selector:
        kind: file
        path:
          - "^rfcs/.+\\.md$"
      apply:
        checks:
          - type: fileName
            params:
              severity: error
              predicates: ["kebab"]
              # Typical dotfiles can be allowed explicitly when needed
              allow:
                - "^README$"          # if matching basename
              disallow:
                - "^draft_.*$"
              message: "RFC filenames must be kebab-case and not start with 'draft_'"

    - id: NO-TEMP-FILES
      selector:
        kind: file
        path: [".+"]
      apply:
        checks:
          - type: fileName
            params:
              severity: warn
              disallow: ["^tmp-.*$", "^.*~$"]
              message: "Temporary files should not be committed"
```

Notes:

- Predicates apply to the file “subject name” (see Reference-Level). By default this is the basename without extension.
- Allow short-circuits other checks; disallow takes precedence over other violations, mirroring `folderName`.

## Reference-Level Explanation

New check: `fileName` with parameters modeled after `folderName` for familiarity and reuse.

- Config (internal): `ChFileName` mirrors `ChFolderName` in `internal/app/config/model/linter.go`.
  - fields: `predicates`, `allow`, `disallow`, `prefix`, `suffix`, `prohibitPrefix`, `prohibitSuffix`, `message`, `severity` (optional override)
  - new field: `target` (string, optional) with values:
    - `basename` (default): validate the basename without extension
    - `full`: validate the full filename including extension
- Registry: register `"fileName"` in `internal/app/lint/plan/factory.go` and map to a new checker in `internal/app/lint/checks/file`.
- Checker: `internal/app/lint/checks/file/file_name.go` with behavior:
  - `ID() string`: `"check.fileName"`
  - `ApplyToNode` only considers `n.Kind == lint.File`
  - Determine the subject name based on `target` (default basename via `filepath.Ext` + trim)
  - Precedence and messages match `folderName`:
    - If `allow` matches => pass
    - Else if `disallow` matches => single diagnostic with pattern hit
    - Else apply predicates and affix checks, emitting one diagnostic per failure
  - `message` prepends custom context to each emitted diagnostic
- Selector: reuse existing `selector.kind: file` and compiled `path` regexes; no changes required (`internal/app/lint/plan/from_config.go`, `internal/app/lint/selector.go`).

Diagnostics format (consistent with folderName):

- Code: provided by rule id mapped through plan/factory
- Msg: `"<path>: <custom message (optional)> (<reason>)"`
- Severity: from config override if provided; else default mapping to error

## Drawbacks

- Slight duplication between folder and file checkers. We can factor shared logic later if desired.
- The `target` option introduces one more knob; default keeps behavior intuitive.

## Rationale and Alternatives

- Rationale: Mirror the proven `folderName` semantics for a predictable user experience, minimizing new config surface.
- Alternative: A single generic `name` check for both files and folders with a `subject: file|folder` parameter. Chosen approach avoids refactors now and stays incremental.
- Alternative: Only regex-based rules. Adding predicates maintains parity and ergonomics for common casing rules.

## Prior Art

- Existing `folderName` check with predicates, allow/disallow, and affix constraints (`internal/app/lint/checks/folder/folder_name.go`).
- Current plan registry and selector support for `kind: file` without a file-name checker yet (`internal/app/lint/plan/factory.go`).

## Implementation Plan

1) Config model
- Add `ChFileName` to `internal/app/config/model/linter.go` (fields listed above) with strict JSON/YAML decoding.
- Update any schema comments/docs accordingly.

2) Plan registry
- In `internal/app/lint/plan/factory.go`, register `registry["fileName"]`.
- Implement `newFileNameCheck` that unmarshals `ChFileName`, maps severity/message, converts predicates, and constructs the checker.

3) Checker implementation
- Add `internal/app/lint/checks/file/file_name.go` modeled after `folder_name.go` with:
  - subject-name extraction by `target`
  - identical precedence and message formatting
  - reuse of predicate validators (copy or share via a small internal helper)

4) Tests
- Mirror folder tests under `internal/app/lint/checks/tests/file/`:
  - Core functionality (predicate fail, disallow hit, affix checks, custom message, empty config)
  - Predicates specific cases for filenames (kebab, lowercase)
  - Precedence ordering (allow > disallow > others)
  - Basename vs full name `target` behavior

5) Docs
- Update `docs/README.md` with a `fileName` section and examples using `selector.kind: file`.
- Add example to `.glint.yaml` if we want self-dogfooding checks for RFC file names.

## Backward Compatibility

- No breaking changes. `fileName` is additive; existing configs and code paths remain unchanged.

## Security, Performance, and UX

- Security: No additional file I/O beyond what `EnrichTree` already does; validation is string-based on in-memory names.
- Performance: O(1) per file; regexes compiled once in constructor; negligible impact relative to tree build.
- UX: Keeps messages and precedence consistent with folders; `target` defaults to intuitive `basename`.

## Open Questions

- Should we offer built-in shortcuts for common file policies (e.g., `allowDotfiles: true`)? Proposal: rely on regex/affix for now to keep surface minimal.
- Do we need extra predicates for filenames (e.g., `snake`)? Proposal: follow folder predicates; extend both checks uniformly in a separate RFC.

## Rollout

- Land behind no flags; additive and off by default until configured.
- Add tests before merge; reuse existing BDD-style utilities in `internal/app/lint/checks/tests/testutil/`.
- Incremental refactor later to deduplicate name-checking logic across file/folder if it proves valuable.
