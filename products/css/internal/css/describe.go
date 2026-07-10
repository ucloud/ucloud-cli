package css

import (
	"fmt"

	"github.com/spf13/cobra"

	uessdk "github.com/ucloud/ucloud-sdk-go/services/ues"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newDescribe ucloud css describe
func newDescribe(ctx *cli.Context) *cobra.Command {
	var instanceID *string
	client := cli.NewServiceClient(ctx, uessdk.NewClient)
	req := client.NewDescribeUESInstanceV2Request()
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Describe UES instance details",
		Long:  "Describe UES instance details",
		Run: func(cmd *cobra.Command, args []string) {
			id := ctx.PickResourceID(*instanceID)
			req.InstanceId = sdk.String(id)
			resp, err := client.DescribeUESInstanceV2(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			cluster := resp.Result.ClusterInfo
			rows := []cli.DescribeRow{
				{Attribute: "InstanceID", Content: cluster.UESInstanceId},
				{Attribute: "InstanceName", Content: cluster.UESInstanceName},
				{Attribute: "Region", Content: cluster.Region},
				{Attribute: "Zone", Content: cluster.Zone},
				{Attribute: "State", Content: cluster.State},
				{Attribute: "ServiceVersion", Content: cluster.ServiceVersion},
				{Attribute: "VPCId", Content: cluster.VPCId},
				{Attribute: "SubnetId", Content: cluster.SubnetId},
				{Attribute: "VIP", Content: cluster.Vip},
				{Attribute: "BusinessId", Content: cluster.BusinessId},
			}
			// Add node information
			if len(resp.Result.NodeInfoList) > 0 {
				rows = append(rows, cli.DescribeRow{Attribute: "--- Nodes ---", Content: fmt.Sprintf("%d nodes", len(resp.Result.NodeInfoList))})
				for i, node := range resp.Result.NodeInfoList {
					prefix := fmt.Sprintf("Node[%d]", i)
					rows = append(rows,
						cli.DescribeRow{Attribute: prefix + ".NodeID", Content: node.NodeId},
						cli.DescribeRow{Attribute: prefix + ".NodeName", Content: node.NodeName},
						cli.DescribeRow{Attribute: prefix + ".NodeRole", Content: node.NodeRole},
						cli.DescribeRow{Attribute: prefix + ".NodeState", Content: node.NodeState},
						cli.DescribeRow{Attribute: prefix + ".NodeIP", Content: node.NodeIP},
						cli.DescribeRow{Attribute: prefix + ".NodeConf", Content: node.NodeConf},
						cli.DescribeRow{Attribute: prefix + ".CPU", Content: fmt.Sprintf("%d", node.CPU)},
						cli.DescribeRow{Attribute: prefix + ".Memory", Content: fmt.Sprintf("%dGB", node.Memory)},
						cli.DescribeRow{Attribute: prefix + ".DiskSize", Content: fmt.Sprintf("%dGB", node.DiskSize)},
						cli.DescribeRow{Attribute: prefix + ".DiskType", Content: node.DiskType},
					)
				}
			}
			ctx.PrintList(rows)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	instanceID = flags.String("instance-id", "", "Required. Instance ID to describe")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	req.Zone = flags.String("zone", ctx.DefaultZone(), "Optional. Assign availability zone")

	command.SetCompletion(cmd, "instance-id", func() []string {
		return getInstanceList(ctx, nil, *req.ProjectId, *req.Region, "")
	})

	cmd.MarkFlagRequired("instance-id")

	return cmd
}

// describeUESInstanceByID returns the poller's describe func
func describeUESInstanceByID(ctx *cli.Context) func(instanceID string, commonBase *request.CommonBase) (interface{}, error) {
	return func(instanceID string, commonBase *request.CommonBase) (interface{}, error) {
		client := cli.NewServiceClient(ctx, uessdk.NewClient)
		req := client.NewDescribeUESInstanceV2Request()
		if commonBase != nil {
			req.CommonBase = *commonBase
		}
		req.InstanceId = sdk.String(instanceID)
		resp, err := client.DescribeUESInstanceV2(req)
		if err != nil {
			return nil, err
		}
		return &resp.Result.ClusterInfo, nil
	}
}
