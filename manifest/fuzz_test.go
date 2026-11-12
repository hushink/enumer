// Copyright (c) 2026 The enumer Authors
// SPDX-License-Identifier: MIT

package manifest

import (
	"os"
	"path/filepath"
	"testing"
)

func FuzzLoad(f *testing.F) {
	f.Add([]byte("version: 1\npackage: p\nenums:\n  - name: A\n    values: [x]\n"))
	f.Add([]byte("version: 1\npackage: p\nenums:\n  - name: A\n    type: int\n    values:\n      - name: one\n        value: 1\n"))
	f.Fuzz(func(t *testing.T, data []byte) {
		dir := t.TempDir()
		path := filepath.Join(dir, "enumer.yaml")
		if err := os.WriteFile(path, data, 0o644); err != nil {
			t.Fatal(err)
		}
		_, _ = Load(path)
	})
}
