package manifest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func writeManifest(t *testing.T, dir, content string) string {
	t.Helper()
	path := filepath.Join(dir, "enumer.yaml")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	return path
}

func TestLoadBasic(t *testing.T) {
	dir := t.TempDir()
	path := writeManifest(t, dir, `version: 1
package: sample

defaults:
  features: [string, parse, valid, text]

enums:
  - name: Color
    type: string
    values:
      - red
      - name: green
        value: GREEN
        aliases: [g]
  - name: Status
    type: int32
    start: 5
    values:
      - pending
      - name: running
        value: 10
      - failed
`)

	m, diags := Load(path)
	require.False(t, diags.HasErrors(), diags.Format())
	require.Equal(t, "sample", m.Package)
	require.Len(t, m.Enums, 2)

	color := m.Enums[0]
	require.Equal(t, "Color", color.Name)
	require.True(t, color.Type.IsString())
	require.Len(t, color.Values, 2)
	require.Equal(t, "ColorRed", color.Values[0].ConstName)
	require.Equal(t, "red", color.Values[0].Text)
	require.Equal(t, "ColorGreen", color.Values[1].ConstName)
	require.Equal(t, "GREEN", color.Values[1].Text)
	require.Equal(t, []string{"g"}, color.Values[1].Aliases)
	require.True(t, strings.HasSuffix(color.Output, "color_enum.go"))

	status := m.Enums[1]
	require.Len(t, status.Values, 3)
	require.Equal(t, "5", status.Values[0].Number)
	require.Equal(t, "10", status.Values[1].Number)
	require.Equal(t, "11", status.Values[2].Number)
	require.Equal(t, "StatusFailed", status.Values[2].ConstName)
	require.True(t, strings.HasSuffix(status.Output, "status_enum.go"))
}

func TestLoadPackageInference(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "doc.go"), []byte("package inferred\n"), 0o644))
	path := writeManifest(t, dir, `version: 1
enums:
  - name: A
    values: [x]
`)
	m, diags := Load(path)
	require.False(t, diags.HasErrors(), diags.Format())
	require.Equal(t, "inferred", m.Package)
}

func TestLoadPackageMismatch(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "doc.go"), []byte("package foo\n"), 0o644))
	path := writeManifest(t, dir, `version: 1
package: bar
enums:
  - name: A
    values: [x]
`)
	_, diags := Load(path)
	require.True(t, diags.HasErrors())
	require.Contains(t, diags.Format(), "does not match existing package")
}

func TestLoadUnknownField(t *testing.T) {
	dir := t.TempDir()
	path := writeManifest(t, dir, `version: 1
package: p
bogus: true
enums:
  - name: A
    values: [x]
`)
	_, diags := Load(path)
	require.True(t, diags.HasErrors())
	require.Contains(t, diags.Format(), "bogus")
	require.Equal(t, "unknown-field", diags[0].Code)
}

func TestLoadDuplicateText(t *testing.T) {
	dir := t.TempDir()
	path := writeManifest(t, dir, `version: 1
package: p
enums:
  - name: A
    type: string
    values:
      - name: one
        value: same
      - name: two
        value: same
`)
	_, diags := Load(path)
	require.True(t, diags.HasErrors())
	require.Contains(t, diags.Format(), "duplicate text")
}

func TestLoadCaseInsensitiveDuplicateText(t *testing.T) {
	dir := t.TempDir()
	path := writeManifest(t, dir, `version: 1
package: p
defaults:
  parse:
    caseSensitive: false
enums:
  - name: A
    type: string
    values:
      - name: one
        value: Red
      - name: two
        value: RED
`)
	_, diags := Load(path)
	require.True(t, diags.HasErrors())
	require.Contains(t, diags.Format(), "duplicate text")
}

func TestLoadFeatureMerge(t *testing.T) {
	dir := t.TempDir()
	path := writeManifest(t, dir, `version: 1
package: p
defaults:
  features: [string, parse, valid, text]
enums:
  - name: A
    features: [sql]
    disableFeatures: [text]
    values: [x]
`)
	m, diags := Load(path)
	require.False(t, diags.HasErrors(), diags.Format())
	e := m.Enums[0]
	require.True(t, e.HasFeature("sql"))
	require.False(t, e.HasFeature("text"))
	require.True(t, e.HasFeature("string"))
	require.True(t, e.HasFeature("parse"))
	require.True(t, e.HasFeature("valid"))
}

func TestLoadSQLIntNumbers(t *testing.T) {
	dir := t.TempDir()
	path := writeManifest(t, dir, `version: 1
package: p
enums:
  - name: A
    type: string
    features: [sql:int]
    values:
      - name: one
        number: 10
      - name: two
        number: 20
`)
	m, diags := Load(path)
	require.False(t, diags.HasErrors(), diags.Format())
	require.True(t, m.Enums[0].Values[0].HasSQLNumber)
	require.Equal(t, "10", m.Enums[0].Values[0].SQLNumber)
	require.Equal(t, "20", m.Enums[0].Values[1].SQLNumber)
}

func TestLoadSQLIntMissingNumber(t *testing.T) {
	dir := t.TempDir()
	path := writeManifest(t, dir, `version: 1
package: p
enums:
  - name: A
    type: string
    features: [sql:int]
    values:
      - one
`)
	_, diags := Load(path)
	require.True(t, diags.HasErrors())
	require.Contains(t, diags.Format(), "must declare number")
}

func TestLoadDuplicateNumericValues(t *testing.T) {
	dir := t.TempDir()
	base := `version: 1
package: p
enums:
  - name: A
    type: int
    %s
    values:
      - name: one
        value: 1
      - name: two
        value: 1
`
	path := writeManifest(t, dir, strings.Replace(base, "%s", "", 1))
	_, diags := Load(path)
	require.True(t, diags.HasErrors())
	require.Contains(t, diags.Format(), "duplicate numeric value")

	dir2 := t.TempDir()
	path2 := writeManifest(t, dir2, strings.Replace(base, "%s", "allowDuplicateValues: true", 1))
	m, diags := Load(path2)
	require.False(t, diags.HasErrors(), diags.Format())
	require.Len(t, m.Enums[0].Values, 2)
}

func TestLoadAggregate(t *testing.T) {
	dir := t.TempDir()
	path := writeManifest(t, dir, `version: 1
package: p
aggregate: true
enums:
  - name: A
    values: [x]
  - name: B
    values: [y]
`)
	m, diags := Load(path)
	require.False(t, diags.HasErrors(), diags.Format())
	require.True(t, m.Aggregate)
	require.True(t, strings.HasSuffix(m.Output, "enumer_enum.go"))
}

func TestLoadUnsupportedVersion(t *testing.T) {
	dir := t.TempDir()
	path := writeManifest(t, dir, `version: 2
package: p
enums:
  - name: A
    values: [x]
`)
	_, diags := Load(path)
	require.True(t, diags.HasErrors())
	require.Contains(t, diags.Format(), "unsupported manifest version")
}

func TestLoadDeprecated(t *testing.T) {
	dir := t.TempDir()
	path := writeManifest(t, dir, `version: 1
package: p
enums:
  - name: A
    values:
      - name: old
        deprecated: true
      - name: older
        deprecated: use old
`)
	m, diags := Load(path)
	require.False(t, diags.HasErrors(), diags.Format())
	require.Equal(t, "Do not use.", m.Enums[0].Values[0].Deprecated)
	require.Equal(t, "use old", m.Enums[0].Values[1].Deprecated)
}

func TestLoadInitialisms(t *testing.T) {
	dir := t.TempDir()
	path := writeManifest(t, dir, `version: 1
package: p
defaults:
  naming:
    extraInitialisms: [grpc]
enums:
  - name: A
    values: [http_status, grpc_status]
`)
	m, diags := Load(path)
	require.False(t, diags.HasErrors(), diags.Format())
	require.Equal(t, "AHTTPStatus", m.Enums[0].Values[0].ConstName)
	require.Equal(t, "AGRPCStatus", m.Enums[0].Values[1].ConstName)
}

func TestLoadRawIdentifier(t *testing.T) {
	dir := t.TempDir()
	path := writeManifest(t, dir, `version: 1
package: p
defaults:
  naming:
    identifier: raw
enums:
  - name: A
    values: [Red, Blue]
`)
	m, diags := Load(path)
	require.False(t, diags.HasErrors(), diags.Format())
	require.Equal(t, "ARed", m.Enums[0].Values[0].ConstName)
	require.Equal(t, "ABlue", m.Enums[0].Values[1].ConstName)
}

func TestLoadRawIdentifierInvalid(t *testing.T) {
	dir := t.TempDir()
	path := writeManifest(t, dir, `version: 1
package: p
defaults:
  naming:
    identifier: raw
enums:
  - name: A
    values: [blue-green]
`)
	_, diags := Load(path)
	require.True(t, diags.HasErrors())
	require.Contains(t, diags.Format(), "not a valid Go constant name")
}

func TestLoadPrefixModes(t *testing.T) {
	dir := t.TempDir()
	path := writeManifest(t, dir, `version: 1
package: p
defaults:
  naming:
    prefix: none
enums:
  - name: A
    values: [red]
`)
	m, diags := Load(path)
	require.False(t, diags.HasErrors(), diags.Format())
	require.Equal(t, "Red", m.Enums[0].Values[0].ConstName)

	dir2 := t.TempDir()
	path2 := writeManifest(t, dir2, `version: 1
package: p
defaults:
  naming:
    prefix: literal:X_
enums:
  - name: A
    values: [red]
`)
	m2, diags2 := Load(path2)
	require.False(t, diags2.HasErrors(), diags2.Format())
	require.Equal(t, "X_Red", m2.Enums[0].Values[0].ConstName)
}

func TestLoadNoManifestPackageWithEnumPackages(t *testing.T) {
	dir := t.TempDir()
	path := writeManifest(t, dir, `version: 1
enums:
  - name: A
    package: foo
    output: foo/a_enum.go
    values: [x]
  - name: B
    package: bar
    output: bar/b_enum.go
    values: [y]
`)
	m, diags := Load(path)
	require.False(t, diags.HasErrors(), diags.Format())
	require.Equal(t, "foo", m.Package)
	require.Equal(t, "foo", m.Enums[0].Package)
	require.Equal(t, "bar", m.Enums[1].Package)
}

func TestLoadAggregateRejectsPerEnumOutput(t *testing.T) {
	dir := t.TempDir()
	path := writeManifest(t, dir, `version: 1
package: p
aggregate: true
enums:
  - name: A
    output: a.go
    values: [x]
`)
	_, diags := Load(path)
	require.True(t, diags.HasErrors())
	require.Contains(t, diags.Format(), "cannot be set when aggregate is true")
}

func TestLoadOutputOverride(t *testing.T) {
	dir := t.TempDir()
	path := writeManifest(t, dir, `version: 1
package: p
enums:
  - name: A
    output: custom/a_enum.go
    values: [x]
`)
	m, diags := Load(path)
	require.False(t, diags.HasErrors(), diags.Format())
	require.True(t, strings.HasSuffix(filepath.ToSlash(m.Enums[0].Output), "custom/a_enum.go"))
}

func TestLoadRangeErrors(t *testing.T) {
	dir := t.TempDir()
	path := writeManifest(t, dir, `version: 1
package: p
enums:
  - name: A
    type: int8
    start: 200
    values: [x]
`)
	_, diags := Load(path)
	require.True(t, diags.HasErrors())
	require.Contains(t, diags.Format(), "does not fit in int8")
}

func TestLoadAutoIncrementOverflow(t *testing.T) {
	dir := t.TempDir()
	path := writeManifest(t, dir, `version: 1
package: p
enums:
  - name: A
    type: int8
    values:
      - name: max
        value: 127
      - name: overflow
`)
	_, diags := Load(path)
	require.True(t, diags.HasErrors())
	require.Contains(t, diags.Format(), "auto-increment overflows")
}
