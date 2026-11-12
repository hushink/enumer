// Package sink writes generated files atomically and avoids rewriting
// unchanged content.
package sink

import (
	"bytes"
	"os"
	"path/filepath"
)

// Write writes data to path. It returns changed=false when the file already
// has identical content, preserving its modification time. Parent directories
// are created as needed.
func Write(path string, data []byte) (changed bool, err error) {
	old, readErr := os.ReadFile(path)
	if readErr == nil && bytes.Equal(old, data) {
		return false, nil
	}
	if readErr != nil && !os.IsNotExist(readErr) {
		return false, readErr
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return false, err
	}
	f, err := os.CreateTemp(dir, ".enumer-*.tmp")
	if err != nil {
		return false, err
	}
	tmp := f.Name()
	defer func() { _ = os.Remove(tmp) }()

	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		return false, err
	}
	if err := f.Chmod(0o644); err != nil {
		_ = f.Close()
		return false, err
	}
	if err := f.Close(); err != nil {
		return false, err
	}
	if err := os.Rename(tmp, path); err != nil {
		return false, err
	}
	return true, nil
}
