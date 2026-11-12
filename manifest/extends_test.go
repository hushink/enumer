// Copyright (c) 2026 The enumer Authors
// SPDX-License-Identifier: MIT

package manifest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hushink/enumer/model"
	"github.com/stretchr/testify/require"
)

func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	return path
}

func TestLoadExtendsInheritsDefaults(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "base.yaml", `version: 1
defaults:
  features: [string, parse, valid, text, sql]
  naming:
    prefix: literal:X_
    identifier: raw
  parse:
    caseSensitive: false
  output:
    suffix: _generated.go
    dir: generated
  comments: none
  typeMode: existing
`)
	path := writeFile(t, dir, "enumer.yaml", `version: 1
package: p
extends: ./base.yaml
defaults:
  naming:
    prefix: type
enums:
  - name: A
    values: [red]
`)

	m, diags := Load(path)
	require.False(t, diags.HasErrors(), diags.Format())
	require.Len(t, m.Enums, 1)
	e := m.Enums[0]
	require.True(t, e.HasFeature("sql"))
	require.False(t, e.CaseSensitive)
	require.Equal(t, model.CommentsNone, e.Comments)
	require.Equal(t, model.TypeModeExisting, e.TypeMode)
	require.Equal(t, model.IdentifierRaw, e.Naming.Identifier)
	require.Equal(t, "Ared", e.Values[0].ConstName)
	require.True(t, strings.HasSuffix(filepath.ToSlash(e.Output), "generated/a_generated.go"))
}

func TestLoadExtendsReplacesFeatures(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "base.yaml", `version: 1
defaults:
  features: [string, parse, valid, text]
`)
	path := writeFile(t, dir, "enumer.yaml", `version: 1
package: p
extends: base.yaml
defaults:
  features: [sql]
enums:
  - name: A
    values: [x]
`)
	m, diags := Load(path)
	require.False(t, diags.HasErrors(), diags.Format())
	e := m.Enums[0]
	require.Equal(t, []string{"sql"}, e.Features)

	path2 := writeFile(t, dir, "enumer2.yaml", `version: 1
package: p
extends: base.yaml
defaults:
  features: []
enums:
  - name: B
    values: [y]
`)
	m2, diags2 := Load(path2)
	require.False(t, diags2.HasErrors(), diags2.Format())
	require.Empty(t, m2.Enums[0].Features)
}

func TestLoadExtendsCycle(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.yaml", `version: 1
extends: b.yaml
package: p
enums:
  - name: A
    values: [x]
`)
	writeFile(t, dir, "b.yaml", `version: 1
extends: a.yaml
defaults:
  features: [string]
`)
	_, diags := Load(filepath.Join(dir, "a.yaml"))
	require.True(t, diags.HasErrors())
	require.Contains(t, diags.Format(), "extends cycle")
}

func TestLoadExtendsMissing(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "enumer.yaml", `version: 1
package: p
extends: missing.yaml
enums:
  - name: A
    values: [x]
`)
	_, diags := Load(path)
	require.True(t, diags.HasErrors())
	require.Contains(t, diags.Format(), "missing.yaml")
}

func TestLoadExtendsInvalidBase(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "base.yaml", `version: 2
defaults:
  features: [string]
`)
	path := writeFile(t, dir, "enumer.yaml", `version: 1
package: p
extends: base.yaml
enums:
  - name: A
    values: [x]
`)
	_, diags := Load(path)
	require.True(t, diags.HasErrors())
	require.Contains(t, diags.Format(), "unsupported manifest version")

	dir2 := t.TempDir()
	writeFile(t, dir2, "base.yaml", `version: 1
bogus: true
`)
	path2 := writeFile(t, dir2, "enumer.yaml", `version: 1
package: p
extends: base.yaml
enums:
  - name: A
    values: [x]
`)
	_, diags2 := Load(path2)
	require.True(t, diags2.HasErrors())
	require.Contains(t, diags2.Format(), "bogus")
}

func TestLoadExtendsIgnoresBaseEnumsAndTargets(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "base.yaml", `version: 1
package: basepkg
aggregate: true
output: base.go
defaults:
  features: [string, parse, valid, text]
enums:
  - name: BaseEnum
    values: [a]
targets:
  - template: base.tmpl
    output: base.txt
`)
	path := writeFile(t, dir, "enumer.yaml", `version: 1
package: p
extends: base.yaml
enums:
  - name: ChildEnum
    values: [x]
`)
	m, diags := Load(path)
	require.False(t, diags.HasErrors(), diags.Format())
	require.Equal(t, "p", m.Package)
	require.False(t, m.Aggregate)
	require.Len(t, m.Enums, 1)
	require.Equal(t, "ChildEnum", m.Enums[0].Name)
	require.Empty(t, m.Targets)
	require.True(t, m.Enums[0].HasFeature("text"))
}
