# CLI reference

`enumer` is a single binary. It reads manifests, writes Go files and reports
diagnostics. This reference covers installation, commands, flags, discovery,
pruning, `go:generate`, CI and the programmatic API.

## Contents

- [Installation](#installation)
- [Commands](#commands)
- [Manifest paths and discovery](#manifest-paths-and-discovery)
- [Global flags](#global-flags)
- [Diagnostics and exit codes](#diagnostics-and-exit-codes)
- [Pruning](#pruning)
- [go:generate](#gogenerate)
- [CI](#ci)
- [Custom targets](#custom-targets)
- [Programmatic API](#programmatic-api)
- [Troubleshooting](#troubleshooting)

## Installation

Building the CLI from source requires Go 1.27+; release binaries need no Go
toolchain. Generated code requires Go 1.20+.

```sh
# Project tool dependency (recommended for reproducible generation):
go get -tool github.com/hushink/enumer/cmd/enumer@latest
go tool enumer --help

# Global binary:
go install github.com/hushink/enumer/cmd/enumer@latest
enumer --help

# One-off run:
go run github.com/hushink/enumer/cmd/enumer@latest --help
```

Prebuilt binaries for Linux, macOS and Windows (amd64 and arm64) are published
at <https://github.com/hushink/enumer/releases>. Use a release binary when the
project's Go toolchain is older than 1.27 but the generated code still needs to
compile with Go 1.20+.

Pin a version in CI. A `tool` directive in `go.mod` plus `go tool enumer` is the
most reproducible option; commit `go.mod` and `go.sum`.

## Commands

```sh
enumer [generate] [paths...]   # generate; no subcommand behaves like generate
enumer generate [paths...]     # explicit generate; alias: gen
enumer check [paths...]        # verify generated files are up to date
enumer init                    # create a sample enumer.yaml and schema in the project root
enumer version                 # print the version
```

With no paths, commands resolve a manifest from the current directory, the
nearest module root, then a workspace root.

### `generate`

```sh
enumer generate ./...
enumer generate path/to/enumer.yaml
enumer generate ./services/order ./services/user
```

- Renders every planned Go file, formats it with `go/format.Source`, and writes
  it atomically.
- Unchanged files are not rewritten.
- Prunes stale generated files for enums removed from the manifest unless
  `--no-prune` is set.
- `--dry-run` prints `would write` and `would remove` lines to stderr without
  touching the filesystem.
- `-v` prints `wrote`, `unchanged` and `removed` paths.

### `check`

```sh
enumer check ./...
enumer check --diff ./...
```

- Renders the expected outputs in memory and compares them with the files on
  disk.
- Reports stale outputs, missing outputs and orphaned generated files as
  diagnostics.
- Exits `1` when anything is stale.
- `--diff` prints unified diffs; with `--json`, diffs go to stderr so stdout
  stays valid JSON.
- `-v` prints `verified N outputs` when everything is current.

Use `check` in CI. It catches a manifest that was edited without regenerating,
and generated files that were edited by hand.

### `init`

```sh
enumer init
enumer init --package order
enumer init --force
enumer init --local
```

Creates `enumer.yaml` in the nearest project root—the directory containing
`go.mod`, or `go.work` when there is no enclosing module—with a sample string
enum and a sample numeric enum. It also writes `enumer.schema.json` (the JSON
Schema used by editors) and `templates/README.md` describing custom targets.
The manifest starts with both `# $schema: ./enumer.schema.json` and
`$schema: ./enumer.schema.json` for editor completion; the generator ignores
them. Keep `enumer.schema.json` next to the manifest and commit both. For
IntelliJ IDEA, the schema works with a global JSON Schema mapping (`enumer.yaml`
/ `enumer.yml` file patterns), the file-level `$schema` lines, or the
status-bar schema switcher.

- `--package` writes a `package:` field.
- `--force` overwrites an existing `enumer.yaml`, schema file and template
  README.
- `--local` writes in the current directory instead of the project root. Use it
  only for a deliberate package-local manifest.

For a focused task, writing the manifest directly is often cleaner than
`init`, which adds sample enums you may not want.

### `version`

```sh
enumer version
enumer --version
```

Both print the version, commit and build date. The `version` subcommand prints
`enumer <version>`; `--version` uses the same format.

## Manifest paths and discovery

A path can be:

- a manifest file, even if it is not named `enumer.yaml`;
- a directory containing `enumer.yaml` or `enumer.yml`;
- `./...` (or any `...` pattern) to recurse.

Place the manifest in the nearest project root: the `go.mod` directory, or the
`go.work` directory when there is no enclosing module. Module-root manifests
take precedence; a workspace-root manifest is the fallback. With no paths,
commands use the current directory's manifest, then the nearest module-root
manifest, then a workspace-root manifest. `enumer generate ./...` discovers
every manifest under the current tree.

Discovery rules:

- Accepts `.yaml` and `.yml`.
- Skips `vendor/`, `testdata/`, hidden directories and directories whose names
  start with `_`.
- A directory containing both `enumer.yaml` and `enumer.yml` is an error.
- Finding no manifests is an error.
- `extends` base files are not discovered unless they are named
  `enumer.yaml` or `enumer.yml`.

## Global flags

These flags are persistent and can be used with any command:

| Flag | Description |
| --- | --- |
| `--strict` | Treat warnings as errors. |
| `--json` | Print diagnostics as a JSON report on stdout. |
| `-v`, `--verbose` | Print written, unchanged, removed or verified files. |
| `--no-prune` | Keep generated files for enums no longer declared. |

Command-specific flags:

| Command | Flag | Description |
| --- | --- | --- |
| `generate` | `--dry-run` | Show planned writes and removals without writing. |
| `check` | `--diff` | Print unified diffs for stale outputs. |
| `init` | `--package` | Go package name to write into the manifest. |
| `init` | `--force` | Overwrite existing files. |
| `init` | `--local` | Write in the current directory instead of the project root. |

Examples:

```sh
enumer --strict generate ./...
enumer --json check ./...
enumer --verbose generate ./...
enumer --no-prune generate ./...
enumer generate --dry-run ./...
enumer check --diff ./...
```

## Diagnostics and exit codes

Human-readable diagnostics go to stderr and include the file position:

```text
path/to/enumer.yaml:12:5: enumer: duplicate text "red" in enum Color (first declared at path/to/enumer.yaml:8:3)
```

`--json` writes a structured report to stdout:

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

- Diagnostic codes are stable; prefer them over matching message text.
- `--strict` upgrades warnings to errors, changing the exit code accordingly.
- Exit codes:
  - `0` — success.
  - `1` — validation, generation or stale-check failure.
  - `2` — command line usage error.

## Pruning

When an enum is removed from a manifest, its old generated file is pruned on the
next `generate`:

```sh
enumer generate ./...
# removed color_enum.go
```

The match is deliberately strict: a file is removed only if its first line is
exactly the generated header for that manifest and path. This protects
hand-written files, files generated from other manifests and custom target
outputs.

- `--no-prune` keeps stale files.
- `enumer check` reports orphaned generated files as stale, so CI catches them.
- `--dry-run` prints `would remove <path>` without deleting anything.

## go:generate

With a project tool dependency:

```go
//go:generate go tool enumer generate
```

With a global binary:

```go
//go:generate enumer generate
```

Then run:

```sh
go generate ./...
```

The directive is often placed in a `doc.go` or next to the package's Go files.
Keep it in the same directory as the manifest, or pass an explicit path.

## CI

A typical check step:

```yaml
- run: go tool enumer check ./...
```

If the project does not use a tool dependency, install a pinned release binary
or run a pinned version:

```sh
go install github.com/hushink/enumer/cmd/enumer@v0.1.0
enumer check ./...
```

`check` fails when generated files are missing, stale or orphaned. It does not
run `go build`; add a separate build/test step:

```sh
go build ./...
go test ./...
```

## Custom targets

Custom targets render extra files from `text/template`. They run after Go files
are generated and participate in `enumer check`.

```yaml
targets:
  - template: ./templates/enums.md.tmpl
    output: ENUMS.md
  - template: ./templates/schema.json.tmpl
    output: "{{.Enum.Name}}.json"
    perEnum: true
```

- Template paths and output paths are relative to the manifest directory.
- `.Manifest`, `.Enums` and `.Enum` are available; `.Enum` is non-nil only when
  `perEnum: true`.
- Functions: `camel`, `snake`, `kebab`, `lower`, `upper`, `title`, `quote`,
  `join`, `json`, `jsonPretty`, `default`.
- Outputs ending in `.go` are formatted with `go/format.Source` and must render
  a complete Go file.
- Custom target outputs are not pruned.

See the [manifest reference](manifest.md#custom-targets) for the full schema.

## Programmatic API

The `model`, `manifest` and `generator` packages are public:

```go
m, diags := manifest.Load("enumer.yaml")
if diags.HasErrors() {
    log.Fatal(diags.Format())
}

g := generator.New()
result, diags := g.Generate(context.Background(), []*model.Manifest{m}, generator.GenerateOptions{})
if diags.HasErrors() {
    log.Fatal(diags.Format())
}
fmt.Println(result.Changed)
```

Custom features implement `generator.Feature` (`Name`, `Requires`, `Validate`,
`Emit`) and can be registered on a private registry:

```go
registry := generator.DefaultRegistry()
if err := registry.Register(myFeature{}); err != nil {
    log.Fatal(err)
}
g := generator.New(generator.WithRegistry(registry))
```

Custom features are responsible for avoiding package-level symbol collisions;
built-in collision checks only know about built-in features.

## Troubleshooting

| Symptom | Cause and fix |
| --- | --- |
| `no manifests found` | The path has no `enumer.yaml`/`.yml`. Pass a manifest file or create one. |
| `package could not be inferred` | The directory has no Go files. Set `package:` explicitly. |
| `type ... exists` with `typeMode: emit` | A hand-written declaration already exists. Use `auto` or remove the duplicate. |
| `missing type` with `typeMode: existing` | Add the hand-written `type <Name> <underlying>` declaration in the same package. |
| `type mismatch` | The hand-written type's underlying type differs from the manifest `type`. |
| `duplicate text` / `duplicate value` | Two values or aliases collide. Rename one, or enable `allowDuplicateValues` for numeric duplicates. |
| `flags requires an unsigned integer type` | Change the enum to `uint`, `uint8`…`uint64`. |
| `must be zero or a power of two` | Fix the bitmask values or remove `flags`. |
| `sql:int requires number` | Add `number` to every value of the string enum. |
| `json:number requires a numeric enum` | Use `json` or change the underlying type. |
| `sql:null requires exactly one of sql or sql:int` | Adjust the feature set. |
| Generated code cannot find `yaml` | Run `go get gopkg.in/yaml.v3 && go mod tidy`. |
| `check` fails after editing generated files | Regenerate; do not edit generated files. |
| Stale file is not pruned | Its header does not match the current manifest, so enumer treats it as hand-written or foreign. Inspect the first line. |
| `go: command not found` | Install Go; the CLI is a Go binary. |
| `module requires Go 1.27` when using `go run` | Use a prebuilt release binary or upgrade the toolchain. |
| Generated names differ from the old DSL | Adjust `naming` or add explicit `name`/`value` entries before regenerating. |
