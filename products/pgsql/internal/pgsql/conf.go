package pgsql

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newPgsqlConf ucloud pgsql conf
func newPgsqlConf(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "conf",
		Short: "List and manipulate parameter templates of UPgSQL instances",
		Long:  "List and manipulate parameter templates of UPgSQL instances",
	}
	cmd.AddCommand(newConfList(ctx))
	cmd.AddCommand(newConfDescribe(ctx))
	cmd.AddCommand(newConfCreate(ctx))
	cmd.AddCommand(newConfDelete(ctx))
	cmd.AddCommand(newConfDownload(ctx))
	cmd.AddCommand(newConfUpload(ctx))
	return cmd
}
