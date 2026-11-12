// Copyright (c) 2026 The enumer Authors
// SPDX-License-Identifier: MIT

package generator

import (
	"bufio"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hushink/enumer/model"
)

// findOrphans returns generated files that carry this manifest's exact
// generated header but are not part of the current plan.
//
// Hand-written files, files generated from other manifests and custom target
// outputs do not carry the matching header and are never returned.
func findOrphans(manifests []*model.Manifest, outputs []Output) []string {
	planned := make(map[string]bool, len(outputs))
	for _, out := range outputs {
		planned[pathKey(out.Path)] = true
	}

	orphans := map[string]bool{}
	for _, m := range manifests {
		if m == nil {
			continue
		}
		seenDirs := map[string]bool{}
		scan := func(path string) {
			if planned[pathKey(path)] || !strings.HasSuffix(strings.ToLower(path), ".go") {
				return
			}
			line, err := firstLine(path)
			if err != nil {
				return
			}
			if strings.TrimRight(line, "\r\n") == headerComment(m, path) {
				orphans[filepath.Clean(path)] = true
			}
		}

		_ = filepath.WalkDir(m.Dir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				if path == m.Dir {
					return nil
				}
				name := d.Name()
				if name == "vendor" || name == "testdata" || strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") {
					return filepath.SkipDir
				}
				return nil
			}
			scan(path)
			return nil
		})

		// Outputs may live outside the manifest tree. Scan the directories of
		// the current plan once; the header check keeps this safe.
		for _, out := range outputs {
			dir := filepath.Dir(out.Path)
			if withinDir(m.Dir, dir) {
				continue
			}
			key := pathKey(dir)
			if seenDirs[key] {
				continue
			}
			seenDirs[key] = true
			entries, err := os.ReadDir(dir)
			if err != nil {
				continue
			}
			for _, entry := range entries {
				if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".go") {
					continue
				}
				scan(filepath.Join(dir, entry.Name()))
			}
		}
	}

	list := make([]string, 0, len(orphans))
	for path := range orphans {
		list = append(list, path)
	}
	sort.Strings(list)
	return list
}

func firstLine(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()

	line, err := bufio.NewReader(f).ReadString('\n')
	if err != nil && line == "" {
		return "", err
	}
	return line, nil
}

func pathKey(path string) string {
	return strings.ToLower(filepath.Clean(path))
}

func withinDir(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	if rel == "." {
		return true
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
