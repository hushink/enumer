// Package model defines the normalized, source-independent representation of
// an enumer manifest. It is the only input to generators and custom templates.
//
// The model package is part of enumer's public API. Fields may be added in a
// backwards-compatible way; existing fields keep their meaning for the whole
// v1 schema.
package model

import (
	"fmt"
	"sort"
)

// SchemaVersion is the version of the normalized intermediate representation.
const SchemaVersion = 1

// Kind describes the underlying kind of an enum type.
type Kind int

const (
	// KindString is a string-backed enum.
	KindString Kind = iota
	// KindSigned is a signed integer-backed enum.
	KindSigned
	// KindUnsigned is an unsigned integer-backed enum.
	KindUnsigned
)

// String implements fmt.Stringer.
func (k Kind) String() string {
	switch k {
	case KindString:
		return "string"
	case KindSigned:
		return "signed integer"
	case KindUnsigned:
		return "unsigned integer"
	default:
		return "unknown"
	}
}

// Type is the resolved underlying Go type of an enum.
type Type struct {
	// Name is the Go type name, for example "string", "int32" or "uint64".
	Name string
	// Kind classifies the type.
	Kind Kind
	// Bits is the width of integer types. It is 0 for string enums.
	// The platform-sized int and uint types are reported as 64 bits.
	Bits int
}

// IsString reports whether t is a string type.
func (t Type) IsString() bool { return t.Kind == KindString }

// IsInteger reports whether t is an integer type.
func (t Type) IsInteger() bool { return t.Kind == KindSigned || t.Kind == KindUnsigned }

var supportedTypes = map[string]Type{
	"string": {Name: "string", Kind: KindString, Bits: 0},
	"int":    {Name: "int", Kind: KindSigned, Bits: 64},
	"int8":   {Name: "int8", Kind: KindSigned, Bits: 8},
	"int16":  {Name: "int16", Kind: KindSigned, Bits: 16},
	"int32":  {Name: "int32", Kind: KindSigned, Bits: 32},
	"int64":  {Name: "int64", Kind: KindSigned, Bits: 64},
	"uint":   {Name: "uint", Kind: KindUnsigned, Bits: 64},
	"uint8":  {Name: "uint8", Kind: KindUnsigned, Bits: 8},
	"uint16": {Name: "uint16", Kind: KindUnsigned, Bits: 16},
	"uint32": {Name: "uint32", Kind: KindUnsigned, Bits: 32},
	"uint64": {Name: "uint64", Kind: KindUnsigned, Bits: 64},
}

// LookupType resolves a Go type name to a Type.
func LookupType(name string) (Type, bool) {
	t, ok := supportedTypes[name]
	return t, ok
}

// SupportedTypes returns the accepted type names in alphabetical order.
func SupportedTypes() []string {
	names := make([]string, 0, len(supportedTypes))
	for name := range supportedTypes {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Import is an import required by a generated fragment.
type Import struct {
	// Path is the import path, for example "encoding/json".
	Path string
	// Alias is the optional import alias. Empty means the default name.
	Alias string
}

// Source points at a location in a manifest.
type Source struct {
	// File is the manifest path used for diagnostics and file headers.
	File string
	// Line is the 1-based line number, or 0 when unknown.
	Line int
	// Column is the 1-based column number, or 0 when unknown.
	Column int
}

// IsZero reports whether s carries no location.
func (s Source) IsZero() bool { return s.File == "" }

// String renders the location as file:line:column.
func (s Source) String() string {
	if s.File == "" {
		return ""
	}
	if s.Line == 0 {
		return s.File
	}
	if s.Column == 0 {
		return fmt.Sprintf("%s:%d", s.File, s.Line)
	}
	return fmt.Sprintf("%s:%d:%d", s.File, s.Line, s.Column)
}

// Manifest is one normalized enumer.yaml file.
type Manifest struct {
	// SchemaVersion is the manifest schema version (currently 1).
	SchemaVersion int
	// Source locates the manifest.
	Source Source
	// Dir is the absolute directory containing the manifest.
	Dir string
	// Package is the default Go package name for enums in this manifest.
	Package string
	// Aggregate controls whether all enums are emitted into one file.
	Aggregate bool
	// Output is the absolute aggregate output path. It is empty unless
	// Aggregate is true.
	Output string
	// Enums are the normalized enum declarations in manifest order.
	Enums []EnumSpec
	// Targets are the custom output targets in manifest order.
	Targets []TargetSpec
}

// EnumSpec is one normalized enum declaration.
type EnumSpec struct {
	// Name is the Go type name.
	Name string
	// Package is the resolved Go package name.
	Package string
	// Doc is the user-provided documentation.
	Doc string
	// Type is the resolved underlying type.
	Type Type
	// Source locates the enum in the manifest.
	Source Source
	// Features are the configured (not yet resolved) generation features.
	Features []string
	// DisabledFeatures are the features explicitly disabled for this enum.
	DisabledFeatures []string
	// AllowDuplicateValues permits multiple names for the same numeric value.
	AllowDuplicateValues bool
	// Start is the canonical numeric start value for numeric enums.
	Start string
	// Naming holds the resolved identifier settings.
	Naming Naming
	// CaseSensitive controls parsing and alias matching.
	CaseSensitive bool
	// Comments controls automatic documentation comments.
	Comments CommentsMode
	// TypeMode controls whether the generated file declares the type.
	TypeMode TypeMode
	// Values are the normalized values in declaration order.
	Values []ValueSpec
	// Output is the absolute output path when the enum is emitted alone.
	// It is empty when the enum is part of an aggregate file.
	Output string
}

// IsString reports whether the enum is string-backed.
func (e EnumSpec) IsString() bool { return e.Type.IsString() }

// HasFeature reports whether name is in the configured feature list.
func (e EnumSpec) HasFeature(name string) bool {
	for _, f := range e.Features {
		if f == name {
			return true
		}
	}
	return false
}

// Naming holds identifier generation settings.
type Naming struct {
	// Prefix is prepended to every constant name. It may be empty.
	Prefix string
	// Identifier selects how value names become Go identifiers.
	Identifier IdentifierStyle
	// Initialisms is the set of words that keep their uppercase form.
	Initialisms map[string]bool
}

// TypeMode controls whether the generated file declares the enum type.
type TypeMode int

const (
	// TypeModeEmit emits the type declaration in the generated file.
	TypeModeEmit TypeMode = iota
	// TypeModeExisting references a type declared elsewhere in the package.
	TypeModeExisting
)

// IdentifierStyle selects the value-name to identifier conversion.
type IdentifierStyle int

const (
	// IdentifierCamel converts value names to CamelCase.
	IdentifierCamel IdentifierStyle = iota
	// IdentifierRaw keeps the value name unchanged.
	IdentifierRaw
)

// CommentsMode controls automatic documentation comments.
type CommentsMode int

const (
	// CommentsAuto emits automatic comments when no user doc is present.
	CommentsAuto CommentsMode = iota
	// CommentsNone suppresses automatic comments but keeps user docs.
	CommentsNone
)

// ValueSpec is one normalized enum value.
type ValueSpec struct {
	// Name is the value name as written in the manifest.
	Name string
	// ConstName is the generated Go constant name.
	ConstName string
	// Text is the canonical text used by String and Parse.
	Text string
	// Number is the canonical numeric literal for numeric enums.
	Number string
	// HasNumber reports whether Number is set.
	HasNumber bool
	// SQLNumber is the canonical int64 literal used by sql:int.
	SQLNumber string
	// HasSQLNumber reports whether SQLNumber is set.
	HasSQLNumber bool
	// Aliases are additional parseable texts.
	Aliases []string
	// Doc is the user-provided documentation.
	Doc string
	// Deprecated is the deprecation reason. It is empty for live values.
	Deprecated string
	// Source locates the value in the manifest.
	Source Source
}

// TargetSpec is one custom template target.
type TargetSpec struct {
	// Template is the absolute path to the template file.
	Template string
	// Output is the output path template, resolved relative to the manifest.
	Output string
	// PerEnum executes the template once per enum instead of once per manifest.
	PerEnum bool
	// Source locates the target in the manifest.
	Source Source
}
