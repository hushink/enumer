// Copyright (c) 2026 The enumer Authors
// SPDX-License-Identifier: MIT

package generator

import (
	"bytes"
	"context"
	"os"

	"github.com/hushink/enumer/internal/diag"
	"github.com/hushink/enumer/internal/sink"
	"github.com/hushink/enumer/model"
)

// Generator renders manifests using a registry of features.
type Generator struct {
	registry *Registry
}

// Option configures a Generator.
type Option func(*Generator)

// WithRegistry replaces the default registry.
func WithRegistry(r *Registry) Option {
	return func(g *Generator) {
		if r != nil {
			g.registry = r
		}
	}
}

// New returns a generator with all built-in features.
func New(opts ...Option) *Generator {
	g := &Generator{registry: DefaultRegistry()}
	for _, opt := range opts {
		opt(g)
	}
	return g
}

// Registry exposes the generator's registry.
func (g *Generator) Registry() *Registry { return g.registry }

// GenerateOptions controls Generate.
type GenerateOptions struct {
	// DryRun plans outputs without writing them.
	DryRun bool
}

// CheckOptions controls Check.
type CheckOptions struct {
	// Diff records old and new content for stale outputs.
	Diff bool
}

// Output is one planned file.
type Output struct {
	// Path is the absolute output path.
	Path string
	// Package is the Go package name for Go outputs.
	Package string
	// Enums are the enums included in this output.
	Enums []*model.EnumSpec
	// Content is the rendered file content.
	Content []byte
	// Kind is "go" for generated Go files and "target" for template outputs.
	Kind string
}

// Result is the outcome of Generate.
type Result struct {
	Outputs   []Output
	Changed   []string
	Unchanged []string
}

// Diff is one stale output.
type Diff struct {
	Path string
	Old  []byte
	New  []byte
}

// CheckResult is the outcome of Check.
type CheckResult struct {
	Outputs []Output
	Stale   []string
	Diffs   []Diff
}

// Generate plans and writes every output.
func (g *Generator) Generate(ctx context.Context, manifests []*model.Manifest, opts GenerateOptions) (*Result, diag.List) {
	outputs, diags := g.plan(ctx, manifests)
	if diags.HasErrors() {
		return nil, diags
	}
	result := &Result{Outputs: outputs}
	if opts.DryRun {
		return result, diags
	}
	for _, output := range outputs {
		changed, err := sink.Write(output.Path, output.Content)
		if err != nil {
			diags = append(diags, errDiag("write-error", err.Error(), output.Path, 0, 0, ""))
			return result, diags
		}
		if changed {
			result.Changed = append(result.Changed, output.Path)
		} else {
			result.Unchanged = append(result.Unchanged, output.Path)
		}
	}
	return result, diags
}

// Check verifies that every planned output already exists with identical
// content.
func (g *Generator) Check(ctx context.Context, manifests []*model.Manifest, opts CheckOptions) (*CheckResult, diag.List) {
	outputs, diags := g.plan(ctx, manifests)
	if diags.HasErrors() {
		return nil, diags
	}
	result := &CheckResult{Outputs: outputs}
	for _, output := range outputs {
		old, err := os.ReadFile(output.Path)
		if os.IsNotExist(err) {
			result.Stale = append(result.Stale, output.Path)
			diags = append(diags, errDiag("stale-output", "generated file is missing", output.Path, 0, 0, ""))
			if opts.Diff {
				result.Diffs = append(result.Diffs, Diff{Path: output.Path, New: output.Content})
			}
			continue
		}
		if err != nil {
			diags = append(diags, errDiag("read-error", err.Error(), output.Path, 0, 0, ""))
			continue
		}
		if bytes.Equal(old, output.Content) {
			continue
		}
		result.Stale = append(result.Stale, output.Path)
		diags = append(diags, errDiag("stale-output", "generated file is out of date", output.Path, 0, 0, ""))
		if opts.Diff {
			result.Diffs = append(result.Diffs, Diff{Path: output.Path, Old: old, New: output.Content})
		}
	}
	diags.Sort()
	return result, diags
}

func errDiag(code, msg, file string, line, col int, enum string) diag.Diagnostic {
	return diag.Diagnostic{
		Severity: diag.Error,
		Code:     code,
		Message:  msg,
		File:     file,
		Line:     line,
		Column:   col,
		Enum:     enum,
	}
}
