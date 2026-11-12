// Copyright (c) 2026 The enumer Authors
// SPDX-License-Identifier: MIT

package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInitGenerateCheck(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	var stdout, stderr bytes.Buffer
	code := Execute([]string{"init", "--package", "sample"}, &stdout, &stderr)
	require.Equal(t, 0, code, stderr.String())
	require.FileExists(t, filepath.Join(dir, "enumer.yaml"))

	stdout.Reset()
	stderr.Reset()
	code = Execute([]string{"generate"}, &stdout, &stderr)
	require.Equal(t, 0, code, stderr.String())
	require.FileExists(t, filepath.Join(dir, "color_enum.go"))
	require.FileExists(t, filepath.Join(dir, "status_enum.go"))

	colorSource, err := os.ReadFile(filepath.Join(dir, "color_enum.go"))
	require.NoError(t, err)
	require.Contains(t, string(colorSource), "type Color string")
	require.Contains(t, string(colorSource), "ColorRed")
	require.Contains(t, string(colorSource), "func ColorNames() []string")
	require.Contains(t, string(colorSource), "func ColorValues() []Color")
	require.Contains(t, string(colorSource), "// Green is the second color.")

	stdout.Reset()
	stderr.Reset()
	code = Execute([]string{"check"}, &stdout, &stderr)
	require.Equal(t, 0, code, stderr.String())

	require.NoError(t, os.WriteFile(filepath.Join(dir, "color_enum.go"), []byte("// drift\n"), 0o644))
	stdout.Reset()
	stderr.Reset()
	code = Execute([]string{"check", "--diff"}, &stdout, &stderr)
	require.Equal(t, 1, code)
	require.Contains(t, stderr.String(), "out of date")
	require.Contains(t, stdout.String(), "generated")
}

func TestUsageExitCode(t *testing.T) {
	t.Chdir(t.TempDir())
	var stdout, stderr bytes.Buffer
	code := Execute([]string{"--badflag"}, &stdout, &stderr)
	require.Equal(t, 2, code)
}

func TestJSONDiagnostics(t *testing.T) {
	t.Chdir(t.TempDir())
	require.NoError(t, os.WriteFile("enumer.yaml", []byte("version: 1\npackage: p\nbogus: true\n"), 0o644))

	var stdout, stderr bytes.Buffer
	code := Execute([]string{"--json", "generate"}, &stdout, &stderr)
	require.Equal(t, 1, code)
	require.Contains(t, stdout.String(), `"code": "unknown-field"`)
}

func TestInitExistingFile(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	require.NoError(t, os.WriteFile("enumer.yaml", []byte("version: 1\n"), 0o644))

	var stdout, stderr bytes.Buffer
	code := Execute([]string{"init"}, &stdout, &stderr)
	require.Equal(t, 1, code)
	require.Contains(t, stderr.String(), "already exists")

	stdout.Reset()
	stderr.Reset()
	code = Execute([]string{"init", "--force", "--package", "sample"}, &stdout, &stderr)
	require.Equal(t, 0, code, stderr.String())
}

func TestDryRunDoesNotWrite(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	require.NoError(t, os.WriteFile("enumer.yaml", []byte(`version: 1
package: p
enums:
  - name: Color
    values: [red]
`), 0o644))

	var stdout, stderr bytes.Buffer
	code := Execute([]string{"generate", "--dry-run"}, &stdout, &stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Contains(t, stderr.String(), "would write")

	_, err := os.Stat("color_enum.go")
	require.True(t, os.IsNotExist(err))
}

func TestCheckJSONStale(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	require.NoError(t, os.WriteFile("enumer.yaml", []byte(`version: 1
package: p
enums:
  - name: Color
    values: [red]
`), 0o644))

	var stdout, stderr bytes.Buffer
	code := Execute([]string{"generate"}, &stdout, &stderr)
	require.Equal(t, 0, code, stderr.String())
	require.NoError(t, os.WriteFile("color_enum.go", []byte("// drift\n"), 0o644))

	stdout.Reset()
	stderr.Reset()
	code = Execute([]string{"--json", "check"}, &stdout, &stderr)
	require.Equal(t, 1, code)
	require.Contains(t, stdout.String(), `"code": "stale-output"`)
}

func TestStrictWarnings(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	require.NoError(t, os.WriteFile("enumer.yaml", []byte(`version: 1
package: p
enums:
  - name: A
    type: string
    values:
      - name: one
        number: 1
`), 0o644))

	var stdout, stderr bytes.Buffer
	code := Execute([]string{"generate"}, &stdout, &stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Contains(t, stderr.String(), "warning:")

	stdout.Reset()
	stderr.Reset()
	code = Execute([]string{"--strict", "generate"}, &stdout, &stderr)
	require.Equal(t, 1, code)
}
