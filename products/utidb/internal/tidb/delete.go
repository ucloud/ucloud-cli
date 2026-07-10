package tidb

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/tidb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newDelete ucloud utidb delete
func newDelete(ctx *cli.Context) *cobra.Command {
	var id string
	var deleteBackup bool
	var yes bool

	client := cli.NewServiceClient(ctx, tidb.NewClient)
	req := client.NewDeleteTiDBClusterServiceRequest()

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a UTiDB instance",
		Long:  "Delete a UTiDB instance",
		Run: func(c *cobra.Command, args []string) {
			ok, err := ctx.Confirm(yes, fmt.Sprintf("Are you sure to delete UTiDB instance %s?", id))
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if !ok {
				return
			}

			pickedID := ctx.PickResourceID(id)
			req.Id = sdk.String(pickedID)
			req.DeleteBackup = sdk.Bool(deleteBackup)

			_, err = client.DeleteTiDBClusterService(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}

			w := ctx.ProgressWriter()
			text := fmt.Sprintf("utidb[%s] is deleting", pickedID)
			ctx.PollerTo(w, describeByID(ctx, req.GetRegion(), req.GetZone(), req.GetProjectId())).Spoll(pickedID, text, []string{stateDeleted, stateDeleteFail})
			ctx.EmitResult(cli.OpResultRow{ResourceID: pickedID, Action: "delete", Status: "Deleted"})
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringVar(&id, "utidb-id", "", "Required. Resource ID of the UTiDB instance to delete")
	flags.BoolVar(&deleteBackup, "delete-backup", false, "Optional. Also delete backup data")
	flags.BoolVarP(&yes, "yes", "y", false, "Optional. Do not prompt for confirmation")

	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)

	cmd.MarkFlagRequired("utidb-id")
	command.SetCompletion(cmd, "utidb-id", func() []string {
		return listResourceIDs(ctx, nil, *req.Region, *req.Zone, *req.ProjectId)
	})

	return cmd
}
