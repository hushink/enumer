package generator

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hushink/enumer/internal/diag"
	"github.com/hushink/enumer/model"
)

type enumEntry struct {
	manifest *model.Manifest
	enum     *model.EnumSpec
	features []Feature
}

type symbolOwner struct {
	enum   string
	source model.Source
}

func (g *Generator) plan(_ context.Context, manifests []*model.Manifest) ([]Output, diag.List) {
	var diags diag.List

	sorted := append([]*model.Manifest(nil), manifests...)
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i] == nil {
			return false
		}
		if sorted[j] == nil {
			return true
		}
		return sorted[i].Source.File < sorted[j].Source.File
	})

	entries := make([][]enumEntry, len(sorted))
	for mi, m := range sorted {
		if m == nil {
			continue
		}
		for ei := range m.Enums {
			e := &m.Enums[ei]
			features, err := g.registry.resolve(e)
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Code:     "invalid-feature",
					Message:  err.Error(),
					File:     e.Source.File,
					Line:     e.Source.Line,
					Column:   e.Source.Column,
					Enum:     e.Name,
				})
				continue
			}
			entries[mi] = append(entries[mi], enumEntry{manifest: m, enum: e, features: features})
		}
	}
	if diags.HasErrors() {
		return nil, diags
	}

	type packageSymbols struct {
		pkg     string
		symbols map[string]symbolOwner
	}
	packages := map[string]*packageSymbols{}
	addSymbols := func(dir, pkg string, e *model.EnumSpec, features []Feature) {
		key := strings.ToLower(filepath.Clean(dir))
		info, ok := packages[key]
		if !ok {
			info = &packageSymbols{pkg: pkg, symbols: map[string]symbolOwner{}}
			packages[key] = info
		} else if info.pkg != pkg {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Code:     "package-mismatch",
				Message:  fmt.Sprintf("output directory %s contains both package %s and package %s", dir, info.pkg, pkg),
				File:     e.Source.File,
				Line:     e.Source.Line,
				Column:   e.Source.Column,
				Enum:     e.Name,
			})
		}
		for _, sym := range enumSymbols(e, features) {
			if prev, exists := info.symbols[sym]; exists {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Code:     "symbol-collision",
					Message:  fmt.Sprintf("generated symbol %q for enum %s conflicts with enum %s (declared at %s)", sym, e.Name, prev.enum, prev.source.String()),
					File:     e.Source.File,
					Line:     e.Source.Line,
					Column:   e.Source.Column,
					Enum:     e.Name,
				})
				continue
			}
			info.symbols[sym] = symbolOwner{enum: e.Name, source: e.Source}
		}
	}

	var outputs []Output
	for mi, m := range sorted {
		if m == nil {
			continue
		}

		if m.Aggregate {
			if m.Output == "" {
				diags = append(diags, diag.Diagnostic{Severity: diag.Error, Code: "invalid-output", Message: "aggregate output path is empty", File: m.Source.File, Line: 1})
				continue
			}
			enums := make([]*model.EnumSpec, 0, len(m.Enums))
			features := make([][]Feature, 0, len(m.Enums))
			for i := range m.Enums {
				enums = append(enums, &m.Enums[i])
			}
			for _, entry := range entries[mi] {
				features = append(features, entry.features)
			}
			if len(features) != len(enums) {
				continue
			}
			pkg := m.Package
			if len(enums) > 0 {
				pkg = enums[0].Package
			}
			for i := range enums {
				addSymbols(filepath.Dir(m.Output), pkg, enums[i], features[i])
			}
			content, err := renderFile(pkg, headerComment(m, m.Output), enums, features)
			if err != nil {
				diags = append(diags, diag.Diagnostic{Severity: diag.Error, Code: "format-error", Message: err.Error(), File: m.Output})
				continue
			}
			outputs = append(outputs, Output{Path: m.Output, Package: pkg, Enums: enums, Content: content, Kind: "go"})
		} else {
			for _, entry := range entries[mi] {
				e := entry.enum
				if e.Output == "" {
					diags = append(diags, diag.Diagnostic{Severity: diag.Error, Code: "invalid-output", Message: "enum output path is empty", File: e.Source.File, Line: e.Source.Line, Column: e.Source.Column, Enum: e.Name})
					continue
				}
				addSymbols(filepath.Dir(e.Output), e.Package, e, entry.features)
				content, err := renderFile(e.Package, headerComment(m, e.Output), []*model.EnumSpec{e}, [][]Feature{entry.features})
				if err != nil {
					diags = append(diags, diag.Diagnostic{Severity: diag.Error, Code: "format-error", Message: err.Error(), File: e.Output, Enum: e.Name})
					continue
				}
				outputs = append(outputs, Output{Path: e.Output, Package: e.Package, Enums: []*model.EnumSpec{e}, Content: content, Kind: "go"})
			}
		}

		targetOutputs, tdiags := renderTargets(m)
		diags = append(diags, tdiags...)
		outputs = append(outputs, targetOutputs...)
	}

	diags = append(diags, checkOutputPaths(outputs)...)
	if diags.HasErrors() {
		return nil, diags
	}

	sort.SliceStable(outputs, func(i, j int) bool { return outputs[i].Path < outputs[j].Path })
	return outputs, diags
}

func enumSymbols(e *model.EnumSpec, features []Feature) []string {
	syms := make([]string, 0, len(e.Values)+8)
	syms = append(syms, e.Name)
	for i := range e.Values {
		syms = append(syms, e.Values[i].ConstName)
	}
	for _, f := range features {
		switch f.Name() {
		case "parse":
			syms = append(syms, "Parse"+e.Name, "ErrInvalid"+e.Name, "_"+e.Name+"Value")
		case "mustparse":
			syms = append(syms, "MustParse"+e.Name)
		case "flags":
			syms = append(syms, "_"+e.Name+"Flags", "_"+e.Name+"FlagNames")
		case "sql:int":
			syms = append(syms, "_"+e.Name+"ByNumber")
		case "sql:null":
			syms = append(syms, "Null"+e.Name)
		case "names":
			syms = append(syms, e.Name+"Names")
		case "values":
			syms = append(syms, e.Name+"Values")
		}
	}
	return syms
}

func checkOutputPaths(outputs []Output) diag.List {
	var diags diag.List
	seen := map[string]string{}
	for _, out := range outputs {
		key := strings.ToLower(filepath.Clean(out.Path))
		if prev, ok := seen[key]; ok {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Code:     "output-collision",
				Message:  fmt.Sprintf("output %s conflicts with %s", out.Path, prev),
				File:     out.Path,
			})
			continue
		}
		seen[key] = out.Path
	}
	return diags
}
