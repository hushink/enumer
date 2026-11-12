# Contributing

## Setup

The generator requires Go 1.27 or newer.

```sh
go build ./...
go test ./...
```

## Common commands

```sh
make test       # go test ./...
make race       # go test -race ./...
make lint       # golangci-lint run ./...
make check      # verify committed generated files are up to date
make golden     # regenerate generator/testdata golden files
make build      # build bin/enumer
make license    # SPDX headers + dependency license check
make vuln       # govulncheck
```

Without `make`:

```sh
go test ./...
golangci-lint run ./...
go run ./cmd/enumer check ./...
go test ./generator -run TestGolden -update
```

## Architecture

The pipeline is fixed:

```
manifest parsing -> model (normalized IR) -> feature planning -> rendering -> atomic write
```

- `manifest` decodes and validates YAML, resolves defaults and produces `model`.
- `model` is the public, source-independent IR.
- `generator` validates feature sets, renders fragments, checks symbol and output
  collisions, renders custom targets and writes files.
- `internal/cli` is a thin command layer.

Read [docs/adr/0001-yaml-manifest-and-feature-registry.md](docs/adr/0001-yaml-manifest-and-feature-registry.md)
before changing the input model or the extension points. Update
[GLOSSARY.md](GLOSSARY.md) when a domain term changes.

## Adding a feature

1. Implement `generator.Feature` in a new `generator/features_*.go` file.
2. Register it in `generator.DefaultRegistry`.
3. Declare its dependencies in `Requires` and validate enum-specific constraints
   in `Validate`.
4. Emit only imports that the fragment actually uses.
5. Add behavior coverage to `generator/generator_test.go` (the compile test
   renders a temp module and runs `go test` inside it).
6. Update the feature table in `README.md` and `docs/manifest-schema.md`.

Custom feature names should use `name` or `name:variant` and must not collide
with built-in names.

## Golden files

`generator/testdata/basic/expected` contains committed output. After an
intentional output change:

```sh
go test ./generator -run TestGolden -update
```

Review the diff before committing. The golden test uses a relative header path,
so it is stable across machines.

## License headers

Every authored Go file starts with:

```go
// Copyright (c) 2026 The enumer Authors
// SPDX-License-Identifier: MIT
```

Add it manually or run:

```sh
go run github.com/apache/skywalking-eyes/cmd/license-eye@v0.9.0 header fix
```

Then verify:

```sh
make license
```

Generated files in `examples/**/*_enum.go` and `generator/testdata/**` are
excluded from the header check. Dependency licenses and known vulnerabilities
are checked in CI as well.

## Release token

The release workflow uses the automatic `GITHUB_TOKEN` with
`permissions: contents: write`. No repository secret is required, and GitHub
reserves the `GITHUB_` prefix for secrets, so a secret with that name cannot be
created.

If you intentionally need a long-lived token — for example, to trigger another
workflow or to bypass branch protection — create a fine-grained PAT with
`contents: write` for this repository, store it as `RELEASE_TOKEN`, and change
the workflow env to `GITHUB_TOKEN: ${{ secrets.RELEASE_TOKEN }}`. The automatic
token is safer and is the default.

## Commits

Use conventional commit prefixes: `feat:`, `fix:`, `docs:`, `test:`, `chore:`.
Keep the subject imperative and under ~72 characters.

## Generated examples

`examples/basic` contains committed generated files. Regenerate them with:

```sh
go run ./cmd/enumer generate ./examples/basic
```

CI runs `go run ./cmd/enumer check ./...` to catch drift.
