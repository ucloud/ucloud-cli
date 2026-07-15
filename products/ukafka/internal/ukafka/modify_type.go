package ukafka

import (
	"fmt"

	"github.com/spf13/cobra"

	ukafkasdk "github.com/ucloud/ucloud-sdk-go/services/ukafka"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newModifyType ucloud ukafka modify-type
func newModifyType(ctx *cli.Context) *cobra.Command {
	var async *bool
	var instanceID *string
	var nodeType *string

	client := cli.NewServiceClient(ctx, ukafkasdk.NewClient)
	req := client.NewModifyUKafkaInstanceTypeRequest()

	cmd := &cobra.Command{
		Use:   "modify-type",
		Short: "Modify UKafka instance type (CPU and memory)",
		Long:  "Modify UKafka instance type, only upgrade CPU and memory",
		Run: func(cmd *cobra.Command, args []string) {
			w := ctx.ProgressWriter()
			req.InstanceId = sdk.String(*instanceID)
			req.NodeType = sdk.String(*nodeType)

			_, err := client.ModifyUKafkaInstanceType(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}

			text := fmt.Sprintf("ukafka[%s] is modifying type to %s", *instanceID, *nodeType)
			if *async {
				fmt.Fprintln(w, text)
			} else {
				ctx.PollerTo(w, describeUKafkaInstanceByID(ctx)).Spoll(*instanceID, text, []string{StateRunning, StateAbnormal})
			}
			ctx.EmitResult(cli.OpResultRow{ResourceID: *instanceID, Action: "modify-type", Status: "Running"})
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	instanceID = flags.String("ukafka-id", "", "Required. Instance ID")
	nodeType = flags.String("node-type", "", "Required. Target node type")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	req.Zone = flags.String("zone", ctx.DefaultZone(), "Optional. Assign availability zone")
	async = flags.Bool("async", false, "Optional. Do not wait for operation to finish")

	cmd.MarkFlagRequired("ukafka-id")
	cmd.MarkFlagRequired("node-type")

	return cmd
}
