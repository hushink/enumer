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

	"github.com/spf13/cobra"
)

const initManifest = `version: 1

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
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create a sample enumer.yaml in the current directory",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if pkg != "" && (!token.IsIdentifier(pkg) || token.IsKeyword(pkg)) {
				return &exitError{Code: 2, Err: fmt.Errorf("invalid package name %q", pkg)}
			}

			manifestPath := "enumer.yaml"
			if _, err := os.Stat(manifestPath); err == nil && !force {
				return &exitError{Code: 1, Err: fmt.Errorf("%s already exists; use --force to overwrite", manifestPath)}
			}
			content := initManifest
			if pkg != "" {
				content = strings.Replace(initManifest, "version: 1\n", "version: 1\npackage: "+pkg+"\n", 1)
			}
			if err := os.WriteFile(manifestPath, []byte(content), 0o644); err != nil {
				return &exitError{Code: 1, Err: err}
			}
			fmt.Fprintf(stdout, "created %s\n", manifestPath)

			templatesDir := "templates"
			if err := os.MkdirAll(templatesDir, 0o755); err != nil {
				return &exitError{Code: 1, Err: err}
			}
			readmePath := filepath.Join(templatesDir, "README.md")
			if _, err := os.Stat(readmePath); os.IsNotExist(err) || force {
				if err := os.WriteFile(readmePath, []byte(initTemplatesReadme), 0o644); err != nil {
					return &exitError{Code: 1, Err: err}
				}
				fmt.Fprintf(stdout, "created %s\n", readmePath)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&pkg, "package", "", "Go package name to write into the manifest")
	cmd.Flags().BoolVar(&force, "force", false, "overwrite existing files")
	return cmd
}
