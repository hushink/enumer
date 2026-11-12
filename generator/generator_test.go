package generator

import (
	"context"
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/hushink/enumer/manifest"
	"github.com/hushink/enumer/model"
	"github.com/stretchr/testify/require"
)

var update = flag.Bool("update", false, "update golden files")

func loadOne(t *testing.T, content string) (*model.Manifest, string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "enumer.yaml")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	m, diags := manifest.Load(path)
	require.False(t, diags.HasErrors(), diags.Format())
	return m, dir
}

func TestGolden(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "basic", "enumer.yaml"))
	require.NoError(t, err)

	dir := t.TempDir()
	path := filepath.Join(dir, "enumer.yaml")
	require.NoError(t, os.WriteFile(path, data, 0o644))
	m, diags := manifest.Load(path)
	require.False(t, diags.HasErrors(), diags.Format())

	result, gdiags := New().Generate(context.Background(), []*model.Manifest{m}, GenerateOptions{})
	require.False(t, gdiags.HasErrors(), gdiags.Format())
	require.Len(t, result.Outputs, 2)

	for _, out := range result.Outputs {
		name := filepath.Base(out.Path)
		wantPath := filepath.Join("testdata", "basic", "expected", name)
		if *update {
			require.NoError(t, os.MkdirAll(filepath.Dir(wantPath), 0o755))
			require.NoError(t, os.WriteFile(wantPath, out.Content, 0o644))
			continue
		}
		want, err := os.ReadFile(wantPath)
		require.NoError(t, err)
		require.Equal(t, string(want), string(out.Content), "golden mismatch for %s", name)
	}
}

func TestGenerateCheckAndDryRun(t *testing.T) {
	m, dir := loadOne(t, `version: 1
package: sample
enums:
  - name: Color
    type: string
    values: [red, green]
`)
	g := New()

	result, diags := g.Generate(context.Background(), []*model.Manifest{m}, GenerateOptions{})
	require.False(t, diags.HasErrors(), diags.Format())
	require.Len(t, result.Changed, 1)

	colorPath := filepath.Join(dir, "color_enum.go")
	content, err := os.ReadFile(colorPath)
	require.NoError(t, err)
	require.Contains(t, string(content), "type Color string")

	check, cdiags := g.Check(context.Background(), []*model.Manifest{m}, CheckOptions{})
	require.False(t, cdiags.HasErrors(), cdiags.Format())
	require.Empty(t, check.Stale)

	require.NoError(t, os.WriteFile(colorPath, []byte("// drift\n"), 0o644))
	check, cdiags = g.Check(context.Background(), []*model.Manifest{m}, CheckOptions{Diff: true})
	require.True(t, cdiags.HasErrors())
	require.Len(t, check.Stale, 1)
	require.Len(t, check.Diffs, 1)

	require.NoError(t, os.Remove(colorPath))
	check, cdiags = g.Check(context.Background(), []*model.Manifest{m}, CheckOptions{})
	require.True(t, cdiags.HasErrors())
	require.Len(t, check.Stale, 1)
	require.Contains(t, cdiags.Format(), "missing")

	_, ddiags := g.Generate(context.Background(), []*model.Manifest{m}, GenerateOptions{DryRun: true})
	require.False(t, ddiags.HasErrors(), ddiags.Format())
	_, err = os.Stat(colorPath)
	require.True(t, os.IsNotExist(err))
}

func TestCustomTarget(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "templates"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "templates", "list.md.tmpl"), []byte("{{range .Enums}}{{.Name}}: {{len .Values}}\n{{end}}"), 0o644))
	path := filepath.Join(dir, "enumer.yaml")
	require.NoError(t, os.WriteFile(path, []byte(`version: 1
package: sample
enums:
  - name: Color
    values: [red, green]
targets:
  - template: ./templates/list.md.tmpl
    output: LIST.md
`), 0o644))

	m, diags := manifest.Load(path)
	require.False(t, diags.HasErrors(), diags.Format())
	_, gdiags := New().Generate(context.Background(), []*model.Manifest{m}, GenerateOptions{})
	require.False(t, gdiags.HasErrors(), gdiags.Format())

	content, err := os.ReadFile(filepath.Join(dir, "LIST.md"))
	require.NoError(t, err)
	require.Equal(t, "Color: 2\n", string(content))
}

func TestCustomTargetPerEnum(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "templates"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "templates", "one.tmpl"), []byte("{{.Enum.Name}}={{len .Enum.Values}}\n"), 0o644))
	path := filepath.Join(dir, "enumer.yaml")
	require.NoError(t, os.WriteFile(path, []byte(`version: 1
package: sample
enums:
  - name: Color
    values: [red, green]
  - name: Status
    type: int
    values: [a, b, c]
targets:
  - template: ./templates/one.tmpl
    output: "{{.Enum.Name}}.txt"
    perEnum: true
`), 0o644))

	m, diags := manifest.Load(path)
	require.False(t, diags.HasErrors(), diags.Format())
	_, gdiags := New().Generate(context.Background(), []*model.Manifest{m}, GenerateOptions{})
	require.False(t, gdiags.HasErrors(), gdiags.Format())

	require.FileExists(t, filepath.Join(dir, "Color.txt"))
	require.FileExists(t, filepath.Join(dir, "Status.txt"))
	content, err := os.ReadFile(filepath.Join(dir, "Status.txt"))
	require.NoError(t, err)
	require.Equal(t, "Status=3\n", string(content))
}

func TestAggregate(t *testing.T) {
	m, dir := loadOne(t, `version: 1
package: sample
aggregate: true
enums:
  - name: A
    values: [x]
  - name: B
    type: int
    values: [y]
`)
	result, diags := New().Generate(context.Background(), []*model.Manifest{m}, GenerateOptions{})
	require.False(t, diags.HasErrors(), diags.Format())
	require.Len(t, result.Outputs, 1)

	content, err := os.ReadFile(filepath.Join(dir, "enumer_enum.go"))
	require.NoError(t, err)
	require.Contains(t, string(content), "type A string")
	require.Contains(t, string(content), "type B int")
}

func TestCommentsNone(t *testing.T) {
	m, dir := loadOne(t, `version: 1
package: sample
defaults:
  comments: none
enums:
  - name: Color
    doc: Color is documented.
    values: [red]
`)
	_, diags := New().Generate(context.Background(), []*model.Manifest{m}, GenerateOptions{})
	require.False(t, diags.HasErrors(), diags.Format())

	content, err := os.ReadFile(filepath.Join(dir, "color_enum.go"))
	require.NoError(t, err)
	require.Contains(t, string(content), "// Color is documented.")
	require.NotContains(t, string(content), "is an enumeration")
}

func TestFeatureError(t *testing.T) {
	m, _ := loadOne(t, `version: 1
package: sample
enums:
  - name: A
    type: string
    features: [json:number]
    values: [x]
`)
	_, diags := New().Generate(context.Background(), []*model.Manifest{m}, GenerateOptions{})
	require.True(t, diags.HasErrors())
	require.Contains(t, diags.Format(), "json:number requires a numeric enum")
}

func TestSymbolCollision(t *testing.T) {
	m, _ := loadOne(t, `version: 1
package: sample
enums:
  - name: A
    values: [x]
  - name: A
    values: [y]
`)
	_, diags := New().Generate(context.Background(), []*model.Manifest{m}, GenerateOptions{})
	require.True(t, diags.HasErrors())
	require.Contains(t, diags.Format(), "conflicts with enum")
}

func TestSymbolsAreScopedToDirectories(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()
	content := `version: 1
package: sample
enums:
  - name: Color
    values: [red, green]
`
	path1 := filepath.Join(dir1, "enumer.yaml")
	path2 := filepath.Join(dir2, "enumer.yaml")
	require.NoError(t, os.WriteFile(path1, []byte(content), 0o644))
	require.NoError(t, os.WriteFile(path2, []byte(content), 0o644))

	m1, diags1 := manifest.Load(path1)
	require.False(t, diags1.HasErrors(), diags1.Format())
	m2, diags2 := manifest.Load(path2)
	require.False(t, diags2.HasErrors(), diags2.Format())

	_, gdiags := New().Generate(context.Background(), []*model.Manifest{m1, m2}, GenerateOptions{})
	require.False(t, gdiags.HasErrors(), gdiags.Format())
	require.FileExists(t, filepath.Join(dir1, "color_enum.go"))
	require.FileExists(t, filepath.Join(dir2, "color_enum.go"))
}

func TestPackageMismatchInDirectory(t *testing.T) {
	dir := t.TempDir()
	path1 := filepath.Join(dir, "a.yaml")
	path2 := filepath.Join(dir, "b.yaml")
	require.NoError(t, os.WriteFile(path1, []byte(`version: 1
package: p
enums:
  - name: A
    values: [x]
`), 0o644))
	require.NoError(t, os.WriteFile(path2, []byte(`version: 1
package: q
enums:
  - name: B
    values: [y]
`), 0o644))

	m1, diags1 := manifest.Load(path1)
	require.False(t, diags1.HasErrors(), diags1.Format())
	m2, diags2 := manifest.Load(path2)
	require.False(t, diags2.HasErrors(), diags2.Format())

	_, gdiags := New().Generate(context.Background(), []*model.Manifest{m1, m2}, GenerateOptions{})
	require.True(t, gdiags.HasErrors())
	require.Contains(t, gdiags.Format(), "contains both package")
}

func TestOutputCollision(t *testing.T) {
	m, _ := loadOne(t, `version: 1
package: sample
enums:
  - name: A
    output: same.go
    values: [x]
  - name: B
    output: same.go
    values: [y]
`)
	_, diags := New().Generate(context.Background(), []*model.Manifest{m}, GenerateOptions{})
	require.True(t, diags.HasErrors())
	require.Contains(t, diags.Format(), "conflicts with")
}

func TestGeneratedCodeCompilesAndRuns(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compile test in short mode")
	}
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not available")
	}

	m, dir := loadOne(t, `version: 1
package: sample
defaults:
  features: [string, parse, valid, text, mustparse, names, values, ptr]
enums:
  - name: Color
    type: string
    features: [json, flag]
    values:
      - red
      - name: green
        value: GREEN
        aliases: [g]
  - name: Status
    type: int32
    features: [json:number, sql, sql:null, flag]
    values:
      - pending
      - running
      - name: failed
        value: 10
  - name: ErrorCode
    type: string
    features: [sql:int]
    values:
      - name: not_found
        number: 404
      - name: internal_error
        number: 500
  - name: Priority
    type: int
    allowDuplicateValues: true
    features: [json:number]
    values:
      - name: high
        value: 1
      - name: urgent
        value: 1
`)
	_, diags := New().Generate(context.Background(), []*model.Manifest{m}, GenerateOptions{})
	require.False(t, diags.HasErrors(), diags.Format())

	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module sample\n\ngo 1.20\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "enums_test.go"), []byte(behaviorTest), 0o644))

	cmd := exec.Command("go", "test", "./...")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))
}

const behaviorTest = `package sample

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestColorBehavior(t *testing.T) {
	if ColorRed.String() != "red" || ColorGreen.String() != "GREEN" {
		t.Fatal("String")
	}
	c, err := ParseColor("g")
	if err != nil || c != ColorGreen {
		t.Fatal("Parse alias")
	}
	if _, err := ParseColor("nope"); !errors.Is(err, ErrInvalidColor) {
		t.Fatal("sentinel")
	}
	if !ColorGreen.IsValid() || Color("nope").IsValid() {
		t.Fatal("IsValid")
	}
	b, err := json.Marshal(ColorGreen)
	if err != nil || string(b) != "\"GREEN\"" {
		t.Fatalf("MarshalJSON: %s %v", b, err)
	}
	var c2 Color
	if err := json.Unmarshal([]byte("\"red\""), &c2); err != nil || c2 != ColorRed {
		t.Fatal("UnmarshalJSON")
	}
	if len(ColorNames()) != 2 || len(ColorValues()) != 2 {
		t.Fatal("Names/Values")
	}
}

func TestStatusBehavior(t *testing.T) {
	if StatusFailed.String() != "failed" || Status(5).String() != "Status(5)" {
		t.Fatal("String")
	}
	if !StatusFailed.IsValid() || Status(5).IsValid() {
		t.Fatal("IsValid")
	}
	b, err := json.Marshal(StatusFailed)
	if err != nil || string(b) != "10" {
		t.Fatalf("MarshalJSON: %s %v", b, err)
	}
	var s Status
	if err := json.Unmarshal([]byte("10"), &s); err != nil || s != StatusFailed {
		t.Fatal("UnmarshalJSON")
	}
	if err := json.Unmarshal([]byte("5"), &s); err == nil {
		t.Fatal("expected invalid JSON number")
	}
	var n NullStatus
	if err := n.Scan(nil); err != nil || n.Valid {
		t.Fatal("null Scan")
	}
	if err := n.Scan(int64(10)); err != nil || !n.Valid || n.Enum != StatusFailed {
		t.Fatal("Scan")
	}
	v, err := n.Value()
	if err != nil || v.(int64) != 10 {
		t.Fatal("Null Value")
	}
	jsonData, jsonErr := json.Marshal(n)
	if jsonErr != nil || string(jsonData) != "10" {
		t.Fatalf("Null MarshalJSON: %s %v", jsonData, jsonErr)
	}
	var n2 NullStatus
	if err := json.Unmarshal([]byte("null"), &n2); err != nil || n2.Valid {
		t.Fatal("Null UnmarshalJSON null")
	}
	if err := json.Unmarshal([]byte("10"), &n2); err != nil || !n2.Valid || n2.Enum != StatusFailed {
		t.Fatal("Null UnmarshalJSON value")
	}
}

func TestFeatureBehaviors(t *testing.T) {
	if MustParseColor("red") != ColorRed {
		t.Fatal("MustParse")
	}
	var c Color
	if err := c.Set("red"); err != nil || c != ColorRed || c.Type() != "color" {
		t.Fatal("flag")
	}
	text, err := ColorRed.MarshalText()
	if err != nil || string(text) != "red" {
		t.Fatal("MarshalText")
	}
	if err := c.UnmarshalText([]byte("green")); err == nil {
		t.Fatal("expected case-sensitive failure")
	}
	if err := c.UnmarshalText([]byte("GREEN")); err != nil || c != ColorGreen {
		t.Fatal("UnmarshalText")
	}
}

func TestMustParsePanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	MustParseColor("nope")
}

func TestSQLIntBehavior(t *testing.T) {
	var code ErrorCode
	if err := code.Scan(int64(404)); err != nil || code != ErrorCodeNotFound {
		t.Fatal("Scan")
	}
	if err := code.Scan(int64(999)); err == nil {
		t.Fatal("expected unknown number error")
	}
	v, err := ErrorCodeNotFound.Value()
	if err != nil || v.(int64) != 404 {
		t.Fatal("Value")
	}
}

func TestDuplicateValues(t *testing.T) {
	if PriorityHigh != PriorityUrgent {
		t.Fatal("expected equal values")
	}
	if PriorityHigh.String() != "high" {
		t.Fatalf("first wins: %q", PriorityHigh.String())
	}
	if !PriorityUrgent.IsValid() {
		t.Fatal("IsValid")
	}
	b, err := json.Marshal(PriorityUrgent)
	if err != nil || string(b) != "1" {
		t.Fatalf("json: %s %v", b, err)
	}
}
`
