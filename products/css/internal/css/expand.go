package css

import (
	"fmt"

	"github.com/spf13/cobra"

	uessdk "github.com/ucloud/ucloud-sdk-go/services/ues"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newExpand ucloud css expand
func newExpand(ctx *cli.Context) *cobra.Command {
	var async *bool
	var instanceID *string
	client := cli.NewServiceClient(ctx, uessdk.NewClient)
	req := client.NewExpandUESInstanceRequest()
	cmd := &cobra.Command{
		Use:   "expand",
		Short: "Expand UES instance node count",
		Long:  "Expand UES instance node count",
		Run: func(cmd *cobra.Command, args []string) {
			id := ctx.PickResourceID(*instanceID)
			req.InstanceId = &id
			w := ctx.ProgressWriter()
			_, err := client.ExpandUESInstance(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			text := fmt.Sprintf("ues[%s] is expanding", id)
			if *async {
				fmt.Fprintln(w, text)
			} else {
				ctx.PollerTo(w, describeUESInstanceByID(ctx)).Spoll(id, text, []string{STATE_RUNNING, STATE_ABNORMAL})
			}
			ctx.EmitResult(cli.OpResultRow{ResourceID: id, Action: "expand", Status: "Expanding"})
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	instanceID = flags.String("instance-id", "", "Required. Instance ID to expand")
	req.NodeRole = flags.String("node-role", "", "Required. Node role to expand ('compute', 'coordinating')")
	req.NodeCount = flags.Int("node-count", 0, "Required. Node count after expansion")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	req.Zone = flags.String("zone", ctx.DefaultZone(), "Optional. Assign availability zone")
	async = flags.Bool("async", false, "Optional. Do not wait for expansion to finish")

	command.SetFlagValues(cmd, "node-role", "compute", "coordinating")
	command.SetCompletion(cmd, "instance-id", func() []string {
		return getInstanceList(ctx, []string{STATE_RUNNING}, *req.ProjectId, *req.Region, "")
	})

	cmd.MarkFlagRequired("instance-id")
	cmd.MarkFlagRequired("node-role")
	cmd.MarkFlagRequired("node-count")

	return cmd
}
