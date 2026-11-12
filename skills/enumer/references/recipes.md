# Recipes

Task-oriented examples. Each recipe starts from an `enumer.yaml` in the package
directory, then shows the generated API and the verification step. Pick the
smallest feature set that satisfies the requirement.

## Contents

- [Minimal string enum](#minimal-string-enum)
- [Numeric enum with explicit values](#numeric-enum-with-explicit-values)
- [JSON API](#json-api)
- [SQL column](#sql-column)
- [Nullable SQL column](#nullable-sql-column)
- [String enum stored as an integer](#string-enum-stored-as-an-integer)
- [CLI flag](#cli-flag)
- [Bit flags](#bit-flags)
- [YAML configuration](#yaml-configuration)
- [Keep a hand-written type](#keep-a-hand-written-type)
- [Aggregate into one file](#aggregate-into-one-file)
- [Shared defaults across packages](#shared-defaults-across-packages)
- [Comments, aliases and deprecation](#comments-aliases-and-deprecation)
- [Custom targets](#custom-targets)
- [Migrate from a comment DSL](#migrate-from-a-comment-dsl)
- [Verify the result](#verify-the-result)

## Minimal string enum

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

Generated with the default features:

```go
const (
    StatusPending Status = "pending"
    StatusPaid    Status = "paid"
    StatusShipped Status = "shipped"
)

status, err := ParseStatus("paid")
status.IsValid()
status.String()      // "paid"
StatusNames()        // []string{"pending", "paid", "shipped"}
StatusValues()       // []Status{StatusPending, StatusPaid, StatusShipped}
```

The default `text` feature also makes `encoding/json` round-trip the value as a
JSON string.

## Numeric enum with explicit values

```yaml
version: 1
package: order

enums:
  - name: Priority
    type: int32
    start: 10
    values:
      - low          # 10
      - normal       # 11
      - name: high
        value: 40    # 40
      - urgent       # 41
```

- Auto-increment starts at `start` (default `0`).
- An explicit `value` resets the sequence for following values.
- Numeric `String()` returns the value name; an undeclared number renders as
  `Priority(99)`.
- `IsValid` checks declared values.

## JSON API

### Names on the wire (recommended for readability)

```yaml
enums:
  - name: Status
    type: string
    features: [json]
    values: [pending, paid, shipped]
```

```json
{"status": "paid"}
```

`json` adds explicit `MarshalJSON`/`UnmarshalJSON`. The default `text` feature
already round-trips strings through `encoding/json`; add `json` when you want
explicit methods or explicit `null` handling. JSON `null` leaves the receiver
unchanged.

### Numbers on the wire

```yaml
enums:
  - name: Status
    type: int32
    features: [json:number]
    start: 100
    values: [pending, paid, shipped]
```

```json
{"status": 101}
```

- `json:number` is numeric-only and conflicts with `json`.
- Unmarshaling validates that the number is declared; with `flags`, it accepts
  any valid bit combination.

## SQL column

```yaml
enums:
  - name: Status
    type: string
    features: [sql]
    values: [pending, paid, shipped]
```

```go
var s Status
err := row.Scan(&s)   // s.Scan(src)
v, err := s.Value()   // string or int64 depending on the enum kind
```

- `Scan` accepts `nil`, `string`, `[]byte` (and integers for numeric enums).
- `Scan` does **not** reject undeclared values. Call `s.IsValid()` if the
  database may contain values outside the enum.
- For numeric enums, `Value` returns `int64`; `uint64` values above
  `math.MaxInt64` return an overflow error.

## Nullable SQL column

```yaml
enums:
  - name: Status
    type: int32
    features: [json:number, sql, sql:null]
    values: [pending, paid, shipped]
```

Generated wrapper:

```go
type NullStatus struct {
    Enum  Status
    Valid bool
}
```

- `Scan(nil)` sets `Valid = false`; `Value()` returns `nil`.
- JSON marshals as `null` when invalid, and unmarshals `null` back to invalid.
- `sql:null` requires exactly one of `sql` or `sql:int`.

## String enum stored as an integer

```yaml
enums:
  - name: Color
    type: string
    features: [sql:int]
    values:
      - name: red
        number: 1
      - name: green
        number: 2
      - name: blue
        number: 3
```

- Every value needs `number`.
- `Scan` maps the integer to a declared constant and rejects unknown numbers.
- `Value` maps the enum back to its number and errors on undeclared values.
- `sql:int` is string-only and conflicts with `sql`.

## CLI flag

```yaml
enums:
  - name: LogLevel
    type: string
    features: [flag]
    values: [debug, info, warn, error]
```

```go
var level LogLevel = LogLevelInfo
flag.Var(&level, "log-level", "log level")
```

`Set` delegates to `ParseLogLevel`; `Type` returns `"loglevel"`. The generated
methods also work with `pflag`.

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
      - name: execute
        value: 4
```

```go
p := PermissionRead | PermissionWrite
p.String()                     // "read|write"
p.Has(PermissionRead)          // true
p.With(PermissionExecute)      // read|write|execute
p.Clear(PermissionWrite)       // read
p.Toggle(PermissionRead)       // write
p.Flags()                      // []Permission{PermissionRead, PermissionWrite}

ParsePermission("read | execute") // spaces are trimmed
```

- Unsigned integer types only.
- Every value must be zero or a power of two; at most one zero; unique values.
- `IsValid` rejects undeclared bits; `String` renders unknown bits as
  `Permission(n)`.

## YAML configuration

```yaml
enums:
  - name: Environment
    type: string
    features: [yaml]
    values: [development, staging, production]
```

- Adds `MarshalYAML`/`UnmarshalYAML` using `gopkg.in/yaml.v3`.
- Run `go get gopkg.in/yaml.v3 && go mod tidy` after enabling it.
- Marshals the canonical text; unmarshals through `ParseEnvironment`.

## Keep a hand-written type

```go
// types.go
package order

// Status is part of the public API and is maintained by hand.
type Status string
```

```yaml
version: 1
package: order

enums:
  - name: Status
    type: string
    typeMode: existing
    values: [pending, paid, shipped]
```

- `existing` never emits the type and fails if no matching declaration exists.
- `auto` (the default) would detect the declaration and behave the same way.
- The declaration must be in a non-test `.go` file in the output package.
- Do not use `emit` while the hand-written declaration exists.

## Aggregate into one file

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

- Both enums are written to `http_enum.go`.
- All enums must share one package.
- Per-enum `output` is forbidden in aggregate manifests.
- If `output` is omitted, the default is `<manifest-stem>_enum.go`.

## Shared defaults across packages

```yaml
# base.yaml
version: 1
defaults:
  features: [string, parse, valid, text, names, values]
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

- Only `defaults` is inherited; the base's enums, targets, package and output are
  ignored.
- Nested `naming`, `parse` and `output` fields merge individually.
- `features` lists replace the inherited list; `features: []` clears it.
- Cycles are errors.
- Run `enumer generate ./...`; base files named `base.yaml` are not discovered as
  standalone manifests.

## Comments, aliases and deprecation

```yaml
version: 1
package: order

defaults:
  comments: auto
  parse:
    caseSensitive: false

enums:
  - name: Color
    comment: Color is the public color enumeration.
    comments: none
    values:
      - name: red
        comment: Red is the default color.
      - name: green
        aliases: [emerald]
      - name: blue
        deprecated: use navy
```

- `comment` wins over `doc`; both override generated comments.
- `comments: none` suppresses automatic comments but keeps user comments and
  deprecation notices.
- `caseSensitive: false` makes parsing and aliases case-insensitive. Texts that
  differ only by case are rejected.
- `aliases` parse but do not generate constants.
- `deprecated: true` emits `Deprecated: Do not use.`; a string is the reason.

## Custom targets

```yaml
targets:
  - template: ./templates/enums.md.tmpl
    output: ENUMS.md
  - template: ./templates/schema.json.tmpl
    output: "{{.Enum.Name}}.json"
    perEnum: true
```

`templates/enums.md.tmpl`:

```gotemplate
# Enums

{{- range .Enums }}
## {{ .Name }} ({{ .Type.Name }})

{{- range .Values }}
- `{{ .ConstName }}` — `{{ .Text }}`
{{- end }}
{{- end }}
```

- `.Manifest`, `.Enums` and `.Enum` are available; `.Enum` is non-nil only with
  `perEnum: true`.
- Functions: `camel`, `snake`, `kebab`, `lower`, `upper`, `title`, `quote`,
  `join`, `json`, `jsonPretty`, `default`.
- Outputs ending in `.go` are formatted and must be complete Go files.
- Custom target outputs are not pruned.

## Migrate from a comment DSL

For a type declared with a comment DSL such as `// ENUM(...)`:

1. Read the existing type declaration, values, aliases, comments and any
   `go:generate` directive.
2. Create `enumer.yaml` in the same package.
3. Move the type, values and settings into the manifest:

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

4. Remove the comment DSL block and the old generated file (for example
   `status_enum.go`) so its symbols do not collide.
5. Decide who owns the type:
   - Delete the hand-written type and let `typeMode: auto` (default) emit it.
   - Keep the type and set `typeMode: existing` (or leave `auto`).
6. Update or remove the old `//go:generate` directive. Prefer:

   ```go
   //go:generate go tool enumer generate
   ```

7. Generate, build and compare constants:

   ```sh
   enumer generate ./...
   go build ./...
   go test ./...
   ```

8. If constant names changed, adjust `naming.prefix`, `naming.identifier`,
   `extraInitialisms` or explicit value `name` fields. Check call sites before
   committing.

enumer does not read `// ENUM(...)` comments; the YAML manifest is the only
input.

## Verify the result

Run from the module or package root:

```sh
enumer generate ./...
go build ./...
go test ./...
enumer check ./...
```

- `go build` catches missing imports and type conflicts.
- `go test` catches behavioral changes in parsing, JSON, SQL or flags.
- `enumer check` catches stale generated files.
- If `yaml` is enabled and the build reports a missing package, run
  `go get gopkg.in/yaml.v3 && go mod tidy`.
- Commit the manifest and the generated files together; do not hand-edit the
  generated files.
