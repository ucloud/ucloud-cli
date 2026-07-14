package pgsql

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/upgsql"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newDelete ucloud pgsql db delete
func newDelete(ctx *cli.Context) *cobra.Command {
	var idNames []string
	var yes bool
	client := newUPgSQLClient(ctx)
	req := client.NewDeleteUPgSQLInstanceRequest()
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete UPgSQL instances by instance-id",
		Long:  "Delete UPgSQL instances by instance-id",
		Run: func(c *cobra.Command, args []string) {
			ok, err := ctx.Confirm(yes, "Are you sure you want to delete the pgsql instance(s)?")
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
				any, err := describePgsqlByID(ctx)(id, nil)
				if err != nil {
					ctx.HandleError(err)
					continue
				}
				req.InstanceID = &id
				ins, ok := any.(*upgsql.UDBInstance)
				if ok && ins.State == PGSQL_RUNNING {
					stopReq := client.NewStopUPgSQLInstanceRequest()
					stopReq.ProjectId = req.ProjectId
					stopReq.Region = req.Region
					stopReq.Zone = req.Zone
					stopReq.InstanceID = req.InstanceID
					stopPgsqlIns(ctx, stopReq, false, w)
				}
				_, err = client.DeleteUPgSQLInstance(req)
				if err != nil {
					ctx.HandleError(err)
					continue
				}
				fmt.Fprintf(w, "pgsql[%s] deleted\n", idname)
				results = append(results, cli.OpResultRow{ResourceID: id, Action: "delete", Status: "Deleted"})
			}
			ctx.EmitResult(results...)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&idNames, "instance-id", nil, "Required. Resource ID of UPgSQL instances to delete")
	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	flags.BoolVarP(&yes, "yes", "y", false, "Optional. Do not prompt for confirmation.")

	cmd.MarkFlagRequired("instance-id")
	command.SetCompletion(cmd, "instance-id", func() []string {
		return getUPgSQLIDList(ctx, req.GetProjectId(), req.GetRegion(), req.GetZone())
	})
	return cmd
}
