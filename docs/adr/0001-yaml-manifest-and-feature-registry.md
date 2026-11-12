# YAML manifest as the single source of truth and the feature registry pipeline

**Status**: accepted

`enumer` does not use an abice/go-enum style comment DSL. Instead, an
`enumer.yaml` manifest is the single source of truth: the manifest declares the
type, values and settings, and the generated file contains the type declaration,
constants and methods. The generator does not use Go source to discover values;
`typeMode: auto` performs only a syntax-level declaration scan. Generation
always follows the pipeline "manifest parsing → normalized IR (model) → feature
registry planning → fragment rendering → atomic write". Every generation
capability implements a common `Feature` interface, can be switched on or off
independently and is tested in isolation; additional artifacts are produced by
`text/template` targets. The reason is that a comment DSL couples syntax,
parsing and generation, and is hard for non-Go tools to consume; a YAML manifest
can be consumed by other languages and tools, and the feature registry lets
capabilities grow without touching the renderer.

## Considered Options

- **Comment DSL (go-enum style)**: cheap migration, but requires parsing Go
  source, offers weak configuration and cannot be reused by non-Go tools.
- **Fixed templates plus a single config**: simple, but capabilities cannot be
  composed and every extension changes the template.
- **Go plugin extensions**: poor cross-platform and Go version compatibility;
  rejected.

## Consequences

- `typeMode: auto` (the default) detects an existing type declaration and emits
  the type only when it is missing; `emit` and `existing` force the behavior.
  Detection uses `go/parser` and is syntax-only; the generator still does not
  load or type-check Go packages.
- The YAML schema is a public contract and must be versioned, strictly
  validated and preserve source locations for diagnostics.
- Adding a capability only requires registering a Feature; adding an artifact
  only requires a template. The renderer stays unchanged.
- Manifests may inherit `defaults` from a base manifest through `extends`;
  enums, targets, package and output are never inherited.
- The generator has no dependency on Go package loading;
  `golang.org/x/tools` is not a dependency.
