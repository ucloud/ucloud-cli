package ukafka

import (
	"fmt"

	"github.com/spf13/cobra"

	ukafkasdk "github.com/ucloud/ucloud-sdk-go/services/ukafka"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newCreate ucloud ukafka create
func newCreate(ctx *cli.Context) *cobra.Command {
	var async *bool
	client := cli.NewServiceClient(ctx, ukafkasdk.NewClient)
	req := client.NewCreateUKafkaInstanceRequest()
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create UKafka instance",
		Long:  "Create UKafka instance",
		Run: func(cmd *cobra.Command, args []string) {
			w := ctx.ProgressWriter()
			resp, err := client.CreateUKafkaInstance(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			text := fmt.Sprintf("ukafka[%s] is creating", resp.InstanceId)
			if *async {
				fmt.Fprintln(w, text)
			} else {
				ctx.PollerTo(w, describeUKafkaInstanceByID(ctx)).Spoll(resp.InstanceId, text, []string{StateRunning, StateAbnormal})
			}
			ctx.EmitResult(cli.OpResultRow{ResourceID: resp.InstanceId, Action: "create", Status: "Creating"})
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.InstanceName = flags.String("name", "", "Required. Instance name")
	req.FrameworkVersion = flags.String("kafka-version", "", "Required. Kafka version, e.g. 2.12-2.4.1")
	req.NodeType = flags.String("node-type", "", "Required. Node type")
	req.DiskSize = flags.Int("disk-size-gb", 0, "Required. Disk size in GB")
	req.NodeCount = flags.Int("node-count", 3, "Optional. Node count, default 3")
	req.LogRetentionHours = flags.String("log-retention-hours", "72", "Optional. Log retention hours (1-240), default 72")
	req.VPCId = flags.String("vpc-id", "", "Optional. VPC ID")
	req.SubnetId = flags.String("subnet-id", "", "Optional. Subnet ID")
	req.BusinessId = flags.String("business-id", "", "Optional. Business group ID")
	req.Quantity = flags.String("quantity", "1", "Optional. Instance quantity, default 1")
	req.DiskControllerType = flags.String("disk-controller-type", "NONE", "Optional. Disk controller type: NONE or CLEAN")
	req.DiskThreshold = flags.String("disk-threshold", "90", "Optional. Disk cleanup threshold (70-90), default 90")
	req.IsSecurityEnabled = flags.String("enable-security", "false", "Optional. Enable security group: true or false")
	async = flags.Bool("async", false, "Optional. Do not wait for creation to finish")

	// Bind common params with Tab completion
	// Note: UKafka SDK uses *string for ChargeType/Quantity (not *int),
	// so we cannot use ctx.BindCommonParams which assumes standard types.
	// Instead we bind region/zone/project-id individually with completion.
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)

	command.SetFlagValues(cmd, "charge-type", "Month", "Year", "Dynamic")

	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("kafka-version")
	cmd.MarkFlagRequired("node-type")
	cmd.MarkFlagRequired("disk-size-gb")

	return cmd
}
