package umodelverse

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

func newModel(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "model",
		Short: "Query uModelVerse models",
		Long:  "Query uModelVerse models.",
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(newModelList(ctx))
	return cmd
}
