package mysql

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newUDBLog ucloud udb log
func newUDBLog(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs",
		Short: "List and manipulate logs of MySQL instance",
		Long:  "List and manipulate logs of MySQL instance",
	}

	cmd.AddCommand(newUDBLogArchiveCreate(ctx))
	cmd.AddCommand(newUDBLogArchiveList(ctx))
	cmd.AddCommand(newUDBLogArchiveGetDownloadURL(ctx))
	cmd.AddCommand(newUDBLogArchiveDelete(ctx))

	return cmd
}
