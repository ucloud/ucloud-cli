package pgsql

import (
	"fmt"

	"github.com/spf13/cobra"

	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newUpgrade ucloud pgsql db upgrade
func newUpgrade(ctx *cli.Context) *cobra.Command {
	var async bool
	var idNames []string
	var machineType string
	var diskSpace int
	client := newUPgSQLClient(ctx)
	req := client.NewUpgradeUPgSQLInstanceRequest()
	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade disk space and/or machine type of UPgSQL instances",
		Long:  "Upgrade disk space and/or machine type of UPgSQL instances",
		Run: func(c *cobra.Command, args []string) {
			if machineType == "" && diskSpace == 0 {
				ctx.HandleError(fmt.Errorf("at least one of --machine-type or --disk-size-gb is required"))
				return
			}
			w := ctx.ProgressWriter()
			results := []cli.OpResultRow{}
			for _, idname := range idNames {
				id := ctx.PickResourceID(idname)
				req.InstanceID = sdk.String(id)
				if machineType != "" {
					req.MachineType = sdk.String(machineType)
				}
				if diskSpace != 0 {
					req.DiskSpace = sdk.Int(diskSpace)
				}
				_, err := client.UpgradeUPgSQLInstance(req)
				if err != nil {
					ctx.HandleError(err)
					continue
				}
				if async {
					fmt.Fprintf(w, "pgsql[%s] is upgrading\n", idname)
				} else {
					text := fmt.Sprintf("pgsql[%s] is upgrading", idname)
					ctx.PollerTo(w, describePgsqlByID(ctx)).Spoll(id, text, []string{PGSQL_RUNNING, PGSQL_INIT_FAILED, PGSQL_START_FAILED})
				}
				results = append(results, cli.OpResultRow{ResourceID: id, Action: "upgrade", Status: "Upgrading"})
			}
			ctx.EmitResult(results...)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&idNames, "instance-id", nil, "Required. Resource ID of UPgSQL instances to upgrade")
	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	flags.StringVar(&machineType, "machine-type", "", "Optional. New machine type ID. See 'ucloud pgsql db list-machine-type'")
	flags.IntVar(&diskSpace, "disk-size-gb", 0, "Optional. New disk size (GiB)")
	flags.BoolVarP(&async, "async", "a", false, "Optional. Do not wait for the long-running operation to finish")

	cmd.MarkFlagRequired("instance-id")

	command.SetCompletion(cmd, "instance-id", func() []string {
		return getUPgSQLIDList(ctx, req.GetProjectId(), req.GetRegion(), req.GetZone())
	})
	command.SetCompletion(cmd, "machine-type", func() []string {
		return listMachineTypeIDNames(ctx, req.GetProjectId(), req.GetRegion(), req.GetZone())
	})

	return cmd
}
