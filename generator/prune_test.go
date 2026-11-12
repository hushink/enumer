// Copyright (c) 2026 The enumer Authors
// SPDX-License-Identifier: MIT

package generator

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/hushink/enumer/manifest"
	"github.com/hushink/enumer/model"
	"github.com/stretchr/testify/require"
)

const pruneTwoEnums = `version: 1
package: sample
enums:
  - name: Color
    type: string
    values: [red, green]
  - name: Status
    type: int
    values: [pending, done]
`

const pruneOneEnum = `version: 1
package: sample
enums:
  - name: Color
    type: string
    values: [red, green]
`

func loadAt(t *testing.T, dir, content string) *model.Manifest {
	t.Helper()
	path := filepath.Join(dir, "enumer.yaml")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	m, diags := manifest.Load(path)
	require.False(t, diags.HasErrors(), diags.Format())
	return m
}

func TestPruneRemovedEnums(t *testing.T) {
	dir := t.TempDir()
	m := loadAt(t, dir, pruneTwoEnums)
	_, diags := New().Generate(context.Background(), []*model.Manifest{m}, GenerateOptions{})
	require.False(t, diags.HasErrors(), diags.Format())

	colorPath := filepath.Join(dir, "color_enum.go")
	statusPath := filepath.Join(dir, "status_enum.go")
	require.FileExists(t, colorPath)
	require.FileExists(t, statusPath)

	m = loadAt(t, dir, pruneOneEnum)
	result, diags := New().Generate(context.Background(), []*model.Manifest{m}, GenerateOptions{})
	require.False(t, diags.HasErrors(), diags.Format())
	require.FileExists(t, colorPath)
	require.NoFileExists(t, statusPath)
	require.Contains(t, result.Orphans, statusPath)
	require.Contains(t, result.Removed, statusPath)
}

func TestNoPruneKeepsOrphans(t *testing.T) {
	dir := t.TempDir()
	m := loadAt(t, dir, pruneTwoEnums)
	_, diags := New().Generate(context.Background(), []*model.Manifest{m}, GenerateOptions{})
	require.False(t, diags.HasErrors(), diags.Format())
	statusPath := filepath.Join(dir, "status_enum.go")

	m = loadAt(t, dir, pruneOneEnum)
	result, diags := New().Generate(context.Background(), []*model.Manifest{m}, GenerateOptions{NoPrune: true})
	require.False(t, diags.HasErrors(), diags.Format())
	require.FileExists(t, statusPath)
	require.Contains(t, result.Orphans, statusPath)
	require.Empty(t, result.Removed)
}

func TestCheckReportsOrphans(t *testing.T) {
	dir := t.TempDir()
	m := loadAt(t, dir, pruneTwoEnums)
	_, diags := New().Generate(context.Background(), []*model.Manifest{m}, GenerateOptions{})
	require.False(t, diags.HasErrors(), diags.Format())
	statusPath := filepath.Join(dir, "status_enum.go")

	m = loadAt(t, dir, pruneOneEnum)
	result, diags := New().Check(context.Background(), []*model.Manifest{m}, CheckOptions{Diff: true})
	require.True(t, diags.HasErrors())
	require.Contains(t, result.Orphans, statusPath)
	require.Contains(t, result.Stale, statusPath)
	require.Len(t, result.Diffs, 1)
	require.Contains(t, diags.Format(), "no longer declared")
}

func TestPruneDryRunDoesNotDelete(t *testing.T) {
	dir := t.TempDir()
	m := loadAt(t, dir, pruneTwoEnums)
	_, diags := New().Generate(context.Background(), []*model.Manifest{m}, GenerateOptions{})
	require.False(t, diags.HasErrors(), diags.Format())
	statusPath := filepath.Join(dir, "status_enum.go")

	m = loadAt(t, dir, pruneOneEnum)
	result, diags := New().Generate(context.Background(), []*model.Manifest{m}, GenerateOptions{DryRun: true})
	require.False(t, diags.HasErrors(), diags.Format())
	require.FileExists(t, statusPath)
	require.Contains(t, result.Orphans, statusPath)
	require.Empty(t, result.Removed)
}

func TestPruneDoesNotRemoveOtherManifest(t *testing.T) {
	dir := t.TempDir()
	aPath := filepath.Join(dir, "a.yaml")
	bPath := filepath.Join(dir, "b.yaml")
	require.NoError(t, os.WriteFile(aPath, []byte(`version: 1
package: sample
enums:
  - name: A
    values: [a1]
  - name: A2
    values: [a2]
`), 0o644))
	require.NoError(t, os.WriteFile(bPath, []byte(`version: 1
package: sample
enums:
  - name: B
    values: [b1]
`), 0o644))

	mA, diags := manifest.Load(aPath)
	require.False(t, diags.HasErrors(), diags.Format())
	mB, diags := manifest.Load(bPath)
	require.False(t, diags.HasErrors(), diags.Format())
	_, diags = New().Generate(context.Background(), []*model.Manifest{mA, mB}, GenerateOptions{})
	require.False(t, diags.HasErrors(), diags.Format())

	bEnum := filepath.Join(dir, "b_enum.go")
	a2Enum := filepath.Join(dir, "a2_enum.go")
	require.FileExists(t, bEnum)
	require.FileExists(t, a2Enum)

	require.NoError(t, os.WriteFile(aPath, []byte(`version: 1
package: sample
enums:
  - name: A
    values: [a1]
`), 0o644))
	mA, diags = manifest.Load(aPath)
	require.False(t, diags.HasErrors(), diags.Format())
	result, diags := New().Generate(context.Background(), []*model.Manifest{mA}, GenerateOptions{})
	require.False(t, diags.HasErrors(), diags.Format())
	require.NoFileExists(t, a2Enum)
	require.FileExists(t, bEnum)
	require.Contains(t, result.Removed, a2Enum)
}

func TestPruneNestedOutput(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "generated"), 0o755))
	m := loadAt(t, dir, `version: 1
package: sample
enums:
  - name: Color
    values: [red]
  - name: Status
    type: int
    output: generated/status.go
    values: [pending]
`)
	_, diags := New().Generate(context.Background(), []*model.Manifest{m}, GenerateOptions{})
	require.False(t, diags.HasErrors(), diags.Format())
	nested := filepath.Join(dir, "generated", "status.go")
	require.FileExists(t, nested)

	m = loadAt(t, dir, pruneOneEnum)
	_, diags = New().Generate(context.Background(), []*model.Manifest{m}, GenerateOptions{})
	require.False(t, diags.HasErrors(), diags.Format())
	require.NoFileExists(t, nested)
}

func TestPruneKeepsHandwrittenFiles(t *testing.T) {
	dir := t.TempDir()
	m := loadAt(t, dir, pruneTwoEnums)
	_, diags := New().Generate(context.Background(), []*model.Manifest{m}, GenerateOptions{})
	require.False(t, diags.HasErrors(), diags.Format())

	notes := filepath.Join(dir, "notes.go")
	require.NoError(t, os.WriteFile(notes, []byte("package sample\n\n// notes.\n"), 0o644))
	m = loadAt(t, dir, pruneOneEnum)
	_, diags = New().Generate(context.Background(), []*model.Manifest{m}, GenerateOptions{})
	require.False(t, diags.HasErrors(), diags.Format())
	require.FileExists(t, notes)
}
