// Copyright (c) 2026 The enumer Authors
// SPDX-License-Identifier: MIT

package cli

import (
	"fmt"
	"io"

	"github.com/hushink/enumer/generator"
	"github.com/pmezard/go-difflib/difflib"
	"github.com/spf13/cobra"
)

func newCheckCmd(g *globalOptions, stdout, stderr io.Writer) *cobra.Command {
	var diff bool
	cmd := &cobra.Command{
		Use:   "check [paths...]",
		Short: "Verify that generated files are up to date",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			manifests, diags := loadManifests(args)
			if hasEffectiveErrors(diags, g) {
				return reportDiagnostics(diags, g, stdout, stderr)
			}

			result, cdiags := newGenerator().Check(cmd.Context(), manifests, generator.CheckOptions{Diff: diff})
			diags = append(diags, cdiags...)
			err := reportDiagnostics(diags, g, stdout, stderr)
			if err != nil {
				if diff && result != nil {
					diffOut := stdout
					if g.json {
						diffOut = stderr
					}
					printDiffs(diffOut, result.Diffs)
				}
				return err
			}
			if result != nil && len(result.Stale) == 0 {
				if g.verbose {
					fmt.Fprintf(stderr, "verified %d outputs\n", len(result.Outputs))
				}
				return nil
			}
			return &exitError{Code: 1, Silent: true}
		},
	}
	cmd.Flags().BoolVar(&diff, "diff", false, "print unified diffs for stale outputs")
	return cmd
}

func printDiffs(w io.Writer, diffs []generator.Diff) {
	for _, d := range diffs {
		text, err := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
			A:        difflib.SplitLines(string(d.Old)),
			B:        difflib.SplitLines(string(d.New)),
			FromFile: d.Path + " (on disk)",
			ToFile:   d.Path + " (generated)",
			Context:  3,
		})
		if err != nil {
			fmt.Fprintf(w, "diff %s: %v\n", d.Path, err)
			continue
		}
		fmt.Fprint(w, text)
	}
}
