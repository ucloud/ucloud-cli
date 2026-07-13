package ukafka

import (
	"fmt"

	"github.com/spf13/cobra"

	ukafkasdk "github.com/ucloud/ucloud-sdk-go/services/ukafka"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newResizeDisk ucloud ukafka resize-disk
func newResizeDisk(ctx *cli.Context) *cobra.Command {
	var async *bool
	var instanceID *string
	var diskSize *int

	client := cli.NewServiceClient(ctx, ukafkasdk.NewClient)
	req := client.NewResizeUKafkaDiskRequest()

	cmd := &cobra.Command{
		Use:   "resize-disk",
		Short: "Resize UKafka instance disk",
		Long:  "Resize UKafka instance disk size",
		Run: func(cmd *cobra.Command, args []string) {
			w := ctx.ProgressWriter()
			req.InstanceId = sdk.String(*instanceID)
			req.DiskSize = sdk.Int(*diskSize)

			_, err := client.ResizeUKafkaDisk(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}

			text := fmt.Sprintf("ukafka[%s] is resizing disk to %dGB", *instanceID, *diskSize)
			if *async {
				fmt.Fprintln(w, text)
			} else {
				ctx.PollerTo(w, describeUKafkaInstanceByID(ctx)).Spoll(*instanceID, text, []string{STATE_RUNNING, STATE_ABNORMAL})
			}
			ctx.EmitResult(cli.OpResultRow{ResourceID: *instanceID, Action: "resize-disk", Status: "Running"})
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	instanceID = flags.String("instance-id", "", "Required. Instance ID")
	diskSize = flags.Int("disk-size-gb", 0, "Required. Target disk size in GB")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	req.Zone = flags.String("zone", ctx.DefaultZone(), "Optional. Assign availability zone")
	async = flags.Bool("async", false, "Optional. Do not wait for operation to finish")

	cmd.MarkFlagRequired("instance-id")
	cmd.MarkFlagRequired("disk-size-gb")

	return cmd
}
