package mysql

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newMysqlDB ucloud mysql db
func newMysqlDB(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "db",
		Short: "Manange MySQL instances",
		Long:  "Manange MySQL instances",
	}

	cmd.AddCommand(newList(ctx))
	cmd.AddCommand(newCreate(ctx))
	cmd.AddCommand(newDelete(ctx))
	cmd.AddCommand(newStart(ctx))
	cmd.AddCommand(newStop(ctx))
	cmd.AddCommand(newRestart(ctx))
	cmd.AddCommand(newResize(ctx))
	cmd.AddCommand(newRestore(ctx))
	cmd.AddCommand(newResetPassword(ctx))
	cmd.AddCommand(newCreateSlave(ctx))
	cmd.AddCommand(newPromoteSlave(ctx))
	cmd.AddCommand(newListMachineType(ctx))
	// cmd.AddCommand(newPromoteToHA(ctx))

	return cmd
}
