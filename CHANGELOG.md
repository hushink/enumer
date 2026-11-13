# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.0] - 2026-11-13

### Added

- JSON Schema for `enumer.yaml` at `schema/enumer.schema.json`, embedded in the
  CLI. `enumer init` writes `enumer.schema.json` and adds `$schema` comment and
  key lines for editor completion and validation. The local schema supports
  IntelliJ IDEA global JSON Schema mappings, file-level `$schema` and the
  status-bar schema switcher.
- A public `schema` package that exposes the embedded JSON Schema.
- Workspace-aware project root resolution: `enumer init` and no-argument
  `generate`/`check` consider both `go.mod` and `go.work`. Module-root
  manifests take precedence over workspace-root manifests.
- An Agent Skill (`skills/enumer/`) compatible with skills.sh and the Agent
  Skills specification, with references for the manifest, features, CLI and
  recipes, plus `skills.sh.json` catalog metadata and README install
  instructions.

### Changed

- `enumer init` now writes to the nearest project root (`go.mod`, or `go.work`
  when there is no enclosing module) instead of always the current directory.
  Use `--local` to keep the previous current-directory behavior.
- With no path arguments, `generate` and `check` fall back from the current
  directory to the module root and then the workspace root.

## [0.1.0] - 2026-11-13

### Added

- YAML manifest as the single source of truth for type declarations, values and
  settings.
- String and integer enums (`int`, `int8`…`int64`, `uint`, `uint8`…`uint64`).
- Built-in features: `string`, `parse`, `valid`, `mustparse`, `text`, `json`,
  `json:number`, `yaml`, `sql`, `sql:int`, `sql:null`, `flag`, `names`, `values`,
  `ptr`. The default set includes `names` and `values`, so every enum exports
  its value list out of the box.
- `typeMode: auto|emit|existing`. `auto` is the default and scans the target
  package: it generates the type when missing, references an existing
  declaration otherwise, and reports aliases, generics and type mismatches.
- `doc`/`comment` overrides for enum and value comments, plus a per-enum
  `comments: auto|none` mode.
- Generated Go files and `.go` custom targets are formatted with the standard
  library `go/format`.
- `extends` for shared defaults across manifests; only `defaults` is inherited.
- `flags` for unsigned integer bit masks with `Has`, `With`, `Clear`, `Toggle`,
  `Flags` and combination-aware `String`/`Parse`/`IsValid`.
- `sql:null` wrappers marshal invalid values as JSON `null` and unmarshal
  `null` without error.
- Feature dependency resolution and strict feature validation.
- Public `model`, `manifest` and `generator` packages.
- CLI commands: `generate`, `check --diff`, `init`, `version`.
- Strict validation with positioned diagnostics, `--json` output and `--strict`.
- Atomic writes and idempotent output; unchanged files are not rewritten.
- Prune generated files for enums removed from a manifest. Only files with the
  manifest's exact generated header are removed; `--no-prune` keeps them and
  `check` reports orphans as stale.
- Custom `text/template` targets, including per-enum targets.
- Golden, compile/behavior, CLI integration, fuzz and race tests.
- README, configuration guide and manifest schema in English and Simplified
  Chinese (English primary), plus ADR, glossary and runnable examples.
- CI, release and Dependabot configuration.
- Compliance: SPDX MIT headers on authored Go files, `license-eye` header and
  dependency checks, third-party notices and `govulncheck` in CI.
