# Manifest reference

`enumer.yaml` is the single source of truth. It declares the enums, their values,
the features to generate, and optional custom template outputs. Unknown fields,
unsupported versions and invalid values are errors with a file position and a
stable diagnostic code.

## Contents

- [File names and discovery](#file-names-and-discovery)
- [Top level](#top-level)
- [Defaults](#defaults)
- [Enums](#enums)
- [Values](#values)
- [Types and value semantics](#types-and-value-semantics)
- [Naming](#naming)
- [Comments and deprecation](#comments-and-deprecation)
- [Type declarations (`typeMode`)](#type-declarations-typemode)
- [Output resolution](#output-resolution)
- [Shared defaults with `extends`](#shared-defaults-with-extends)
- [Custom targets](#custom-targets)
- [Validation rules](#validation-rules)
- [Diagnostics](#diagnostics)
- [Complete example](#complete-example)

## File names and discovery

- A manifest may be named `enumer.yaml` or `enumer.yml`.
- `enumer generate ./...` discovers manifests recursively. It skips `vendor/`,
  `testdata/`, hidden directories and directories whose names start with `_`.
- A directory must not contain both `enumer.yaml` and `enumer.yml`.
- An explicit path to a file is accepted even when the file has another name.
- `extends` base files are not discovered by `./...` unless they are themselves
  named `enumer.yaml` or `enumer.yml`.

## Top level

| Field | Type | Default | Description |
| --- | --- | --- | --- |
| `version` | integer | `1` | Schema version. Omitted means `1`; only `1` is supported. |
| `extends` | string | `""` | Path to another manifest whose `defaults` are inherited. |
| `package` | string | inferred | Go package name. Inferred from Go files in the manifest directory; set it when the directory has no Go files. |
| `aggregate` | bool | `false` | Write every enum of this manifest into one file. |
| `output` | string | derived | Aggregate output path, relative to the manifest directory. Only valid with `aggregate: true`. |
| `defaults` | mapping | — | Shared settings applied to every enum in the manifest. |
| `enums` | list | — | Enum declarations. At least one is required. |
| `targets` | list | — | Custom `text/template` outputs. |

## Defaults

| Field | Type | Default | Description |
| --- | --- | --- | --- |
| `features` | list of strings | `[string, parse, valid, text, names, values]` | Base feature set. Omit to use the built-in default; `features: []` disables all features. |
| `disableFeatures` | list of strings | `[]` | Features removed from the base set. Disabling `sql` also removes `sql:int` and `sql:null`; disabling `json` also removes `json:number`. |
| `naming.prefix` | string | `type` | Constant prefix: `type` (enum name), `none`, or `literal:<text>`. |
| `naming.identifier` | string | `camel` | `camel` converts value names to CamelCase; `raw` keeps them unchanged and requires valid Go identifiers. |
| `naming.extraInitialisms` | list of strings | `[]` | Additional words to render in all caps, for example `GRPC`. |
| `naming.disableInitialisms` | bool | `false` | Disable the built-in initialism list (`HTTP`, `URL`, `ID`, …). |
| `parse.caseSensitive` | bool | `true` | When false, parsing and alias matching ignore case; value texts that differ only by case are rejected. |
| `output.suffix` | string | `_enum.go` | Suffix for per-enum output files. `.go` is appended when missing. |
| `output.dir` | string | `""` | Directory for per-enum output files, relative to the manifest directory. |
| `comments` | string | `auto` | `auto` adds generated documentation comments; `none` suppresses automatic comments but keeps user docs and deprecation notices. |
| `allowDuplicateValues` | bool | `false` | Permit several names for the same numeric value. |
| `typeMode` | string | `auto` | `auto`, `emit` or `existing`. See [Type declarations](#type-declarations-typemode). |

## Enums

| Field | Type | Default | Description |
| --- | --- | --- | --- |
| `name` | string | required | Go type name. Must be a valid Go identifier. |
| `type` | string | `string` | Underlying type: `string`, any signed integer type, or any unsigned integer type. |
| `doc` | string | `""` | Documentation comment for the type. |
| `comment` | string | `""` | Comment override; wins over `doc`. |
| `comments` | string | inherited | Per-enum `auto` or `none` comment mode. |
| `package` | string | manifest package | Per-enum package override. Requires an explicit `output` unless the manifest is aggregated. |
| `output` | string | derived | Per-enum output path, relative to the manifest directory. Must end in `.go`. |
| `features` | list of strings | `[]` | Features added to the resolved defaults. |
| `disableFeatures` | list of strings | `[]` | Features removed from the resolved defaults. |
| `typeMode` | string | inherited | Per-enum override of `defaults.typeMode`. |
| `start` | integer | `0` | Initial value for numeric auto-incrementing. Not valid for string enums. |
| `allowDuplicateValues` | bool | inherited | Per-enum override. |
| `values` | list | required | At least one value. |

## Values

A value is either a scalar or a mapping:

```yaml
values:
  - red
  - name: green
    value: GREEN
    doc: A custom green.
    comment: A comment override.
    aliases: [g]
    deprecated: use teal
```

| Field | Type | Description |
| --- | --- | --- |
| `name` | string | Value name as written. Required in mapping form; a scalar sets only the name. |
| `value` | string or integer | String enums: canonical text (defaults to `name`). Numeric enums: the integer value (defaults to auto-increment). |
| `number` | integer | Integer storage used by `sql:int` on string enums. Required for every value when `sql:int` is enabled. |
| `doc` | string | Documentation comment for the generated constant. |
| `comment` | string | Comment override; wins over `doc`. |
| `aliases` | list of strings | Additional parseable texts. Aliases do not generate constants. |
| `deprecated` | bool or string | Adds a Go `Deprecated:` comment. `true` uses "Do not use."; a string is the reason. |

## Types and value semantics

| Type | Kind |
| --- | --- |
| `string` | string |
| `int`, `int8`, `int16`, `int32`, `int64` | signed integer |
| `uint`, `uint8`, `uint16`, `uint32`, `uint64` | unsigned integer |

- `type` defaults to `string`.
- String enums use `value` as the canonical text. The constant name comes from
  `name`, not `value`.
- Numeric enums use `value` as the number. Values without an explicit `value`
  auto-increment from `start`; an explicit value resets the sequence for the
  values that follow.
- Numeric values are range-checked against the underlying type. `int` and `uint`
  are validated as 64-bit; values used on 32-bit targets must fit the target
  platform's type.
- `start` is only valid for numeric enums.
- `number` is only valid for string enums using `sql:int`; numeric enums must use
  `value` instead.
- Duplicate numeric values are errors unless `allowDuplicateValues` is enabled.
  When enabled, `String()` returns the first declared name and `Parse` accepts
  all names.
- Duplicate constant names, duplicate canonical texts, duplicate aliases and
  alias/text collisions are always errors.
- `aliases` are extra parseable texts. They never generate constants.
- `deprecated: true` emits `Deprecated: Do not use.`; a string emits that string
  as the deprecation reason.

## Naming

Value names are converted to Go constant names using `defaults.naming`:

```yaml
defaults:
  naming:
    prefix: type
    identifier: camel
    extraInitialisms: [GRPC]
    disableInitialisms: false
```

- `prefix: type` prefixes constants with the enum name: `ColorRed`.
- `prefix: none` uses the bare identifier: `Red`. This is only safe when the
  package has no collisions.
- `prefix: literal:<text>` uses a fixed prefix: `<text>Red`.
- `identifier: camel` converts `http_status` to `HTTPStatus`, `blue-green` to
  `BlueGreen`, and `red_orange` to `RedOrange`. The built-in initialism list
  includes `HTTP`, `URL`, `ID` and others; `extraInitialisms` adds words such as
  `GRPC`, and `disableInitialisms: true` turns the built-in list off.
- `identifier: raw` keeps the name unchanged and requires it to already be a
  valid Go identifier.

With the defaults, `http_status` with enum `Color` becomes `ColorHTTPStatus`.

## Comments and deprecation

- `doc` and `comment` override generated comments for the enum or value.
  `comment` wins when both are set.
- `comments: auto` generates documentation when no user comment is present.
- `comments: none` suppresses automatic comments but keeps user docs, comment
  overrides and `Deprecated:` notices.
- `comments` can be set in `defaults` and overridden per enum.
- `deprecated` accepts `true` or a reason string.

## Type declarations (`typeMode`)

| Mode | Behavior |
| --- | --- |
| `auto` (default) | Scan the output package. Emit the type when no declaration exists; reuse an existing declaration; error on aliases, generics or underlying-type mismatch. |
| `emit` | Always emit the type. A declaration in another file is an error. |
| `existing` | Never emit. A matching declaration must exist. |

```yaml
enums:
  - name: Color
    type: string
    typeMode: existing
    values: [red, green, blue]
```

```go
// types.go
package example

type Color string
```

Detection uses a syntax-only `go/parser` scan of non-test `.go` files in the
output package. It does not load or type-check packages. `_test.go` files are
ignored, so the declaration must live in a regular Go file.

## Output resolution

- Per-enum output defaults to `<snake_case(enum_name)><suffix>` in the manifest
  directory. The default suffix is `_enum.go`.
- `defaults.output.dir` changes the directory; `defaults.output.suffix` changes
  the suffix. `.go` is appended to a suffix that lacks it.
- A per-enum `output` is resolved relative to the manifest directory and must
  end in `.go`.
- Aggregate output defaults to `<dir>/<manifest-stem><suffix>`. Set `output` to
  choose a different path.
- Aggregate manifests require one package across all enums and forbid per-enum
  `output`.
- Output paths are compared case-insensitively to avoid cross-platform
  collisions.
- Generated symbols are checked per package: type names, constant names,
  `Parse…`, `ErrInvalid…`, `Null…`, `…Names`, `…Values` and private generated
  identifiers must not collide.

Generated files carry the manifest's exact header. On `generate`, files with that
header that are no longer planned are removed (pruned). Hand-written files, files
from other manifests and custom target outputs are never pruned. Use
`--no-prune` to keep stale files; `enumer check` reports them as stale.

## Shared defaults with `extends`

`extends` inherits **only** `defaults` from another manifest:

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

Merge rules:

- Only `defaults` is inherited. The base's `package`, `aggregate`, `output`,
  `enums` and `targets` are ignored.
- Nested `naming`, `parse` and `output` fields merge individually; unspecified
  child fields keep the inherited value.
- Lists such as `features` replace the inherited list. `features: []` clears it;
  omitting `features` inherits the base list.
- `extends` chains are resolved recursively; cycles are errors.
- Base files are not discovered by `./...` unless they are named `enumer.yaml`
  or `enumer.yml`.

## Custom targets

```yaml
targets:
  - template: ./templates/enums.md.tmpl
    output: ENUMS.md
  - template: ./templates/schema.json.tmpl
    output: "{{.Enum.Name}}.json"
    perEnum: true
```

| Field | Type | Description |
| --- | --- | --- |
| `template` | string | Path to a `text/template` file, relative to the manifest directory. |
| `output` | string | Output path template, relative to the manifest directory. May reference template data. |
| `perEnum` | bool | Execute once per enum instead of once per manifest. |

Template data:

```go
type TargetData struct {
    Manifest *model.Manifest
    Enums    []*model.EnumSpec
    Enum     *model.EnumSpec // non-nil only when perEnum is true
}
```

Template functions: `camel`, `snake`, `kebab`, `lower`, `upper`, `title`,
`quote`, `join`, `json`, `jsonPretty`, `default`.

Targets whose output path ends in `.go` are formatted with `go/format.Source`
and must render a complete Go file. Custom target outputs are not pruned.

## Validation rules

- Enum names must be valid Go identifiers.
- Value constant names must be valid Go identifiers after naming conversion.
- Duplicate constant names, canonical texts, aliases and alias/text collisions
  are errors.
- Duplicate numeric values are errors unless `allowDuplicateValues` is enabled.
- Numeric values must fit the underlying type; auto-increment overflow is an
  error.
- Empty `values` lists and manifests without enums are errors.
- `int` and `uint` are validated as 64-bit; values used on 32-bit targets must
  fit the target platform's type.
- String enums cannot use `start`.
- Numeric enums cannot use `number`; use `value`.
- `sql:int` requires `number` on every value.
- `sql` and `sql:int` scanning does not validate that the scanned value is
  declared; `IsValid` is the explicit validation API. `sql:int` rejects unknown
  numbers because it must map an integer to a declared value.
- `json` and `json:number` are mutually exclusive.
- `sql` and `sql:int` are mutually exclusive.
- `json:number` is only valid for numeric enums.
- `sql:int` is only valid for string enums.
- `sql:null` requires exactly one of `sql` or `sql:int`.
- `flags` requires an unsigned integer type; every value must be zero or a power
  of two, at most one value may be zero, and values must be unique.

## Diagnostics

Human-readable output includes the path and position:

```text
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

- `--strict` turns warnings into errors.
- Exit codes: `0` success, `1` validation/generation/stale check failure, `2`
  command line usage error.

## Complete example

```yaml
version: 1
package: order

defaults:
  features: [string, parse, valid, text, names, values]
  naming:
    prefix: type
    identifier: camel
  output:
    suffix: _enum.go

enums:
  - name: Status
    type: string
    doc: Status is the lifecycle state of an order.
    features: [json, flag]
    values:
      - pending
      - name: paid
        aliases: [settled]
      - name: shipped
        deprecated: use delivered
      - delivered

  - name: Priority
    type: int32
    start: 10
    features: [json:number, sql, sql:null]
    values:
      - low
      - normal
      - name: high
        value: 40
      - urgent
```
