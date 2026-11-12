# Feature reference

Features are independently switchable code-generation capabilities. The manifest
selects a base set in `defaults.features`; each enum can add features with
`features` and remove them with `disableFeatures`.

## Contents

- [Resolution rules](#resolution-rules)
- [Feature summary](#feature-summary)
- [Core features](#core-features)
- [JSON and text](#json-and-text)
- [YAML](#yaml)
- [SQL](#sql)
- [Flags](#flags)
- [Lists and pointers](#lists-and-pointers)
- [Imports and module dependencies](#imports-and-module-dependencies)
- [Choosing a feature set](#choosing-a-feature-set)

## Resolution rules

- The built-in default set is `[string, parse, valid, text, names, values]`.
- `features` adds to the resolved set; `disableFeatures` removes from it.
- Dependencies are expanded automatically. For example, `json` pulls in `parse`
  and `string`.
- If a required feature is explicitly disabled, generation fails with
  `feature "<name>" is required but disabled`.
- Disabling a parent also disables its variants: `disableFeatures: [sql]`
  removes `sql`, `sql:int` and `sql:null`; `disableFeatures: [json]` removes
  `json` and `json:number`.
- An explicit `features: []` in `defaults` disables every feature. A per-enum
  `features: [json]` then adds `json` plus its dependencies, so the enum still
  gets `string` and `parse`.
- Feature names are case-sensitive and use the exact names below.
- `json`/`json:number` and `sql`/`sql:int` are mutually exclusive.
- Applicability is enforced: `json:number` is numeric-only, `sql:int` is
  string-only, and `flags` is unsigned-only.

## Feature summary

| Feature | Generated API | Requires | Notes |
| --- | --- | --- | --- |
| `string` | `String() string` | — | Skipped when `flags` is enabled; `flags` emits its own `String`. |
| `parse` | `ErrInvalid<Name>`, `Parse<Name>(string) (<Name>, error)`, value map | — | Flags-aware; accepts aliases. |
| `valid` | `IsValid() bool` | `parse` | Flags-aware mask check. |
| `mustparse` | `MustParse<Name>(string) <Name>` | `parse` | Panics on error. |
| `text` | `MarshalText`, `UnmarshalText` | `parse`, `string` | Also enables default `encoding/json` text round-tripping. |
| `json` | `MarshalJSON`, `UnmarshalJSON` | `parse`, `string` | Name/text-based; conflicts with `json:number`. |
| `json:number` | `MarshalJSON`, `UnmarshalJSON` | `parse` | Numeric enums only; conflicts with `json`. |
| `yaml` | `MarshalYAML`, `UnmarshalYAML` | `parse`, `string` | Imports `gopkg.in/yaml.v3`. |
| `sql` | `Scan(any) error`, `Value() (driver.Value, error)` | `parse`, `string` | Supports string and numeric enums. |
| `sql:int` | `Scan`, `Value`, integer mapping | `parse`, `string` | String enums only; every value needs `number`; conflicts with `sql`. |
| `sql:null` | `Null<Name>{Enum, Valid}` with `Scan`, `Value`, JSON | exactly one of `sql` or `sql:int` | Nullable wrapper. |
| `flag` | `Set(string) error`, `Type() string` | `parse`, `string` | Compatible with `flag` and `pflag`. |
| `flags` | `Has`, `With`, `Clear`, `Toggle`, `Flags`, combination-aware `String`/`Parse`/`IsValid` | `string`, `parse`, `valid` | Unsigned enums only; values zero or powers of two. |
| `names` | `<Name>Names() []string` | — | Canonical texts in declaration order. |
| `values` | `<Name>Values() []<Name>` | — | Constants in declaration order. |
| `ptr` | `Ptr() *<Name>` | — | Returns a pointer to the receiver. |

## Core features

### `string`

```go
func (x Color) String() string
```

- String enums return the underlying text.
- Numeric enums switch over declared values and return the canonical name.
- Undeclared numeric values render as `Color(42)`.
- When `flags` is enabled, this feature emits nothing; `flags` provides a
  combination-aware `String`.

### `parse`

```go
var ErrInvalidColor = errors.New("invalid Color")

func ParseColor(name string) (Color, error)
```

- The parser accepts canonical texts and aliases.
- `parse.caseSensitive: false` lowercases both the lookup table and the input.
  Value texts that differ only by case are rejected at validation time.
- Unknown input returns an error wrapping `ErrInvalidColor`.
- With `flags`, `Parse` also accepts `|`-separated combinations, trims spaces
  around each part, and ORs the bits.
- The private `map[string]Color` is an implementation detail; do not depend on
  its name.

### `valid`

```go
func (x Color) IsValid() bool
```

- For string enums, checks membership in the parse map (including aliases).
- For numeric enums, checks the declared constants.
- With `flags`, checks that no undeclared bits are set:
  `x &^ (Read | Write) == 0`.

### `mustparse`

```go
func MustParseColor(name string) Color
```

Panics with the parse error. Use it for package-level initialization and tests,
not for untrusted input.

### `text`

```go
func (x Color) MarshalText() ([]byte, error)
func (x *Color) UnmarshalText(text []byte) error
```

`encoding/json` uses these methods by default, so the default feature set already
round-trips enums through JSON as strings. Add `json` when you want explicit
`MarshalJSON`/`UnmarshalJSON` methods or explicit `null` handling.

## JSON and text

### `json`

```go
func (x Color) MarshalJSON() ([]byte, error)
func (x *Color) UnmarshalJSON(data []byte) error
```

- Marshals `x.String()` as a JSON string.
- Unmarshals a JSON string through `Parse<Name>`.
- JSON `null` leaves the receiver unchanged and returns no error.
- Requires `parse` and `string`; conflicts with `json:number`.
- `json` is name-based for numeric enums too: `StatusRunning` marshals as
  `"running"`, not `1`.

### `json:number`

```go
func (x Status) MarshalJSON() ([]byte, error)
func (x *Status) UnmarshalJSON(data []byte) error
```

- Marshals the underlying integer as a JSON number.
- Unmarshals a JSON number and validates it against declared values. With
  `flags`, it uses `IsValid` so combinations are accepted.
- JSON `null` leaves the receiver unchanged and returns no error.
- Numeric enums only; conflicts with `json`.
- Requires `parse`.

### Which one to use

| Wire format | Feature |
| --- | --- |
| JSON string with the canonical name | `json` (or rely on the default `text`) |
| JSON number with the underlying value | `json:number` |
| Both, depending on context | Not supported on one enum; use separate DTO types or custom code |

## YAML

```go
func (x Color) MarshalYAML() (any, error)
func (x *Color) UnmarshalYAML(value *yaml.Node) error
```

- Marshals `x.String()`.
- Unmarshals through `Parse<Name>`.
- Requires `parse` and `string`.
- Adds an import of `gopkg.in/yaml.v3`. After enabling it, run:

```sh
go get gopkg.in/yaml.v3
go mod tidy
```

## SQL

### `sql`

```go
func (x *Color) Scan(src any) error
func (x Color) Value() (driver.Value, error)
```

- String enums scan `string` and `[]byte` directly into the enum type.
- Numeric enums scan `int64`, `string` and `[]byte`, with range checks for
  narrower underlying types.
- `nil` sets the zero value.
- `Value` returns a string for string enums and an `int64` for numeric enums.
  `uint64` values above `math.MaxInt64` return an overflow error.
- **`Scan` does not validate that the value is declared.** Use `IsValid` after
  scanning when that matters.
- Requires `parse` and `string`. Imports `database/sql/driver` and `fmt`.

### `sql:int`

Stores a **string** enum as an integer column.

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

- Every value must declare `number`.
- `Scan` accepts `int64`, `string` and `[]byte` integers, maps the number to a
  declared constant, and **rejects unknown numbers**.
- `Value` maps the enum back to its declared number; undeclared enum values
  return an error.
- String enums only; conflicts with `sql`.
- Requires `parse` and `string`.

### `sql:null`

```go
type NullColor struct {
    Enum  Color
    Valid bool
}
```

- `Scan(nil)` sets `Valid` to false; `Value()` writes `nil`.
- `Scan(non-nil)` delegates to `Color.Scan` and sets `Valid` to true.
- `MarshalJSON` writes `null` when invalid.
- `UnmarshalJSON` accepts `null` and sets `Valid` to false.
- Requires exactly one of `sql` or `sql:int`.

Typical manifest:

```yaml
enums:
  - name: Status
    type: int32
    features: [json:number, sql, sql:null]
    values: [pending, running, failed]
```

## Flags

### `flag`

```go
func (x *Color) Set(s string) error
func (x Color) Type() string
```

Implements `flag.Value` and `pflag.Value`. `Set` delegates to `Parse<Name>`.
`Type` returns the lowercased enum name (for example `"color"`).

### `flags`

Bitmask helpers for **unsigned integer** enums:

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

Generated API:

```go
func (x Permission) String() string              // "read|write", "none", "Permission(8)"
func (x Permission) Has(flag Permission) bool    // all bits set
func (x Permission) With(flag Permission) Permission
func (x Permission) Clear(flag Permission) Permission
func (x Permission) Toggle(flag Permission) Permission
func (x Permission) Flags() []Permission
func ParsePermission(s string) (Permission, error) // "read|write"
func (x Permission) IsValid() bool                 // only declared bits
```

Rules:

- Underlying type must be unsigned (`uint`, `uint8`…`uint64`).
- Every value must be zero or a power of two.
- At most one zero value is allowed.
- Values must be unique.
- `String` returns the declared zero value's text when `x == 0`. With no zero
  value declared, it joins set flag names with `|`; the empty set renders as
  `Permission(0)`.
- `Parse` accepts a single name or a `|`-separated combination.
- `IsValid` reports whether every set bit is declared.
- `flags` pulls in `string`, `parse` and `valid`.
- JSON/YAML/text/SQL/flag features build on these methods.

## Lists and pointers

### `names`

```go
func ColorNames() []string
```

Returns canonical texts in declaration order. Aliases are not included.

### `values`

```go
func ColorValues() []Color
```

Returns constants in declaration order. With duplicate numeric values, both
constants appear.

### `ptr`

```go
func (x Color) Ptr() *Color
```

Handy for optional struct fields and API payloads.

## Imports and module dependencies

| Feature | Imports |
| --- | --- |
| `string` (numeric only) | `fmt` |
| `parse` | `errors`, `fmt`; `strings` when case-insensitive or flags |
| `valid` | `strings` when case-insensitive |
| `flags` | `fmt`, `strings` |
| `json` | `encoding/json` |
| `json:number` | `fmt`, `strconv` |
| `yaml` | `gopkg.in/yaml.v3` |
| `sql` | `database/sql/driver`, `fmt`; `strconv` for numeric enums |
| `sql:int` | `database/sql/driver`, `fmt`, `strconv` |
| `sql:null` | `database/sql/driver`, `encoding/json` |
| `text`, `mustparse`, `names`, `values`, `ptr`, `flag` | stdlib only |

Only the external dependency is `gopkg.in/yaml.v3`, and only when `yaml` is
enabled. Imports are deduplicated and the final file is formatted with
`go/format.Source`.

## Choosing a feature set

| Requirement | Manifest |
| --- | --- |
| Minimal type-safe string enum | `defaults.features: [string, parse, valid]` |
| JSON API, names on the wire | `features: [json]` |
| JSON API, numbers on the wire | `type: int32`, `features: [json:number]` |
| Database column | `features: [sql]` |
| Nullable database column | `features: [sql, sql:null]` |
| Integer database column for a string enum | `features: [sql:int]`, `number` per value |
| CLI flag | `features: [flag]` |
| Bitmask | `type: uint32`, `features: [flags]` |
| YAML config | `features: [yaml]` |
| Panic-on-invalid | `features: [mustparse]` |
| Only methods, no automatic comments | `defaults.comments: none` |
| No features at all | `defaults.features: []` |
