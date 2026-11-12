# Enumer

`enumer` generates type-safe Go enumerations from YAML manifests. This file
defines the core terms of the project's domain language.

## Language

**Manifest**:
An `enumer.yaml` file. It is the single source of truth for enum declarations.
_Avoid_: config file, DSL

**Enum Declaration**:
One enum entry inside a manifest, describing an enum's type, values and settings.
_Avoid_: enum annotation, ENUM directive

**EnumSpec**:
The normalized, source-independent description of a single enum after parsing.
It is the only input to generation.
_Avoid_: enum model, enum object

**ValueSpec**:
The normalized description of one enum value (constant), including its
identifier, canonical value, aliases and documentation.
_Avoid_: enum item, member

**Feature**:
An independently switchable and testable code generation unit (for example
String, SQL, JSON). Features compose into the generated result of an enum.
_Avoid_: plugin, generator

**Fragment**:
The unformatted Go source and import declarations contributed by one feature.
_Avoid_: code block, template output

**Feature Registry**:
The catalog that records every available feature by name and resolves feature
dependencies.
_Avoid_: plugin system, factory

**Generation Pipeline**:
The fixed sequence from declaration parsing to file output: parse → EnumSpec →
generation plan → render → write.
_Avoid_: generation flow, processing chain

**Diagnostic**:
One validation or generation report carrying a source location and a stable
error code, renderable as text or JSON.
_Avoid_: error log, message

**Target**:
A file to produce in one generation run. Built-in targets are `_enum.go` files;
custom targets are additional files.
_Avoid_: output file, artifact

**Custom Target**:
An additional file declared under `targets` in a manifest and produced by an
external template.
_Avoid_: extra artifact, plugin

**Sink**:
The abstraction that writes rendered content to disk, responsible for atomic
writes and for not rewriting unchanged content.
_Avoid_: writer, outputter

**Defaults**:
Settings at the top of a manifest that apply to every enum in that manifest.
_Avoid_: global config, base config

**Enum Settings**:
Fields on one enum entry that override the defaults.
_Avoid_: override rules, special case

**Type Mode**:
Whether generated code declares the enum type (`emit`) or references a
hand-written declaration (`existing`).
_Avoid_: type source, declaration mode
