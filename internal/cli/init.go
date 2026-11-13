// Copyright (c) 2026 The enumer Authors
// SPDX-License-Identifier: MIT

package cli

import (
	"fmt"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/hushink/enumer/schema"
	"github.com/spf13/cobra"
)

const initSchemaRef = "./" + schema.FileName

const initManifest = `# $schema: ` + initSchemaRef + `
$schema: ` + initSchemaRef + `
version: 1

defaults:
  features: [string, parse, valid, text, names, values]
  typeMode: auto
  comments: auto
  naming:
    prefix: type
    identifier: camel
  parse:
    caseSensitive: true
  output:
    suffix: _enum.go

enums:
  - name: Color
    type: string
    doc: Color is an example string enumeration.
    values:
      - red
      - name: green
        comment: Green is the second color.
      - blue

  - name: Status
    type: int
    doc: Status is an example numeric enumeration.
    values:
      - pending
      - running
      - completed
`

const initTemplatesReadme = `# Custom targets

Files in this directory can be referenced from the ` + "`targets`" + ` section of
` + "`enumer.yaml`" + ` to render extra files next to the Go output. For example:

` + "```yaml" + `
targets:
  - template: ./templates/enums.md.tmpl
    output: ENUMS.md
    perEnum: false
` + "```" + `

Template data:

- .Manifest: the normalized manifest
- .Enums: all enums in the manifest
- .Enum: the current enum when perEnum is true

Template functions: camel, snake, kebab, lower, upper, title, quote, join,
json, jsonPretty, default.
`

func newInitCmd(stdout io.Writer) *cobra.Command {
	var pkg string
	var force bool
	var local bool
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create a sample enumer.yaml in the module or workspace root (go.mod / go.work)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if pkg != "" && (!token.IsIdentifier(pkg) || token.IsKeyword(pkg)) {
				return &exitError{Code: 2, Err: fmt.Errorf("invalid package name %q", pkg)}
			}

			root, err := initDir(local)
			if err != nil {
				return &exitError{Code: 1, Err: err}
			}
			manifestPath := filepath.Join(root, "enumer.yaml")
			manifestDisplay := displayPath(manifestPath)
			if _, err := os.Stat(manifestPath); err == nil && !force {
				return &exitError{Code: 1, Err: fmt.Errorf("%s already exists; use --force to overwrite", manifestDisplay)}
			}
			content := initManifest
			if pkg != "" {
				content = strings.Replace(initManifest, "version: 1\n", "version: 1\npackage: "+pkg+"\n", 1)
			}
			if err := os.WriteFile(manifestPath, []byte(content), 0o644); err != nil {
				return &exitError{Code: 1, Err: err}
			}
			fmt.Fprintf(stdout, "created %s\n", manifestDisplay)

			schemaPath := filepath.Join(root, schema.FileName)
			if err := os.WriteFile(schemaPath, schema.Bytes(), 0o644); err != nil {
				return &exitError{Code: 1, Err: err}
			}
			fmt.Fprintf(stdout, "created %s\n", displayPath(schemaPath))

			templatesDir := filepath.Join(root, "templates")
			if err := os.MkdirAll(templatesDir, 0o755); err != nil {
				return &exitError{Code: 1, Err: err}
			}
			readmePath := filepath.Join(templatesDir, "README.md")
			if _, err := os.Stat(readmePath); os.IsNotExist(err) || force {
				if err := os.WriteFile(readmePath, []byte(initTemplatesReadme), 0o644); err != nil {
					return &exitError{Code: 1, Err: err}
				}
				fmt.Fprintf(stdout, "created %s\n", displayPath(readmePath))
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&pkg, "package", "", "Go package name to write into the manifest")
	cmd.Flags().BoolVar(&force, "force", false, "overwrite existing files")
	cmd.Flags().BoolVar(&local, "local", false, "create the manifest in the current directory instead of the project root")
	return cmd
}

// initDir returns the directory for init outputs: the nearest directory
// containing go.mod or go.work, or the current directory with --local or
// outside a module/workspace.
func initDir(local bool) (string, error) {
	if local {
		return filepath.Abs(".")
	}
	return projectRoot(".")
}

// projectRoot returns the nearest directory containing go.mod or go.work.
// go.mod wins when both files are in the same directory. If neither exists,
// it returns the starting directory.
func projectRoot(start string) (string, error) {
	if root, ok := findUp(start, "go.mod", "go.work"); ok {
		return root, nil
	}
	return filepath.Abs(start)
}

// workspaceRoot returns the nearest directory containing go.work.
func workspaceRoot(start string) (string, bool) {
	return findUp(start, "go.work")
}

func findUp(start string, names ...string) (string, bool) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", false
	}
	for {
		for _, name := range names {
			if info, err := os.Stat(filepath.Join(dir, name)); err == nil && !info.IsDir() {
				return dir, true
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

func displayPath(path string) string {
	cwd, err := os.Getwd()
	if err == nil {
		if rel, err := filepath.Rel(cwd, path); err == nil {
			return rel
		}
	}
	return path
}
