package css

import (
	"fmt"

	"github.com/spf13/cobra"

	uessdk "github.com/ucloud/ucloud-sdk-go/services/ues"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newRestart ucloud css restart
func newRestart(ctx *cli.Context) *cobra.Command {
	var async *bool
	var yes *bool
	var instanceID *string
	client := cli.NewServiceClient(ctx, uessdk.NewClient)
	req := client.NewRestartUESInstanceRequest()
	cmd := &cobra.Command{
		Use:   "restart",
		Short: "Restart UES instance",
		Long:  "Restart UES instance",
		Run: func(cmd *cobra.Command, args []string) {
			id := ctx.PickResourceID(*instanceID)
			ok, err := ctx.Confirm(*yes, fmt.Sprintf("Are you sure to restart UES instance[%s]?", id))
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if !ok {
				return
			}
			req.InstanceId = &id
			w := ctx.ProgressWriter()
			_, err = client.RestartUESInstance(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			text := fmt.Sprintf("ues[%s] is restarting", id)
			if *async {
				fmt.Fprintln(w, text)
			} else {
				ctx.PollerTo(w, describeUESInstanceByID(ctx)).Spoll(id, text, []string{STATE_RUNNING, STATE_ABNORMAL})
			}
			ctx.EmitResult(cli.OpResultRow{ResourceID: id, Action: "restart", Status: "Restarting"})
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	instanceID = flags.String("instance-id", "", "Required. Instance ID to restart")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	req.Zone = flags.String("zone", ctx.DefaultZone(), "Optional. Assign availability zone")
	yes = flags.BoolP("yes", "y", false, "Optional. Skip confirmation prompt")
	async = flags.Bool("async", false, "Optional. Do not wait for restart to finish")

	command.SetCompletion(cmd, "instance-id", func() []string {
		return getInstanceList(ctx, []string{STATE_RUNNING, STATE_ABNORMAL}, *req.ProjectId, *req.Region, "")
	})

	cmd.MarkFlagRequired("instance-id")

	return cmd
}
