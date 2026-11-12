package cli

import (
	"fmt"
	"io"

	"github.com/hushink/enumer/generator"
	"github.com/spf13/cobra"
)

func newGenerateCmd(g *globalOptions, stdout, stderr io.Writer) *cobra.Command {
	var dryRun bool
	cmd := &cobra.Command{
		Use:     "generate [paths...]",
		Aliases: []string{"gen"},
		Short:   "Generate Go files from manifests",
		Args:    cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGenerate(cmd, args, g, stdout, stderr, dryRun)
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print planned outputs without writing")
	return cmd
}

func runGenerate(cmd *cobra.Command, args []string, g *globalOptions, stdout, stderr io.Writer, dryRun bool) error {
	manifests, diags := loadManifests(args)
	if hasEffectiveErrors(diags, g) {
		return reportDiagnostics(diags, g, stdout, stderr)
	}

	result, gdiags := newGenerator().Generate(cmd.Context(), manifests, generator.GenerateOptions{DryRun: dryRun})
	diags = append(diags, gdiags...)
	if err := reportDiagnostics(diags, g, stdout, stderr); err != nil {
		return err
	}
	if result == nil {
		return &exitError{Code: 1, Silent: true}
	}

	if dryRun {
		for _, output := range result.Outputs {
			fmt.Fprintf(stderr, "would write %s\n", output.Path)
		}
		return nil
	}
	if g.verbose {
		for _, path := range result.Changed {
			fmt.Fprintf(stderr, "wrote %s\n", path)
		}
		for _, path := range result.Unchanged {
			fmt.Fprintf(stderr, "unchanged %s\n", path)
		}
	}
	return nil
}
