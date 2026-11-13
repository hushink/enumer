// Copyright (c) 2026 The enumer Authors
// SPDX-License-Identifier: MIT

// Package cli implements the enumer command line interface.
package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/hushink/enumer/generator"
	"github.com/hushink/enumer/internal/buildinfo"
	"github.com/hushink/enumer/internal/diag"
	"github.com/hushink/enumer/manifest"
	"github.com/hushink/enumer/model"
	"github.com/spf13/cobra"
)

type exitError struct {
	Code   int
	Err    error
	Silent bool
}

func (e *exitError) Error() string {
	if e.Err == nil {
		return ""
	}
	return e.Err.Error()
}

func (e *exitError) Unwrap() error { return e.Err }

// Execute runs the CLI and returns a process exit code.
func Execute(args []string, stdout, stderr io.Writer) int {
	root := NewRootCmd(stdout, stderr)
	root.SetArgs(args)
	err := root.Execute()
	if err == nil {
		return 0
	}
	var ee *exitError
	if errors.As(err, &ee) {
		if !ee.Silent && ee.Err != nil {
			fmt.Fprintln(stderr, ee.Err)
		}
		return ee.Code
	}
	fmt.Fprintln(stderr, err)
	return 2
}

type globalOptions struct {
	strict  bool
	json    bool
	verbose bool
	noPrune bool
}

// NewRootCmd builds the command tree.
func NewRootCmd(stdout, stderr io.Writer) *cobra.Command {
	g := &globalOptions{}
	root := &cobra.Command{
		Use:   "enumer",
		Short: "Generate type-safe Go enumerations from YAML manifests",
		Long: `enumer reads enumer.yaml manifests and generates type-safe Go enums
with String, Parse, JSON, SQL, flag and other opt-in capabilities.

Run "enumer" without a subcommand to generate from the current directory.`,
		Version:       buildinfo.String(),
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGenerate(cmd, args, g, stdout, stderr, false)
		},
	}
	root.SetVersionTemplate("enumer {{.Version}}\n")
	root.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		return &exitError{Code: 2, Err: err}
	})
	root.PersistentFlags().BoolVar(&g.strict, "strict", false, "treat warnings as errors")
	root.PersistentFlags().BoolVar(&g.json, "json", false, "print diagnostics as JSON")
	root.PersistentFlags().BoolVarP(&g.verbose, "verbose", "v", false, "print generated file details")
	root.PersistentFlags().BoolVar(&g.noPrune, "no-prune", false, "keep generated files for enums no longer declared")

	root.AddCommand(
		newGenerateCmd(g, stdout, stderr),
		newCheckCmd(g, stdout, stderr),
		newInitCmd(stdout),
		newVersionCmd(stdout),
	)
	return root
}

// hasManifest reports whether dir contains a manifest that no-argument
// discovery would use.
func hasManifest(dir string) bool {
	for _, name := range []string{"enumer.yaml", "enumer.yml"} {
		if info, err := os.Stat(filepath.Join(dir, name)); err == nil && !info.IsDir() {
			return true
		}
	}
	return false
}

func loadManifests(paths []string) ([]*model.Manifest, diag.List) {
	if len(paths) == 0 {
		if cwd, err := os.Getwd(); err == nil && !hasManifest(cwd) {
			if root, err := projectRoot(cwd); err == nil && root != cwd && hasManifest(root) {
				paths = []string{root}
			} else if ws, ok := workspaceRoot(cwd); ok && ws != cwd && hasManifest(ws) {
				paths = []string{ws}
			}
		}
	}
	files, diags := manifest.Discover(paths)
	if diags.HasErrors() {
		return nil, diags
	}
	var manifests []*model.Manifest
	for _, file := range files {
		m, mdiags := manifest.Load(file)
		diags = append(diags, mdiags...)
		if m != nil {
			manifests = append(manifests, m)
		}
	}
	diags.Sort()
	return manifests, diags
}

func hasEffectiveErrors(diags diag.List, g *globalOptions) bool {
	if g.strict {
		return diags.UpgradeWarnings().HasErrors()
	}
	return diags.HasErrors()
}

func reportDiagnostics(diags diag.List, g *globalOptions, stdout, stderr io.Writer) error {
	if g.strict {
		diags = diags.UpgradeWarnings()
	}
	if g.json {
		data, err := diags.JSON()
		if err != nil {
			return &exitError{Code: 1, Err: err}
		}
		fmt.Fprintln(stdout, string(data))
	} else {
		for _, d := range diags {
			fmt.Fprintln(stderr, d.String())
		}
	}
	if diags.HasErrors() {
		return &exitError{Code: 1, Silent: true}
	}
	return nil
}

func newGenerator() *generator.Generator {
	return generator.New()
}
