package mysql

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newUDBConf ucloud udb conf
func newUDBConf(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "conf",
		Short: "List and manipulate configuration files of MySQL instances",
		Long:  "List and manipulate configuration files of MySQL instances",
	}
	cmd.AddCommand(newUDBConfList(ctx))
	cmd.AddCommand(newUDBConfDescribe(ctx))
	cmd.AddCommand(newUDBConfClone(ctx))
	cmd.AddCommand(newUDBConfUpload(ctx))
	cmd.AddCommand(newUDBConfUpdate(ctx))
	cmd.AddCommand(newUDBConfDelete(ctx))
	cmd.AddCommand(newUDBConfApply(ctx))
	cmd.AddCommand(newUDBConfDownload(ctx))
	return cmd
}
