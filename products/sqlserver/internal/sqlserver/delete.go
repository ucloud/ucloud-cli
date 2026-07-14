package sqlserver

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/udb"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newDelete returns the "delete" command for SQL Server instances.
func newDelete(ctx *cli.Context) *cobra.Command {
	var idNames []string
	var yes bool
	client := cli.NewServiceClient(ctx, udb.NewClient)
	req := client.NewDeleteUDBInstanceRequest()
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete SQL Server instances by udb-id",
		Long:  "Delete SQL Server instances by udb-id",
		Run: func(c *cobra.Command, args []string) {
			ok, err := ctx.Confirm(yes, "Are you sure you want to delete the udb(s)?")
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if !ok {
				return
			}
			w := ctx.ProgressWriter()
			results := []cli.OpResultRow{}
			for _, idname := range idNames {
				id := ctx.PickResourceID(idname)
				any, err := describeUdbByID(ctx)(id, nil)
				if err != nil {
					ctx.HandleError(err)
					continue
				}
				req.DBId = &id
				ins, ok := any.(*udb.UDBInstanceSet)
				if ok && ins.State == UDB_RUNNING {
					stopReq := client.NewStopUDBInstanceRequest()
					stopReq.ProjectId = req.ProjectId
					stopReq.Region = req.Region
					stopReq.Zone = req.Zone
					stopReq.DBId = req.DBId
					stopUdbIns(ctx, stopReq, false, w)
				}
				_, err = client.DeleteUDBInstance(req)
				if err != nil {
					ctx.HandleError(err)
					continue
				}
				fmt.Fprintf(w, "udb[%s] deleted\n", idname)
				results = append(results, cli.OpResultRow{ResourceID: id, Action: "delete", Status: "Deleted"})
			}
			ctx.EmitResult(results...)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&idNames, "udb-id", nil, "Required. Resource ID of UDB instances to delete")
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)
	flags.BoolVarP(&yes, "yes", "y", false, "Optional. Do not prompt for confirmation.")

	cmd.MarkFlagRequired("udb-id")
	command.SetCompletion(cmd, "udb-id", func() []string {
		return getUDBIDList(ctx, nil, "", *req.ProjectId, *req.Region, *req.Zone)
	})
	return cmd
}
