package mysql

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/udb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newRestore ucloud udb restore
func newRestore(ctx *cli.Context) *cobra.Command {
	var datetime, diskType string
	var async bool
	client := cli.NewServiceClient(ctx, udb.NewClient)
	req := client.NewCreateUDBInstanceByRecoveryRequest()
	cmd := &cobra.Command{
		Use:   "restore",
		Short: "Create MySQL instance and restore the newly created db to the specified DB at a specified point in time",
		Long:  "Create MySQL instance and restore the newly created db to the specified DB at a specified point in time",
		Run: func(c *cobra.Command, args []string) {
			t, err := time.Parse(time.RFC3339, datetime)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			req.RecoveryTime = sdk.Int(int(t.Unix()))
			req.SrcDBId = sdk.String(ctx.PickResourceID(*req.SrcDBId))
			if diskType == "" {
				any, err := describeUdbByID(ctx)(*req.SrcDBId, nil)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				ins, ok := any.(*udb.UDBInstanceSet)
				if !ok {
					ctx.HandleError(fmt.Errorf("fetch udb[%s] instance", *req.SrcDBId))
					return
				}
				req.UseSSD = &ins.UseSSD
			} else if diskType == "normal" {
				req.UseSSD = sdk.Bool(false)
			} else if diskType == "ssd" {
				req.UseSSD = sdk.Bool(true)
			}
			resp, err := client.CreateUDBInstanceByRecovery(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			w := ctx.ProgressWriter()
			if async {
				fmt.Fprintf(w, "udb[%s] is restorting from udb[%s] at time point %s", resp.DBId, *req.SrcDBId, datetime)
			} else {
				text := fmt.Sprintf("udb[%s] is restorting from udb[%s] at time point %s", resp.DBId, *req.SrcDBId, datetime)
				ctx.PollerTo(w, describeUdbByID(ctx)).Spoll(resp.DBId, text, []string{UDB_RUNNING, UDB_RECOVER_FAIL, UDB_FAIL})
			}
			ctx.EmitResult(cli.OpResultRow{ResourceID: resp.DBId, Action: "restore", Status: "Restoring"})
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	req.Name = flags.String("name", "", "Required. Name of UDB instance to create")
	req.SrcDBId = flags.String("src-udb-id", "", "Required. Resource ID of source UDB")
	flags.StringVar(&datetime, "restore-to-time", "", "Required. The date and time to restore the DB to. Value must be a time in Universal Coordinated Time (UTC) format.Example: 2019-02-23T23:45:00Z")
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)
	flags.StringVar(&diskType, "disk-type", "", "Optional. Disk type. The default is to be consistent with the source database. Accept values: normal, ssd")
	ctx.BindChargeType(cmd, req)
	ctx.BindQuantity(cmd, req)
	flags.BoolVarP(&async, "async", "a", false, "Optional. Do not wait for the long-running operation to finish")

	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("src-udb-id")
	cmd.MarkFlagRequired("restore-to-time")

	command.SetFlagValues(cmd, "disk-type", "noraml", "ssd")
	command.SetCompletion(cmd, "src-udb-id", func() []string {
		return getUDBIDList(ctx, nil, "sql", *req.ProjectId, *req.Region, *req.Zone)
	})

	return cmd
}
