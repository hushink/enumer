// Copyright (c) 2026 The enumer Authors
// SPDX-License-Identifier: MIT

package manifest

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/hushink/enumer/internal/diag"
	"gopkg.in/yaml.v3"
)

// loadExtendedDefaults loads a base manifest and returns its resolved defaults.
// Only version, extends and defaults are used; enums, targets, package and
// output are ignored.
func loadExtendedDefaults(path, baseDir string, visited map[string]bool) (defaultsSchema, diag.List) {
	p := path
	if !filepath.IsAbs(p) {
		p = filepath.Join(baseDir, p)
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		return defaultsSchema{}, diag.List{errDiag("extends-read", err.Error(), path, 0, 0, "", "")}
	}
	display := filepath.Clean(p)

	if visited[abs] {
		return defaultsSchema{}, diag.List{errDiag("extends-cycle", fmt.Sprintf("extends cycle detected at %s", display), display, 0, 0, "", "")}
	}
	visited[abs] = true
	defer delete(visited, abs)

	data, err := os.ReadFile(abs)
	if err != nil {
		return defaultsSchema{}, diag.List{errDiag("extends-read", err.Error(), display, 0, 0, "", "")}
	}

	var base fileSchema
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(&base); err != nil {
		return defaultsSchema{}, decodeDiagnostics(display, err)
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		if err == nil {
			return defaultsSchema{}, diag.List{errDiag("manifest-decode", "manifest contains more than one YAML document", display, 0, 0, "", "")}
		}
		return defaultsSchema{}, decodeDiagnostics(display, err)
	}

	version := base.Version
	if version == 0 {
		version = 1
	}
	if version != 1 {
		return defaultsSchema{}, diag.List{errDiag("unsupported-version", fmt.Sprintf("unsupported manifest version %d; supported version is 1", version), display, 1, 0, "", "")}
	}

	var inherited defaultsSchema
	var diags diag.List
	if base.Extends != "" {
		inherited, diags = loadExtendedDefaults(base.Extends, filepath.Dir(abs), visited)
		if diags.HasErrors() {
			return defaultsSchema{}, diags
		}
	}
	return mergeDefaults(inherited, base.Defaults), diags
}

func mergeDefaults(base, child defaultsSchema) defaultsSchema {
	out := base
	if child.Features != nil {
		out.Features = child.Features
	}
	if child.DisableFeatures != nil {
		out.DisableFeatures = child.DisableFeatures
	}
	if child.Naming != nil {
		out.Naming = mergeNaming(base.Naming, child.Naming)
	}
	if child.Parse != nil {
		out.Parse = mergeParse(base.Parse, child.Parse)
	}
	if child.Output != nil {
		out.Output = mergeOutput(base.Output, child.Output)
	}
	if child.Comments != nil {
		out.Comments = child.Comments
	}
	if child.AllowDuplicateValues != nil {
		out.AllowDuplicateValues = child.AllowDuplicateValues
	}
	if child.TypeMode != nil {
		out.TypeMode = child.TypeMode
	}
	return out
}

func mergeNaming(base, child *namingSchema) *namingSchema {
	if base == nil {
		return child
	}
	if child == nil {
		return base
	}
	out := *base
	if child.Prefix != nil {
		out.Prefix = child.Prefix
	}
	if child.Identifier != nil {
		out.Identifier = child.Identifier
	}
	if child.ExtraInitialisms != nil {
		out.ExtraInitialisms = child.ExtraInitialisms
	}
	if child.DisableInitialisms != nil {
		out.DisableInitialisms = child.DisableInitialisms
	}
	return &out
}

func mergeParse(base, child *parseSchema) *parseSchema {
	if base == nil {
		return child
	}
	if child == nil {
		return base
	}
	out := *base
	if child.CaseSensitive != nil {
		out.CaseSensitive = child.CaseSensitive
	}
	return &out
}

func mergeOutput(base, child *outputSchema) *outputSchema {
	if base == nil {
		return child
	}
	if child == nil {
		return base
	}
	out := *base
	if child.Suffix != nil {
		out.Suffix = child.Suffix
	}
	if child.Dir != nil {
		out.Dir = child.Dir
	}
	return &out
}
