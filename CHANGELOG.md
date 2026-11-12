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
  `ptr`.
- `typeMode: emit|existing` to generate the type declaration or reference a
  hand-written one.
- `sql:null` wrappers marshal invalid values as JSON `null` and unmarshal
  `null` without error.
- Feature dependency resolution and strict feature validation.
- Public `model`, `manifest` and `generator` packages.
- CLI commands: `generate`, `check --diff`, `init`, `version`.
- Strict validation with positioned diagnostics, `--json` output and `--strict`.
- Atomic writes and idempotent output; unchanged files are not rewritten.
- Custom `text/template` targets, including per-enum targets.
- Golden, compile/behavior, CLI integration, fuzz and race tests.
- README, manifest schema, ADR, glossary and runnable examples.
- CI, release and Dependabot configuration.
