# Pleat

Pleat checks code layout for human readability in Go, Python 3, and plain TypeScript files.
It produces advisory findings for developers and coding agents.

## Build and run

Use Go 1.23 or newer and a C compiler.
On macOS, the Xcode Command Line Tools provide the compiler required by Tree-sitter.
Python and Node are not required to run the built executable.

Build from the repository root.

```sh
go build -o bin/pleat ./cmd/pleat
```

Check files with text output or JSON.
Options can appear before or after file arguments.

```sh
bin/pleat check internal/app/command.go
bin/pleat check internal/app/testdata/orders.go internal/app/testdata/orders.py internal/app/testdata/orders.ts --format json
```

The second command produces six advisory findings from the dense evaluation inputs.
Check complete tracked working-tree files changed against a commit with `--diff`.

```sh
bin/pleat check --diff HEAD --format json
bin/pleat --help
```

`--diff HEAD` includes net staged and unstaged changes.
It excludes untracked files and reports deletions and unsupported extensions as skipped.
Explicit file selection includes ignored and untracked files.
Explicit files and `--diff` cannot be combined.
Use `--` before a filename that starts with a dash.

Diff selection disables Git hooks, fsmonitor, external diff commands, text conversion, and configured clean and process filters.
Analysis does not execute project code or modify source.

## Tune the layout profile

Defaults live in `internal/app/config.go`.
Rule settings are defined in `internal/analysis/settings.go`.
The initial policies are provisional, not calibrated quality measures.

| Rule | Initial policy | Level |
| --- | --- | --- |
| `layout.max-statement-run` | More than six executable sibling statements without a blank line. | `medium` |
| `layout.control-cuddle` | A control statement needs separation unless the preceding group is one assignment mentioned in its header. | `medium` |
| `layout.return-separation` | A return needs separation when its block contains more than two executable statements. | `low` |

Multiline statements count once.
Comments do not count as statements or separate a run by themselves.
Nested blocks are checked independently.
Python docstrings and TypeScript type-only declarations do not count as executable statements.

Create `.pleat.toml` in the invocation directory to override the profile.
There is no parent-directory search.
Use `--config PATH` to select another file.

```toml
[rules."layout.max-statement-run"]
enabled = true
level = "medium"
max = 6

[rules."layout.control-cuddle"]
enabled = true
level = "medium"

[rules."layout.return-separation"]
enabled = true
level = "low"
block_limit = 2
```

Each rule accepts `enabled` and `level`.
Only the statement-run rule accepts `max`, which must be positive.
Only the return rule accepts `block_limit`, which must be nonnegative.
Unknown rules, unknown settings, invalid types, and invalid values are errors.

## Consume reports in an agent pipeline

JSON contains `status`, `findings`, `coverage`, and `errors`.
Each finding contains `rule`, `level`, `comment`, `path`, and a one-based `line`.
Findings sort by path, line, and rule name.
Text reports render the same findings and coverage.

Levels are `low`, `medium`, and `high`.
They express review priority, not correctness, confidence, or mandatory rejection.
An agent can address a finding or explain its disagreement.

| Exit code | Meaning |
| --- | --- |
| `0` | Advisory analysis completed without run errors. Findings, empty scope, and explicitly reduced coverage are allowed. |
| `2` | Invocation, configuration, file-read, parser, or runtime failure. Other files may retain useful findings. |

Always inspect coverage.
`complete` means all applicable checks ran on eligible files.
`partial` means an eligible file, region, or check was omitted.
`failed` means an operational failure prevented analysis altogether.
`empty` means selection found no eligible code.
A parser failure can produce a partial report and exit `2`.
A known unsupported construct can produce a partial report and exit `0`.

## Run behavior evals

```sh
go test ./...
go test -race ./...
go vet ./...
```

Package-local fixtures cover dense and grouped code, adjacency exceptions, comments, multiline statements, and damaged regions.
CLI evals cover configuration, mixed-language output, Git selection, and disabled repository helpers.
Tests assert findings and coverage rather than exact explanation wording.

## Limits and design

The current verification target is macOS arm64.
Linux packaging and a readability baseline remain future work.
Syntax support is best effort, with conservative recovery around damaged declarations or independent top-level regions.
There is no type checking, source rewriting, semantic grouping inference, metric backend, or built-in enforcement gate.

[DESIGN.md](DESIGN.md) describes the implemented package layout.
[VISION.md](VISION.md) records the product boundaries, parser tradeoffs, and calibration work.
