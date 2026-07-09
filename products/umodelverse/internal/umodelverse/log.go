package umodelverse

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

func newLog(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "log",
		Short: "Query and export uModelVerse inference logs",
		Long:  "Query and export uModelVerse inference logs.",
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(newLogList(ctx))
	cmd.AddCommand(newLogDescribe(ctx))
	cmd.AddCommand(newLogExport(ctx))
	return cmd
}
