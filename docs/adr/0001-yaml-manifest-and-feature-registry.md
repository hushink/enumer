# YAML manifest as the single source of truth and the feature registry pipeline

**Status**: accepted

`enumer` does not use an abice/go-enum style comment DSL. Instead, an
`enumer.yaml` manifest is the single source of truth: the generator never parses
Go source, the manifest declares the type, values and settings, and the
generated file contains the type declaration, constants and methods. Generation
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

- The tool generates the type declaration; users must not declare the same type
  in Go source.
- The YAML schema is a public contract and must be versioned, strictly
  validated and preserve source locations for diagnostics.
- Adding a capability only requires registering a Feature; adding an artifact
  only requires a template. The renderer stays unchanged.
- The generator has no dependency on Go package loading;
  `golang.org/x/tools` is not a dependency.
