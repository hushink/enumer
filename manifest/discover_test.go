// Copyright (c) 2026 The enumer Authors
// SPDX-License-Identifier: MIT

package manifest

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func writeDiscoverManifest(t *testing.T, root, rel string) string {
	t.Helper()
	path := filepath.Join(root, rel)
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte("version: 1\npackage: p\nenums:\n  - name: A\n    values: [x]\n"), 0o644))
	return path
}

func TestDiscoverRecursive(t *testing.T) {
	root := t.TempDir()
	rootManifest := writeDiscoverManifest(t, root, "enumer.yaml")
	subManifest := writeDiscoverManifest(t, root, filepath.Join("sub", "enumer.yaml"))
	writeDiscoverManifest(t, root, filepath.Join("vendor", "v", "enumer.yaml"))
	writeDiscoverManifest(t, root, filepath.Join("testdata", "t", "enumer.yaml"))
	writeDiscoverManifest(t, root, filepath.Join(".hidden", "enumer.yaml"))
	writeDiscoverManifest(t, root, filepath.Join("_skip", "enumer.yaml"))

	files, diags := Discover([]string{filepath.Join(root, "...")})
	require.False(t, diags.HasErrors(), diags.Format())
	require.ElementsMatch(t, []string{rootManifest, subManifest}, files)
}

func TestDiscoverDirectoryAndExplicitFile(t *testing.T) {
	root := t.TempDir()
	manifest := writeDiscoverManifest(t, root, "enumer.yaml")
	custom := writeDiscoverManifest(t, root, filepath.Join("custom", "anything.yaml"))

	files, diags := Discover([]string{root})
	require.False(t, diags.HasErrors(), diags.Format())
	require.Equal(t, []string{manifest}, files)

	files, diags = Discover([]string{custom})
	require.False(t, diags.HasErrors(), diags.Format())
	require.Equal(t, []string{custom}, files)
}

func TestDiscoverNoManifests(t *testing.T) {
	root := t.TempDir()
	_, diags := Discover([]string{filepath.Join(root, "...")})
	require.True(t, diags.HasErrors())
	require.Contains(t, diags.Format(), "no enumer.yaml manifests found")
}

func TestDiscoverBothExtensions(t *testing.T) {
	dir := t.TempDir()
	writeDiscoverManifest(t, dir, "enumer.yaml")
	writeDiscoverManifest(t, dir, "enumer.yml")

	_, diags := Discover([]string{dir})
	require.True(t, diags.HasErrors())
	require.Contains(t, diags.Format(), "both")
}

func TestDiscoverMissingPath(t *testing.T) {
	_, diags := Discover([]string{filepath.Join(t.TempDir(), "missing")})
	require.True(t, diags.HasErrors())
}
