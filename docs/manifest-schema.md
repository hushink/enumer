# Manifest schema

**English** | [简体中文](manifest-schema.zh-CN.md)

An `enumer.yaml` file declares one or more enums and, optionally, additional
template targets. The manifest is the single source of truth: `enumer` does not
parse Go source files.

Every manifest is strictly decoded. Unknown fields, unsupported versions and
invalid values are reported with a file position and a stable diagnostic code.

## Top level

| Field | Type | Default | Description |
| --- | --- | --- | --- |
| `version` | integer | `1` | Manifest schema version. Only `1` is supported. |
| `extends` | string | `""` | Path to a base manifest whose `defaults` are inherited. Only `defaults` is inherited. |
| `package` | string | inferred | Go package name. Inferred from the directory's Go files when omitted; required when the directory has no Go files. |
| `aggregate` | bool | `false` | Emit all enums of this manifest into one file. |
| `output` | string | derived | Aggregate output path. Only valid when `aggregate` is true. Must be relative to the manifest directory. |
| `defaults` | mapping | — | Settings shared by every enum in this manifest. |
| `enums` | list | — | Enum declarations. At least one is required. |
| `targets` | list | — | Custom template outputs. |

## `defaults`

| Field | Type | Default | Description |
| --- | --- | --- | --- |
| `features` | list of strings | `[string, parse, valid, text, names, values]` | Base feature set. When omitted, the built-in default set is used; an explicit empty list disables all features. |
| `disableFeatures` | list of strings | `[]` | Features removed from the base set. Disabling `sql` also removes `sql:int` and `sql:null`; disabling `json` also removes `json:number`. |
| `naming.prefix` | string | `type` | `type` prefixes constants with the enum name; `none` uses no prefix; `literal:<text>` uses a fixed prefix. |
| `naming.identifier` | string | `camel` | `camel` converts value names to CamelCase; `raw` uses the name unchanged and requires a valid Go identifier. |
| `naming.extraInitialisms` | list of strings | `[]` | Additional all-uppercase words, for example `GRPC`. |
| `naming.disableInitialisms` | bool | `false` | Disable the built-in initialism list (`HTTP`, `URL`, `ID`, …). |
| `parse.caseSensitive` | bool | `true` | When false, parsing and alias matching are case-insensitive; value texts that differ only by case are rejected. |
| `output.suffix` | string | `_enum.go` | Suffix for per-enum output files. `.go` is appended when missing. |
| `output.dir` | string | `""` | Directory for per-enum output files, relative to the manifest directory. |
| `comments` | string | `auto` | `auto` adds documentation comments when no user doc is present; `none` suppresses automatic comments but keeps user docs, comment overrides and deprecation notices. |
| `allowDuplicateValues` | bool | `false` | Permit several names for the same numeric value. |
| `typeMode` | string | `auto` | `auto` detects an existing declaration and emits the type only when it is missing; `emit` always emits; `existing` never emits and requires a matching declaration. |

## `enums`

| Field | Type | Default | Description |
| --- | --- | --- | --- |
| `name` | string | required | Go type name. |
| `type` | string | `string` | Underlying type: `string`, any signed integer type or any unsigned integer type. |
| `doc` | string | `""` | Documentation comment for the type. |
| `comment` | string | `""` | Comment override; wins over `doc`. |
| `comments` | string | inherited | Per-enum `auto` or `none` comment mode. |
| `package` | string | manifest package | Per-enum package override. Requires an explicit `output` unless the manifest is aggregated. |
| `output` | string | derived | Per-enum output path, relative to the manifest directory. Must end with `.go`. |
| `features` | list of strings | `[]` | Features added to the resolved defaults. |
| `disableFeatures` | list of strings | `[]` | Features removed from the resolved defaults. |
| `typeMode` | string | inherited | Per-enum override of `defaults.typeMode`. |
| `start` | integer | `0` | Initial value for numeric auto-incrementing. Not valid for string enums. |
| `allowDuplicateValues` | bool | inherited | Per-enum override. |
| `values` | list | required | At least one value. |

## `values`

A value is either a scalar or a mapping:

```yaml
values:
  - red
  - name: green
    value: GREEN
    doc: A custom green.
    aliases: [g]
    deprecated: use teal
```

| Field | Type | Description |
| --- | --- | --- |
| `name` | string | Value name as written. Required in the mapping form. Scalar values set only the name. |
| `value` | string or integer | For string enums, the canonical text (defaults to `name`). For numeric enums, the integer value (defaults to auto-increment). |
| `number` | integer | Integer representation used by `sql:int` on string enums. Required for every value when `sql:int` is enabled. |
| `doc` | string | Documentation comment for the generated constant. |
| `comment` | string | Comment override; wins over `doc`. |
| `aliases` | list of strings | Additional parseable texts. Aliases do not generate constants. |
| `deprecated` | bool or string | Adds a Go `Deprecated:` comment. `true` uses "Do not use."; a string is used as the reason. |

Value names are converted to identifiers according to `naming`. `http_status`
becomes `HTTPStatus` with the default initialism list; `blue-green` becomes
`BlueGreen`.

For numeric enums, the canonical text is the value name exactly as written.

## `targets`

| Field | Type | Description |
| --- | --- | --- |
| `template` | string | Path to a `text/template` file, relative to the manifest directory. |
| `output` | string | Output path template, relative to the manifest directory. May reference template data. |
| `perEnum` | bool | Execute the template once per enum instead of once per manifest. |

The template context is:

```go
type TargetData struct {
	Manifest *model.Manifest
	Enums    []*model.EnumSpec
	Enum     *model.EnumSpec // non-nil only when perEnum is true
}
```

Template functions:

| Function | Usage |
| --- | --- |
| `camel`, `snake`, `kebab`, `lower`, `upper`, `title` | Text conversion. |
| `quote` | Go-quoted string literal. |
| `join` | `{{ join ", " .Names }}` or `{{ .Names \| join ", " }}`. |
| `json`, `jsonPretty` | JSON encoding. |
| `default` | `{{ default "none" .Enum.Doc }}`; returns the fallback for zero values. |

## Shared defaults with extends

`extends` points to another YAML manifest and inherits its `defaults`:

```yaml
# base.yaml
version: 1
defaults:
  features: [string, parse, valid, text]
  naming:
    prefix: type
    identifier: camel
```

```yaml
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

Merge rules:

- Only `defaults` is inherited. `package`, `aggregate`, `output`, `enums` and
  `targets` in the base are ignored.
- Nested `naming`, `parse` and `output` fields merge individually; unspecified
  child fields keep the inherited value.
- Lists such as `features` replace the inherited list. `features: []` clears it;
  omitted `features` inherits the base list.
- `extends` chains are resolved recursively and cycles are errors.
- Base files are not discovered by `./...` unless they are named `enumer.yaml`.

See [examples/extends](../examples/extends) for a runnable example.

## Type declarations

`typeMode` controls whether the generated file declares the enum type:

- `auto` (default): scan the target package. Emit the type when no declaration
  exists; reference an existing declaration when it does. Aliases, generic
  types and underlying-type mismatches are errors.
- `emit`: always emit the type. A declaration in another file is an error.
- `existing`: never emit. The declaration must exist and match.

```yaml
enums:
  - name: Color
    type: string
    values: [red, green, blue]
```

```go
// types.go
package example

type Color string
```

With `auto`, enumer detects the declaration and emits only the constants and
methods. If the file is absent, it emits the type from the manifest. Detection
uses a syntax-only `go/parser` scan; enumer does not load or type-check
packages.

See [examples/existing](../examples/existing) for a hand-written declaration.

## Comments

- `doc` and `comment` override generated comments; `comment` wins.
- `comments: auto|none` controls automatic comments globally and per enum.
- User comments are emitted even with `comments: none`.

## Formatting

Generated Go files and custom targets whose output path ends with `.go` are
formatted with the standard library's `go/format.Source`. A custom target that
outputs `.go` must render a complete Go file.

## Features

| Feature | Generated API | Requires |
| --- | --- | --- |
| `string` | `String() string` | — |
| `parse` | `ErrInvalid<Name>`, `Parse<Name>` and the value map | — |
| `valid` | `IsValid() bool` | `parse` |
| `flags` | Flag helpers and combination-aware `String`/`Parse`/`IsValid` | `string`, `parse`, `valid` |
| `mustparse` | `MustParse<Name>` | `parse` |
| `text` | `MarshalText` / `UnmarshalText` | `parse`, `string` |
| `json` | Name-based `MarshalJSON` / `UnmarshalJSON` | `parse`, `string` |
| `json:number` | Numeric `MarshalJSON` / `UnmarshalJSON` | `parse` |
| `sql` | `Scan` / `Value` | `parse`, `string` |
| `sql:int` | `sql` semantics using integer storage | `parse`, `string` |
| `sql:null` | `Null<Name>` wrapper | exactly one of `sql`, `sql:int` |
| `flag` | `Set` / `Type` | `parse`, `string` |
| `names` | `<Name>Names() []string` | — |
| `values` | `<Name>Values() []<Name>` | — |
| `ptr` | `Ptr() *<Name>` | — |

`json` and `json:number` are mutually exclusive. `sql` and `sql:int` are
mutually exclusive. `json:number` is only valid for numeric enums; `sql:int` is
only valid for string enums.

`sql:null` generates:

```go
type Null<Name> struct {
	Enum  <Name>
	Valid bool
}
```

The wrapper implements `sql.Scanner` and `driver.Valuer`. A `nil` database value
sets `Valid` to false and `Value()` writes `nil` back.

## Bit flags

`flags` is only valid for unsigned integer enums. Every value must be zero or a
power of two, at most one value may be zero, and values must be unique.

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

Generated helpers:

- `Has(flag) bool`
- `With(flag)`, `Clear(flag)`, `Toggle(flag)`
- `Flags() []<Name>`
- `String()` joins set flag names with `|`; unknown bits are rendered as
  `Name(n)`.
- `Parse` accepts a single name or a `|`-separated combination.
- `IsValid` reports whether only declared bits are set.

`flags` adds `string`, `parse` and `valid` unless they are explicitly disabled.
See [examples/flags](../examples/flags) for a runnable example.

## Output resolution

- Per-enum output defaults to `<dir>/<snake_case(enum_name)><suffix>`.
- An explicit `output` is resolved relative to the manifest directory and must
  end with `.go`.
- Aggregate output defaults to `<dir>/<manifest-stem><suffix>`.
- Output paths are compared case-insensitively to avoid cross-platform
  collisions.
- Generated symbols are checked per package: type names, constant names,
  `Parse…`, `ErrInvalid…`, `Null…`, `…Names`, `…Values` and private generated
  identifiers must not collide.

## Validation rules

- Enum names must be valid Go identifiers.
- Value constant names must be valid Go identifiers after naming conversion.
- Duplicate constant names, duplicate canonical texts, duplicate aliases and
  alias/text collisions are errors.
- Duplicate numeric values are errors unless `allowDuplicateValues` is enabled.
  When enabled, `String()` returns the first declared name and `Parse` accepts
  all names.
- Numeric values must fit the underlying type. Auto-increment overflow is an
  error.
- Empty `values` lists and manifests without enums are errors.
- `int` and `uint` are treated as 64-bit for validation. Values used on 32-bit
  targets must fit the target platform's type.
- String enums cannot use `start`.
- Numeric enums cannot use `number`; use `value` instead.
- `sql:int` requires a `number` on every value.
- `sql` and `sql:int` scanning does not validate that the scanned value is
  declared; `IsValid` is the explicit validation API. `sql:int` must map an
  integer to a declared value and therefore rejects unknown numbers.

## Diagnostics

Human output:

```
path/to/enumer.yaml:12:5: enumer: duplicate text "red" in enum Color (first declared at path/to/enumer.yaml:8:3)
```

`--json` output:

```json
{
  "version": 1,
  "diagnostics": [
    {
      "severity": "error",
      "code": "duplicate-text",
      "message": "duplicate text \"red\" in enum Color (first declared at ...)",
      "file": "path/to/enumer.yaml",
      "line": 12,
      "column": 5,
      "enum": "Color",
      "value": "red"
    }
  ]
}
```

Exit codes: `0` success, `1` validation/generation/stale check failure, `2`
command line usage error.

## Example: aggregate output

```yaml
version: 1
package: http

aggregate: true
output: http_enum.go

enums:
  - name: Method
    type: string
    values: [GET, POST, PUT, DELETE]
  - name: Status
    type: int
    start: 100
    values: [continue, ok, created, accepted, no_content]
```

`enumer generate` writes a single `http_enum.go` containing both enums.
