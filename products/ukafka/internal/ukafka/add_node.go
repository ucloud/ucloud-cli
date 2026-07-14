package ukafka

import (
	"fmt"

	"github.com/spf13/cobra"

	ukafkasdk "github.com/ucloud/ucloud-sdk-go/services/ukafka"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newAddNode ucloud ukafka add-node
func newAddNode(ctx *cli.Context) *cobra.Command {
	var async *bool
	var nodeCount *int
	var nodeType *string
	var instanceID *string

	client := cli.NewServiceClient(ctx, ukafkasdk.NewClient)
	req := client.NewAddUKafkaInstanceNodeRequest()

	cmd := &cobra.Command{
		Use:   "add-node",
		Short: "Add nodes to UKafka instance",
		Long:  "Add nodes to UKafka instance",
		Run: func(cmd *cobra.Command, args []string) {
			w := ctx.ProgressWriter()
			req.InstanceId = sdk.String(*instanceID)
			req.NodeCount = sdk.String(fmt.Sprintf("%d", *nodeCount))
			req.NodeType = sdk.String(*nodeType)

			_, err := client.AddUKafkaInstanceNode(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}

			text := fmt.Sprintf("ukafka[%s] adding %d node(s)", *instanceID, *nodeCount)
			if *async {
				fmt.Fprintln(w, text)
			} else {
				ctx.PollerTo(w, describeUKafkaInstanceByID(ctx)).Spoll(*instanceID, text, []string{StateRunning, StateAbnormal})
			}
			ctx.EmitResult(cli.OpResultRow{ResourceID: *instanceID, Action: "add-node", Status: "Running"})
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	instanceID = flags.String("ukafka-id", "", "Required. Instance ID")
	nodeCount = flags.Int("node-count", 1, "Required. Number of nodes to add")
	nodeType = flags.String("node-type", "", "Required. Node type")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	req.Zone = flags.String("zone", ctx.DefaultZone(), "Optional. Assign availability zone")
	async = flags.Bool("async", false, "Optional. Do not wait for operation to finish")

	cmd.MarkFlagRequired("ukafka-id")
	cmd.MarkFlagRequired("node-type")

	return cmd
}
