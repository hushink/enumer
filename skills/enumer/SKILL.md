---
name: enumer
description: >-
  Generate and maintain type-safe Go enumerations from an enumer.yaml manifest
  with the enumer CLI. Use this skill whenever a task involves Go enums or
  enum-like constant sets: declaring string, integer, or bitmask enums;
  generating constants plus String, Parse, IsValid, JSON, YAML, SQL, flag,
  names, or values helpers; adding enumer.yaml to a package; migrating away
  from a comment-based enum DSL; or wiring enumer generate/check into
  go:generate and CI. Trigger it when the user says "Go enum", "string enum",
  "type-safe enum", "enum with JSON/SQL methods", "bit flags", "Go 枚举",
  "枚举生成", or asks to generate code from enumer.yaml — even if they do not
  name enumer. Do not use it for enums in non-Go languages or for building an
  enum library by hand.
license: MIT
compatibility: >-
  Building the enumer CLI requires Go 1.27+; prebuilt binaries are available
  for Linux, macOS and Windows and need no Go toolchain. Generated code
  requires Go 1.20+. The yaml feature adds a dependency on gopkg.in/yaml.v3.
metadata:
  author: hushink
  version: "0.1.0"
  repository: https://github.com/hushink/enumer
---

# enumer: type-safe Go enums from YAML

`enumer` reads an `enumer.yaml` manifest and writes Go files containing the enum
type, constants and opt-in methods. The manifest is the single source of truth.
Generated files are formatted with the standard library `go/format`; never edit
them by hand.

## Golden rules

- **Only declared enums exist.** If an enum or value is not in the manifest, it
  is not generated. Add it to `enumer.yaml` first; removing it prunes its old
  generated file on the next run.
- **Edit the manifest, then regenerate.** Manual edits to `*_enum.go` are
  overwritten and make `enumer check` fail.
- **Do not duplicate the type declaration.** With the default `typeMode: auto`,
  enumer emits the type when it is missing and reuses a hand-written declaration
  when one exists. Never declare the same type twice in one package.
- **Verify with the Go toolchain.** Formatting is automatic, but compilation and
  behavior still need `go build ./...` and `go test ./...`.
- **Use `enumer check` as the CI gate.** It detects generated files that no
  longer match the manifest.

## Workflow

1. **Inspect the target package.**
   - Locate the Go package directory and its package name. `go list -f
     '{{.Name}}' .` is reliable; reading a `.go` file also works.
   - Check for an existing `enumer.yaml` or `enumer.yml` in that directory.
   - Check whether an enum type already exists: `type <Name> <underlying>` in a
     non-test `.go` file.
   - Check whether `enumer` or `go tool enumer` is available.
2. **Choose the manifest location.**
   - Put one `enumer.yaml` in the package directory when the enums belong to
     that package. Package inference and default output paths then do the right
     thing.
   - For several packages, add one manifest per package and run `enumer
     generate ./...`. Use `extends` for shared defaults.
   - If the directory has no Go files yet, set `package` explicitly.
3. **Write or update the manifest.** Preserve existing enums, features, comments
   and targets; make the smallest change that satisfies the request.
4. **Generate and verify.** Run the generator, then build and test the package.
5. **Report what changed.** List the manifest path, enums and features added or
   removed, generated files, commands run, and any new module dependency.

### Getting the CLI

Prefer an existing `enumer` binary or a project tool dependency. If neither is
available, choose the least invasive option:

```sh
# One-off run; no persistent install (Go 1.27+):
go run github.com/hushink/enumer/cmd/enumer@latest generate

# Reproducible project tool dependency (modifies go.mod/go.sum):
go get -tool github.com/hushink/enumer/cmd/enumer@latest
go tool enumer generate

# Global binary:
go install github.com/hushink/enumer/cmd/enumer@latest
```

Prebuilt binaries for Linux, macOS and Windows are available at
<https://github.com/hushink/enumer/releases> and need no Go toolchain. Building
the CLI from source (or using `go run`) requires Go 1.27+; generated code only
requires Go 1.20+.

For CI, pin the version instead of using `@latest`. A tool dependency plus
`go tool enumer` is the most reproducible setup.

## Manifest essentials

A minimal manifest in the package directory:

```yaml
version: 1
package: order

enums:
  - name: Status
    type: string
    values:
      - pending
      - paid
      - shipped
```

`enumer generate` writes `status_enum.go` next to the manifest. With the default
feature set it contains `Status`, `StatusPending`, `StatusPaid`,
`StatusShipped`, `String`, `ParseStatus`, `IsValid`, text marshaling,
`StatusNames` and `StatusValues`.

- `version` is optional and defaults to `1`; include `version: 1` explicitly.
- `package` is inferred from Go files in the manifest directory; set it
  explicitly when in doubt or when the directory has no Go files.
- `enums` must contain at least one enum, and each enum at least one value.
- A manifest can be named `enumer.yaml` or `enumer.yml`. Do not put both in the
  same directory.

### Values

Values are either scalars or mappings:

```yaml
values:
  - red
  - name: green
    value: GREEN          # string enums: canonical text
    aliases: [g]          # extra parseable texts; no constants
    comment: Green is the second color.
  - name: blue
    deprecated: use teal  # emits a Go Deprecated: comment
```

For **string enums**, the canonical text defaults to the value name; `value`
overrides it. The constant name is always derived from `name`, so `green` with
`value: GREEN` produces `ColorGreen` whose text is `"GREEN"`.

For **numeric enums**, `value` is the number. Without it, values auto-increment
from `start` (default `0`); an explicit value resets the sequence for following
auto values. The canonical text is the value name exactly as written.

`aliases` are parseable texts but do not generate constants. `deprecated` accepts
`true` or a reason string. `doc` and `comment` override generated comments;
`comment` wins.

### Types and naming

Supported underlying types: `string`, `int`, `int8`, `int16`, `int32`, `int64`,
`uint`, `uint8`, `uint16`, `uint32`, `uint64`. `type` defaults to `string`.

Default naming produces `ColorRed` from `red`. These settings change it:

```yaml
defaults:
  naming:
    prefix: type          # type | none | literal:<prefix>
    identifier: camel     # camel | raw
    extraInitialisms: [GRPC]
    disableInitialisms: false
```

- `prefix: type` uses the enum name (`ColorRed`); `none` uses the bare name
  (`Red`, which can collide); `literal:<prefix>` uses a fixed prefix.
- `identifier: camel` converts `http_status` to `HTTPStatus` and `blue-green`
  to `BlueGreen`, applying the built-in initialism list (`HTTP`, `URL`, `ID`, …).
- `identifier: raw` leaves the name unchanged and requires it to already be a
  valid Go identifier.

### Features

The default set is:

```text
string, parse, valid, text, names, values
```

`features` **adds** to the resolved defaults; `disableFeatures` **removes** from
them. An explicit `features: []` in `defaults` disables all features. Dependencies
are pulled in automatically, and conflicts are errors.

Choose features by intent:

| Goal | Add | Notes |
| --- | --- | --- |
| JSON as names/text | `json` | Default `text` already gives JSON round-trip; add `json` for explicit methods |
| JSON as numbers | `json:number` | Numeric enums only; conflicts with `json` |
| SQL `Scan`/`Value` | `sql` | Works for string and numeric enums |
| Nullable SQL + JSON | `sql:null` | Requires exactly one of `sql` or `sql:int` |
| Store a string enum as an integer column | `sql:int` | Every value needs `number`; conflicts with `sql` |
| CLI flag (`flag`/`pflag`) | `flag` | Adds `Set` and `Type` |
| Bitmask helpers | `flags` | Unsigned enums only; values zero or powers of two |
| YAML marshal/unmarshal | `yaml` | Adds `gopkg.in/yaml.v3` to the module |
| Panic-on-invalid parsing | `mustparse` | Requires `parse` |
| Pointer helper | `ptr` | Adds `Ptr()` |
| Value lists | `names`, `values` | Already in the default set |

Disabling `sql` also removes `sql:int` and `sql:null`; disabling `json` also
removes `json:number`.

### Type declarations: `typeMode`

| Mode | Behavior |
| --- | --- |
| `auto` (default) | Emit the type only when no declaration exists; reuse an existing one; error on aliases, generics or underlying-type mismatch. |
| `emit` | Always emit the type; an existing declaration in another file is an error. |
| `existing` | Never emit the type; a matching declaration must already exist. |

Detection is a syntax-only `go/parser` scan of non-test `.go` files in the output
package. It never type-checks the package.

Use `existing` when the type is part of a public API and must stay hand-written:

```go
// types.go
package order

type Status string
```

```yaml
enums:
  - name: Status
    type: string
    typeMode: existing
    values: [pending, paid, shipped]
```

With `auto`, the same declaration would be detected and reused automatically.
With `emit`, generation would fail because the type is declared twice.

### Output and pruning

- Per-enum output defaults to `<snake_case(enum_name)>_enum.go` in the manifest
  directory. Change it with `defaults.output.suffix`, `defaults.output.dir`, or a
  per-enum `output` (relative to the manifest, must end in `.go`).
- `aggregate: true` writes all enums of one manifest to a single file; use
  `output` to name it. Aggregate manifests require one package and forbid
  per-enum `output`.
- Generated files carry a manifest-specific header. When an enum is removed,
  `enumer generate` deletes its old generated file. Hand-written files, files
  from other manifests and custom target outputs are never pruned.
- `--no-prune` disables deletion; `enumer check` reports orphaned generated
  files as stale.

## Verify

Run these after generation:

```sh
enumer generate ./...
go build ./...
go test ./...
enumer check ./...
```

- `enumer check --diff ./...` prints unified diffs for stale outputs.
- `enumer --json generate ./...` emits structured diagnostics for tooling.
- `enumer --strict generate ./...` treats warnings as errors.
- Exit codes: `0` success, `1` validation/generation/stale failure, `2` command
  line usage error.

If you enabled `yaml`, the generated file imports `gopkg.in/yaml.v3`. Add it to
the module before building:

```sh
go get gopkg.in/yaml.v3
go mod tidy
```

## Common pitfalls

- **Editing generated files.** Change `enumer.yaml` and regenerate instead.
- **Declaring the type twice.** If you keep a hand-written type, use `auto` or
  `existing`; do not leave an existing declaration while using `emit`.
- **Putting the type in another package.** `auto`/`existing` only scan the
  output package. The hand-written type must be in the same package and
  directory as the generated file.
- **`json` + `json:number`, or `sql` + `sql:int`.** These pairs are mutually
  exclusive; choose one.
- **`sql:int` without `number`.** Every value of a string enum using `sql:int`
  must declare `number`.
- **`flags` on a signed type.** `flags` requires `uint`, `uint8`…`uint64`, and
  every value must be zero or a power of two, with at most one zero and no
  duplicates.
- **Assuming `sql.Scan` validates values.** `sql` scanning accepts undeclared
  values; call `IsValid` when validation matters. `sql:int` does reject unknown
  numbers.
- **Forgetting the YAML dependency.** `yaml` generates an import of
  `gopkg.in/yaml.v3`; run `go mod tidy`.
- **Expecting undeclared enums.** enumer never invents enums; add them to the
  manifest.
- **Using `@latest` in CI.** Pin a version or use a project tool dependency.

## References

Read only what the task needs:

- [Manifest reference](references/manifest.md) — every field, default, output
  rule, `extends` merge semantics and validation rule.
- [Feature reference](references/features.md) — generated API, dependencies,
  conflicts and behavioral details for all 16 features.
- [CLI reference](references/cli.md) — installation, commands, flags, paths,
  diagnostics, pruning, `go:generate`, CI and the programmatic API.
- [Recipes](references/recipes.md) — task-oriented examples for string, numeric,
  JSON, SQL, nullable, bitmask, YAML, existing-type, aggregate, extends, custom
  targets and migration from a comment DSL.

## Completion checklist

When the work is done, report:

- the manifest path and the enums/values/features changed;
- the generated file paths;
- the result of `go build ./...`, `go test ./...` and `enumer check ./...`;
- any new module dependency (for example `gopkg.in/yaml.v3`);
- whether the CLI was installed or added as a tool dependency;
- any warnings or diagnostics that remain.
