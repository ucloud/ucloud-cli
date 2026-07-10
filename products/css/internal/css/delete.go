package css

import (
	"fmt"

	"github.com/spf13/cobra"

	uessdk "github.com/ucloud/ucloud-sdk-go/services/ues"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newDelete ucloud css delete
func newDelete(ctx *cli.Context) *cobra.Command {
	var yes *bool
	var instanceIDs *[]string
	client := cli.NewServiceClient(ctx, uessdk.NewClient)
	req := client.NewDeleteUESInstanceRequest()
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete UES instances",
		Long:  "Delete UES instances",
		Run: func(cmd *cobra.Command, args []string) {
			ok, err := ctx.Confirm(*yes, "Are you sure to delete UES instance(s)?")
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if !ok {
				return
			}
			w := ctx.ProgressWriter()
			results := []cli.OpResultRow{}
			for _, idName := range *instanceIDs {
				id := ctx.PickResourceID(idName)
				req.InstanceId = &id
				_, err := client.DeleteUESInstance(req)
				if err != nil {
					ctx.HandleError(err)
					continue
				}
				fmt.Fprintf(w, "ues[%s] deleted\n", id)
				results = append(results, cli.OpResultRow{ResourceID: id, Action: "delete", Status: "Deleted"})
			}
			ctx.EmitResult(results...)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	instanceIDs = flags.StringSlice("instance-id", nil, "Required. Instance ID(s) to delete")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	req.Zone = flags.String("zone", ctx.DefaultZone(), "Optional. Assign availability zone")
	yes = flags.BoolP("yes", "y", false, "Optional. Skip confirmation prompt")

	command.SetCompletion(cmd, "instance-id", func() []string {
		return getInstanceList(ctx, []string{STATE_RUNNING, STATE_STOPPED, STATE_ABNORMAL}, *req.ProjectId, *req.Region, "")
	})

	cmd.MarkFlagRequired("instance-id")

	return cmd
}
