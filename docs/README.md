# Glint: A Directory Naming Linter

A Product Owner asks a BSA to display a VIP badge for high-value clients with a historical lifetime net value of more than $100,000. The developer rendered the badge based on a CRM tag sync for gross value, while QE tested future value. Confusion ensued. 

This is an over-simplified example based on the [Tree Swing Story](https://pmac-agpc.ca/project-management-tree-swing-story), but it highlights the difficulties we face keeping teams aligned, as one missing or misunderstood term can send people in the wrong direction. 

Glint is an exploratory project to see how linting can keep human teams aligned with our new AI co-workers. 

## The Problem

### Aligning People

Anecdotally, people seem to have a tendency of forgetting 10% of team processes and standards. What's worse is that everyone forgets a different 10%. Time pressures, stress, and busy workloads exacerbate the problem. Coupled with a lack of domain knowledge and poor discovery tools (i.e., the information exists somewhere, but good luck finding it), people tend to have Swiss cheese gaps in their understanding for many problem domains. The information landscape is difficult to navigate, and it's easy for us to land on different interpretations of the same problem.

### Aligning AI

When left to its own devices, LLMs tend to start ignoring existing project conventions. They may add a second logging library, ignore existing logging patterns in favour of dumping stack traces, or start breaking project boundaries (ex: mixing logging and config code). Giving LLMs project guidelines and allowing them to iterate on their designs helps, but these are very soft-guardrails which requires keeping significant (person) time to review and re-prompt. 

## The Solution

Glint aims to align human/AI teams with hard guardrails. If processes and standards are not followed, work does not proceed. The expectation is that teams align to Glint more than Glint aligns to teams. If Glint were to fully align to how teams work, the status quo would remain; misalignments would remain. 

### Aligning People

The main mechanism for aligning people is with a text‑based RFC process. Completing an RFC will naturally align teams on the problem domain and solution by requiring the following before work begins:
- Clarification of goals and scope
- A thorough analysis of the current landscape, existing implementations, etc
Smaller projects can be concatenated together and fed to the LLM, which will give it complete context.

## AI Prompts

### Service Alignment

```
Review the following code with a focus on consistency and alignment. Your task is to:

1. Documentation vs. Implementation: Identify mismatches between code comments, docstrings, READMEs, or API documentation and the actual implementation.

2. Naming Conventions: Check that identifiers (variables, methods, classes, constants, files, modules) follow a consistent style across the project. Highlight cases where similar entities are named differently without reason.

3. Design Patterns & Usage: Verify that patterns (factories, adapters, dependency injection, etc.) are applied consistently for similar use cases. Point out where equivalent problems are solved in different or contradictory ways.

4. Cross-Cutting Consistency: Look for differences in error handling, logging, configuration, or testing approaches across comparable modules.

5. Backwards compatability: Highlight places where there are multiple implementations which exist for backwards purposes. Each such instance should be clearly documented as being explicitly required for backwards compatability support. 

When reporting issues:

* Clearly explain the inconsistency, why it matters, and how it could undermine readability, maintainability, or architectural coherence.

* Suggest specific, concrete improvements that bring the codebase into stronger alignment with a common vision and agreed-upon standards.

Your goal is to provide a detail-oriented, constructive review that strengthens the project’s internal consistency and professional polish.
```

## Features

-   **Path-based Targeting**: Apply rules to specific parts of your project using regular-expression selectors (e.g., `^src/components/[^/]+$`).
-   **Casing Rules**: Enforce casing predicates such as `kebab`, `snake`, `camel`, `pascal`, `lower`, or `upper` via predicates.
-   **Affix Rules**: Require or prohibit specific prefixes and suffixes (e.g., require `comp-` prefix, prohibit `_` prefix).
-   **Disallow Patterns**: Block specific directory names or names matching a regular expression (e.g., disallow `^legacy$`, `^temp-.*$`).
-   **Flexible Configuration**: Define rules in a project-root `.glint.yaml` file, override values via `GLINT_*` environment variables, or supply CLI flags for ad-hoc changes.
-   **Structured Diagnostics**: Choose between human-friendly text output or machine-readable JSON, and control how much detail is surfaced.
-   **Debuggable Runs**: Enable debug mode to inspect the normalized runtime and linter configuration that Glint executes with.

## Configuration

`glint` layers configuration additively, starting with baked-in defaults and then walking outward-in:

1.   `$HOME/.glint.yaml`
2.   The current working directory’s `.glint.yaml`
3.   An explicit `--config` file, if provided
4.   Environment variables prefixed with `GLINT_`
5.   CLI flags

Later sources override earlier ones, so you can keep shared organization defaults in `$HOME` and let per-project files or flag overrides tighten the rules. If you need to run without a home-directory baseline, temporarily move or rename that file before executing Glint.

## Project Layout

Glint’s repository layout is intentionally constrained. At the root level, these directories are expected (enforced by `.glint.yaml`):

- `.glint` (optional local directory) and `.glint.yaml` (project config file)
- `.github` (workflows and repo automation)
- `cmd/glint` (CLI entrypoint)
- `internal/app` (application code: `config`, `infra`, `lint`, `runtime`)
- `rfcs` (design documents; see `CONTRIBUTING.md` for naming)
- `schemas` (JSON/YAML schemas)
- `docs` (user and developer documentation)

### Example Configuration

Here is an example that demonstrates several common rules for directory names:

```yaml
# .glint.yaml
linter:
  version: 1
  rules:
    - id: R101
      selector:
        # Select all directories directly under src/components
        kind: folder
        path:
          - "^src/components/[^/]+$"
      apply:
        checks:
          - folderName:
              # Enforce that component folders are in kebab-case.
              predicates:
                - "kebab"
              severity: "error"

    - id: R102
      selector:
        # Select every directory in the project
        kind: folder
        path:
          - ".+"
      apply:
        checks:
          - folderName:
              # Prohibit temporary or private folders from being committed.
              prohibitPrefix:
                - "_"
              disallow:
                - "^temp-.*$"
                - "^tmp-.*$"
              severity: "warn"

    - id: W201
      selector:
        # Select directories at the root of the project
        kind: folder
        path:
          - "^[^/]+$"
      apply:
        checks:
          - folderName:
              # Discourage legacy patterns at the top level.
              disallow:
                - "^legacy$"
                - "^old$"
              message: "Avoid creating top-level legacy directories."
              severity: "warn"
```

## Usage

Download the source code repository and build with Go.

```bash
go build ./cmd/glint
```

To run the linter on the current project, navigate to the project root and execute:

```bash
glint --dir . \
      --config /path/to/config.yaml \
      --env prod \
      --debug \
      --diag-level warn \
      --error-format json \
      --error-output stdout \
      --log-output logs/glint.log
```

### Common CLI Flags

`--dir`
: Root directory whose tree will be linted.

`--config`
: Absolute or relative path to an additional config file loaded after the home/project defaults; values in this file override earlier layers.

`--env`
: Selects an environment-specific config overlay (e.g., `dev`, `stg`, `prod`).

`--debug`
: Writes normalized application and linter configuration to the debug stream for inspection.

`--diag-level`
: Diagnostic verbosity (`warn` by default; set to `info` to include informative notes).

`--error-format`
: Output format for diagnostics (`text`, `json`).

`--error-output`
: Destination for diagnostics (`stderr`, `stdout`, or file path).

`--log-output`
: Destination for structured logs (`stderr`, `stdout`, `none`, or file path).

`--log-level`
: Logging verbosity (`warn` by default; set to `info` for additional detail).

`--error-output` and `--log-output` accept filesystem paths; `glint` cleans and validates these to prevent escaping the intended directory.

Environment variables prefixed with `GLINT_` mirror the configuration hierarchy. Double underscores (`__`) denote nested keys, and values are merged before CLI overrides are applied. CLI flags always win over environment variables, which themselves win over config files, allowing you to tweak behaviour without editing committed configuration.
