# Configuration guide

**English** | [简体中文](configuration.zh-CN.md)

This guide walks through the common ways to configure `enumer`. The exhaustive
field-by-field reference is [manifest-schema.md](manifest-schema.md).

## Minimal manifest

```yaml
version: 1
package: example

enums:
  - name: Color
    type: string
    values:
      - red
      - green
      - blue
```

`enumer generate` writes `color_enum.go` next to the manifest. The generated
file contains the `Color` type, `ColorRed`/`ColorGreen`/`ColorBlue`, String,
Parse, IsValid, text marshaling, names and values.

If the directory contains Go files, `package` can be omitted; enumer infers it
from the package clause. When in doubt, set it explicitly.

## Where manifests live

- Any directory can contain an `enumer.yaml`.
- `enumer generate ./...` discovers `enumer.yaml` and `enumer.yml` recursively,
  skipping `vendor/`, `testdata/`, hidden directories and `_`-prefixed
  directories.
- An explicit file path is accepted even when it is not named `enumer.yaml`.

## Enums and values

```yaml
enums:
  - name: Status
    type: int32
    start: 100
    values:
      - pending
      - name: running
        value: 110
        comment: Running is the active state.
      - name: failed
        aliases: [error]
        deprecated: use failed_permanently
```

- `type` is `string`, a signed integer type or an unsigned integer type.
- String enums use `value` as the canonical text; it defaults to the value name.
- Numeric enums use `value` as the number. Without it, values auto-increment
  from `start` (default `0`); an explicit value resets the sequence.
- `aliases` are additional parseable texts and do not generate constants.
- `deprecated` adds a Go `Deprecated:` comment.
- `comment` overrides the generated comment and wins over `doc`.

Value names are converted to Go identifiers:

```yaml
values:
  - http_status   # ColorHTTPStatus
  - blue-green    # ColorBlueGreen
  - red_orange    # ColorRedOrange
```

## Features

The default feature set is:

```yaml
defaults:
  features: [string, parse, valid, text, names, values]
```

Features are additive per enum:

```yaml
defaults:
  features: [string, parse, valid, text, names, values]

enums:
  - name: Color
    features: [json, sql]        # added to the defaults
    disableFeatures: [names]     # removed from the defaults
```

Available features: `string`, `parse`, `valid`, `flags`, `mustparse`, `text`,
`json`, `json:number`, `yaml`, `sql`, `sql:int`, `sql:null`, `flag`, `names`,
`values`, `ptr`. Dependencies are resolved automatically. See
[manifest-schema.md](manifest-schema.md#features) for the full table.

An explicit empty `features: []` disables all features:

```yaml
defaults:
  features: []
```

## Output layout

Per-enum output defaults to `<snake(name)>_enum.go` in the manifest directory:

```yaml
defaults:
  output:
    suffix: _enum.go
    dir: generated

enums:
  - name: Color
    output: custom/color.go  # exact path, relative to the manifest
```

Aggregate output combines every enum in one manifest:

```yaml
aggregate: true
output: enums_enum.go
```

Aggregate manifests require one package across all enums and forbid per-enum
`output`.

## Pruning removed enums

Removing an enum from the manifest removes its old generated `.go` file on the
next `enumer generate`. Only files carrying the manifest's exact generated
header are pruned; hand-written files, files from other manifests and custom
target outputs are never touched. `enumer check` reports orphaned generated
files as stale, and `--no-prune` disables pruning.

## Comments

```yaml
defaults:
  comments: auto              # auto | none

enums:
  - name: Color
    comment: Color is the public color enumeration.
    comments: none            # per-enum override
    values:
      - name: red
        comment: Red is the default color.
```

- `auto` generates comments when no user comment is present.
- `none` suppresses automatic comments; user comments and deprecation notices
  remain.
- `doc` and `comment` are equivalent comment overrides; `comment` wins.

## Type declarations

`typeMode` controls the generated type declaration:

| Mode | Behavior |
| --- | --- |
| `auto` (default) | Emit the type when missing; use an existing declaration when present; error on aliases, generics or underlying-type mismatch. |
| `emit` | Always emit; an existing declaration elsewhere is an error. |
| `existing` | Never emit; the declaration must exist and match. |

```yaml
# types.go
package example

// Color is a hand-written type.
type Color string
```

With the default `auto`, the manifest below generates only constants and
methods, because `Color` already exists:

```yaml
enums:
  - name: Color
    type: string
    values: [red, green, blue]
```

Detection uses a syntax-only `go/parser` scan of non-test `.go` files in the
output package. It does not load or type-check packages.

## Shared defaults with extends

```yaml
# base.yaml
version: 1
defaults:
  features: [string, parse, valid, text, names, values]
  naming:
    prefix: type

# color/enumer.yaml
version: 1
package: color
extends: ../base.yaml

defaults:
  naming:
    identifier: camel
enums:
  - name: Color
    values: [red, green, blue]
```

Only `defaults` is inherited. Nested `naming`, `parse` and `output` fields merge
individually; lists replace the inherited list. Cycles are errors. Base files
are not discovered by `./...` unless named `enumer.yaml`.

## Custom targets

```yaml
targets:
  - template: ./templates/enums.md.tmpl
    output: ENUMS.md
  - template: ./templates/schema.go.tmpl
    output: schema_generated.go
```

Template data:

- `.Manifest`
- `.Enums`
- `.Enum` (when `perEnum: true`)

Functions: `camel`, `snake`, `kebab`, `lower`, `upper`, `title`, `quote`,
`join`, `json`, `jsonPretty`, `default`.

Targets whose output path ends with `.go` are formatted with
`go/format.Source`; they must render a complete Go file.

## Bit flags

```yaml
enums:
  - name: Permission
    type: uint32
    features: [flags]
    values:
      - name: none
        value: 0
      - name: read
        value: 1
      - name: write
        value: 2
```

`flags` requires an unsigned integer type. Values must be zero or powers of two,
unique, and at most one zero value is allowed. `String`, `Parse` and `IsValid`
handle combinations.

## Diagnostics

```sh
enumer generate ./...             # text diagnostics, exit 1 on errors
enumer --json generate ./...      # JSON diagnostics
enumer --strict generate ./...    # warnings become errors
enumer check --diff ./...         # fail when generated files are stale
```

Exit codes: `0` success, `1` validation/generation/stale, `2` usage error.

## Formatting and idempotency

- Built-in Go output uses `go/format.Source`.
- `.go` custom targets are formatted too.
- Writes are atomic and files with identical content are not rewritten.
- `enumer check` compares rendered content with the files on disk.

## Compliance

The library is MIT licensed. CI checks SPDX headers, dependency licenses and
known vulnerabilities. See the repository README for the exact commands.
