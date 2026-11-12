# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
- Custom `text/template` targets, including per-enum targets.
- Golden, compile/behavior, CLI integration, fuzz and race tests.
- README, configuration guide and manifest schema in English and Simplified
  Chinese (English primary), plus ADR, glossary and runnable examples.
- CI, release and Dependabot configuration.
- Compliance: SPDX MIT headers on authored Go files, `license-eye` header and
  dependency checks, third-party notices and `govulncheck` in CI.
