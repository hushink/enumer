package cli

import (
	"fmt"
	"io"

	"github.com/hushink/enumer/internal/buildinfo"
	"github.com/spf13/cobra"
)

func newVersionCmd(stdout io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the enumer version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(stdout, "enumer "+buildinfo.String())
			return nil
		},
	}
}
