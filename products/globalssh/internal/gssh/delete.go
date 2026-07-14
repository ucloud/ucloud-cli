package gssh

import (
	"fmt"

	"github.com/spf13/cobra"

	pathxsdk "github.com/ucloud/ucloud-sdk-go/services/pathx"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newDelete ucloud gssh delete
func newDelete(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, pathxsdk.NewClient)
	req := client.NewDeleteGlobalSSHInstanceRequest()
	gsshIds := []string{}
	cmd := &cobra.Command{
		Use:     "delete",
		Short:   "Delete GlobalSSH instance",
		Long:    "Delete GlobalSSH instance",
		Example: "ucloud gssh delete --gssh-id uga-xx1  --id uga-xx2",
		Run: func(cmd *cobra.Command, args []string) {
			req.ProjectId = sdk.String(ctx.PickResourceID(*req.ProjectId))
			results := []cli.OpResultRow{}
			for _, idname := range gsshIds {
				id := ctx.PickResourceID(idname)
				req.InstanceId = sdk.String(id)
				_, err := client.DeleteGlobalSSHInstance(req)
				if err != nil {
					ctx.HandleError(err)
					continue
				}
				fmt.Fprintf(ctx.ProgressWriter(), "gssh[%s] deleted\n", idname)
				results = append(results, cli.OpResultRow{ResourceID: id, Action: "delete", Status: "Deleted"})
			}
			ctx.EmitResult(results...)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	flags.StringSliceVar(&gsshIds, "gssh-id", make([]string, 0), "Required. ID of the GlobalSSH instances you want to delete. Multiple values specified by multiple commas")
	ctx.BindProjectID(cmd, req)
	cmd.MarkFlagRequired("gssh-id")
	ctx.SetCompletion(cmd, "gssh-id", func() []string {
		return getAllGsshIDNames(ctx, *req.ProjectId)
	})
	return cmd
}
