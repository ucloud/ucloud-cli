package pgsql

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newPgsqlDB ucloud pgsql db
func newPgsqlDB(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "db",
		Short: "Manage UPgSQL instances",
		Long:  "Manage UPgSQL instances",
	}

	cmd.AddCommand(newList(ctx))
	cmd.AddCommand(newGet(ctx))
	cmd.AddCommand(newCreate(ctx))
	cmd.AddCommand(newDelete(ctx))
	cmd.AddCommand(newStart(ctx))
	cmd.AddCommand(newStop(ctx))
	cmd.AddCommand(newRestart(ctx))
	cmd.AddCommand(newUpgrade(ctx))
	cmd.AddCommand(newListMachineType(ctx))
	cmd.AddCommand(newListVersion(ctx))
	cmd.AddCommand(newCreateReadonly(ctx))
	cmd.AddCommand(newStopCreatingReadonly(ctx))
	cmd.AddCommand(newUpdateName(ctx))
	cmd.AddCommand(newUpdateRemark(ctx))
	cmd.AddCommand(newResetPassword(ctx))
	cmd.AddCommand(newPrice(ctx))
	cmd.AddCommand(newUpgradePrice(ctx))

	return cmd
}
