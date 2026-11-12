// Copyright (c) 2026 The enumer Authors
// SPDX-License-Identifier: MIT

package manifest

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hushink/enumer/internal/diag"
)

const (
	manifestName    = "enumer.yaml"
	manifestNameAlt = "enumer.yml"
)

// Discover resolves command line paths to manifest files.
//
// A path may be a manifest file, a directory containing enumer.yaml, or a
// Go-style recursive pattern such as "./...".
func Discover(paths []string) ([]string, diag.List) {
	if len(paths) == 0 {
		paths = []string{"."}
	}

	var found []string
	var diags diag.List
	seen := map[string]bool{}

	for _, arg := range paths {
		recursive := strings.HasSuffix(arg, "...")
		root := arg
		if recursive {
			root = strings.TrimSuffix(arg, "...")
			root = strings.TrimRight(root, `/\`)
			if root == "" {
				root = "."
			}
		}

		info, err := os.Stat(root)
		if err != nil {
			diags = append(diags, errDiag("manifest-read", err.Error(), arg, 0, 0, "", ""))
			continue
		}

		if !info.IsDir() {
			if !isYAMLFile(root) {
				diags = append(diags, errDiag("manifest-read", fmt.Sprintf("%s is not a YAML manifest file", root), root, 0, 0, "", ""))
				continue
			}
			addFound(root, &found, seen)
			continue
		}

		if recursive {
			diags = append(diags, walkManifests(root, &found, seen)...)
			continue
		}

		files, err := manifestsInDir(root)
		if err != nil {
			diags = append(diags, errDiag("manifest-read", err.Error(), root, 0, 0, "", ""))
			continue
		}
		if len(files) == 0 {
			diags = append(diags, errDiag("manifest-read", fmt.Sprintf("no %s found in %s", manifestName, root), root, 0, 0, "", ""))
			continue
		}
		for _, f := range files {
			addFound(f, &found, seen)
		}
	}

	if len(found) == 0 && !diags.HasErrors() {
		diags = append(diags, errDiag("manifest-read", "no enumer.yaml manifests found", "", 0, 0, "", ""))
	}
	sort.Strings(found)
	return found, diags
}

func walkManifests(root string, found *[]string, seen map[string]bool) diag.List {
	var diags diag.List
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			diags = append(diags, errDiag("manifest-read", err.Error(), path, 0, 0, "", ""))
			return nil
		}
		if d.IsDir() {
			if path == root {
				return nil
			}
			name := d.Name()
			if name == "vendor" || name == "testdata" || strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") {
				return filepath.SkipDir
			}
			return nil
		}
		if !isManifestFile(path) {
			return nil
		}
		files, err := manifestsInDir(filepath.Dir(path))
		if err != nil {
			diags = append(diags, errDiag("manifest-read", err.Error(), filepath.Dir(path), 0, 0, "", ""))
			return nil
		}
		for _, f := range files {
			addFound(f, found, seen)
		}
		return nil
	})
	if err != nil {
		diags = append(diags, errDiag("manifest-read", err.Error(), root, 0, 0, "", ""))
	}
	return diags
}

func manifestsInDir(dir string) ([]string, error) {
	yamlPath := filepath.Join(dir, manifestName)
	ymlPath := filepath.Join(dir, manifestNameAlt)
	yamlOK := fileExists(yamlPath)
	ymlOK := fileExists(ymlPath)
	if yamlOK && ymlOK {
		return nil, fmt.Errorf("both %s and %s exist in %s", manifestName, manifestNameAlt, dir)
	}
	if yamlOK {
		return []string{yamlPath}, nil
	}
	if ymlOK {
		return []string{ymlPath}, nil
	}
	return nil, nil
}

func isManifestFile(path string) bool {
	base := strings.ToLower(filepath.Base(path))
	return base == manifestName || base == manifestNameAlt
}

func isYAMLFile(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".yaml", ".yml":
		return true
	default:
		return false
	}
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func addFound(path string, found *[]string, seen map[string]bool) {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = filepath.Clean(path)
	}
	if seen[abs] {
		return
	}
	seen[abs] = true
	*found = append(*found, filepath.Clean(path))
}
