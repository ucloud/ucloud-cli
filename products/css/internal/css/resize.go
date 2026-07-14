package css

import (
	"fmt"

	"github.com/spf13/cobra"

	uessdk "github.com/ucloud/ucloud-sdk-go/services/ues"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newResize ucloud css resize
func newResize(ctx *cli.Context) *cobra.Command {
	var async *bool
	var instanceID *string
	client := cli.NewServiceClient(ctx, uessdk.NewClient)
	req := client.NewResizeUESInstanceRequest()
	cmd := &cobra.Command{
		Use:   "resize",
		Short: "Resize UES instance node configuration",
		Long:  "Resize UES instance node configuration. Set node-conf to change spec, or node-disk-size-gb to change disk (leave the other at zero/empty).",
		Run: func(cmd *cobra.Command, args []string) {
			id := ctx.PickResourceID(*instanceID)
			req.InstanceId = &id
			w := ctx.ProgressWriter()
			_, err := client.ResizeUESInstance(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			text := fmt.Sprintf("ues[%s] is resizing", id)
			if *async {
				fmt.Fprintln(w, text)
			} else {
				ctx.PollerTo(w, describeUESInstanceByID(ctx)).Spoll(id, text, []string{STATE_RUNNING, STATE_ABNORMAL})
			}
			ctx.EmitResult(cli.OpResultRow{ResourceID: id, Action: "resize", Status: "Resizing"})
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	instanceID = flags.String("instance-id", "", "Required. Instance ID to resize")
	req.NodeRole = flags.String("node-role", "", "Required. Node role ('compute', 'master', 'coordinating', 'kibana', 'dashboard')")
	req.NodeConf = flags.String("node-conf", "", "Optional. Target node configuration. When empty, resize by node-disk-size-gb")
	req.NodeDiskSize = flags.Int("node-disk-size-gb", 0, "Optional. Target node disk size in GB. When 0, resize by node-conf")
	req.ForceResizing = flags.Bool("force", false, "Optional. Force resize without cluster health check, default false")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	req.Zone = flags.String("zone", ctx.DefaultZone(), "Optional. Assign availability zone")
	async = flags.Bool("async", false, "Optional. Do not wait for resize to finish")

	command.SetFlagValues(cmd, "node-role", "compute", "master", "coordinating", "kibana", "dashboard")
	command.SetCompletion(cmd, "instance-id", func() []string {
		return getInstanceList(ctx, []string{STATE_RUNNING}, *req.ProjectId, *req.Region, "")
	})

	cmd.MarkFlagRequired("instance-id")
	cmd.MarkFlagRequired("node-role")

	return cmd
}
