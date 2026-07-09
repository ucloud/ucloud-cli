package mysql

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// NewCommand builds the `mysql` root command and mounts the `db` subtree.
// Mirrors cmd/mysql.go NewCmdMysql + NewCmdMysqlDB.
func NewCommand(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mysql",
		Short: "Manipulate MySQL on UCloud platform",
		Long:  "Manipulate MySQL on UCloud platform",
	}
	cmd.AddCommand(newMysqlDB(ctx))
	cmd.AddCommand(newUDBConf(ctx))
	cmd.AddCommand(newUDBBackup(ctx))
	cmd.AddCommand(newUDBLog(ctx))
	return cmd
}
