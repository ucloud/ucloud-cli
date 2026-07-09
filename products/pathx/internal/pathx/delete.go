package pathx

import (
	"fmt"

	"github.com/spf13/cobra"

	pathxsdk "github.com/ucloud/ucloud-sdk-go/services/pathx"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newDelete ucloud pathx delete
func newDelete(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, pathxsdk.NewClient)
	deleteUga3Req := client.NewDeleteUGA3InstanceRequest()
	deleteUga3PortReq := client.NewDeleteUGA3PortRequest()
	yes := false
	var instanceId string
	removeCmd := &cobra.Command{
		Use:     "delete",
		Short:   "Delete the pathx resource and port",
		Long:    "Delete the pathx resource and port",
		Example: "ucloud pathx delete --id uga3-xxx",
		Run: func(cmd *cobra.Command, args []string) {
			if !yes {
				ok, err := ctx.Confirm(false, "Are you sure you want to delete this resource ?")
				if err != nil {
					fmt.Fprintln(ctx.ProgressWriter(), err)
					return
				}
				if !ok {
					return
				}
			}
			w := ctx.ProgressWriter()
			fmt.Fprintf(w, "Starting delete the pathx[%s] resource port\n", instanceId)
			deleteUga3PortReq.InstanceId = &instanceId
			_, deletePortErr := client.DeleteUGA3Port(deleteUga3PortReq)
			if deletePortErr != nil {
				ctx.HandleError(deletePortErr)
				return
			}
			fmt.Fprintf(w, "Starting delete the pathx[%s] resource\n", instanceId)
			deleteUga3Req.InstanceId = &instanceId
			deleteUga3Req.SetProjectIdRef(deleteUga3PortReq.GetProjectIdRef())
			deleteUga3Req.SetRegionRef(deleteUga3PortReq.GetRegionRef())
			deleteUga3Req.SetZoneRef(deleteUga3PortReq.GetZoneRef())
			_, err := client.DeleteUGA3Instance(deleteUga3Req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			ctx.EmitResult(cli.OpResultRow{ResourceID: instanceId, Action: "delete", Status: "Deleted"})
		},
	}
	flags := removeCmd.Flags()
	flags.SortFlags = false
	flags.StringVar(&instanceId, "id", "", "Required. It is the resource ID of pathx, and the deletion will be performed according to this")
	ctx.BindProjectID(removeCmd, deleteUga3PortReq)
	ctx.BindRegion(removeCmd, deleteUga3PortReq)
	ctx.BindZone(removeCmd, deleteUga3PortReq)
	removeCmd.MarkFlagRequired("id")
	flags.BoolVarP(&yes, "yes", "y", false, "Optional. Do not prompt for confirmation.")
	ctx.SetCompletion(removeCmd, "id", func() []string {
		return getPathxList(ctx, *deleteUga3PortReq.ProjectId, *deleteUga3PortReq.Region, *deleteUga3PortReq.Zone)
	})
	return removeCmd
}
