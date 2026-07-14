package clickhouse

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	uclickhousesdk "github.com/ucloud/ucloud-sdk-go/services/uclickhouse"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"
	"github.com/ucloud/ucloud-sdk-go/ucloud/response"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

type describeUClickhouseClusterResponse struct {
	response.CommonBase
	Data    describeUClickhouseClusterResponseData
	Message string
}

type describeUClickhouseClusterResponseData struct {
	ClickhouseNodes []uclickhousesdk.ClickhouseNode
	Cluster         uclickhousesdk.ClickhouseCluster
	Payment         uclickhousePayment
	ZookeeperNodes  []uclickhousesdk.ZookeeperNode
}

type uclickhousePayment struct {
	ChargeType      string
	CreateTimestamp int
	ExpireTimestamp int
	OriginalPrice   flexibleString
	Price           flexibleString
	ResourceId      string
}

type flexibleString string

func (v *flexibleString) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*v = ""
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*v = flexibleString(s)
		return nil
	}
	var number json.Number
	if err := json.Unmarshal(data, &number); err != nil {
		return err
	}
	*v = flexibleString(number.String())
	return nil
}

func (v flexibleString) String() string {
	return string(v)
}

// newDescribe ucloud clickhouse describe
func newDescribe(ctx *cli.Context) *cobra.Command {
	var clusterID *string
	client := cli.NewServiceClient(ctx, uclickhousesdk.NewClient)
	req := client.NewDescribeUClickhouseClusterRequest()
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Describe UClickhouse cluster details",
		Long:  "Describe UClickhouse cluster details",
		Args:  noArgs,
		Run: func(cmd *cobra.Command, args []string) {
			id := ctx.PickResourceID(*clusterID)
			req.ClusterId = sdk.String(id)
			resp, err := describeUClickhouseCluster(client, req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			ctx.PrintList(describeRows(resp.Data))
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	clusterID = flags.String("clickhouse-id", "", "Required. UClickhouse cluster ID to describe")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")

	command.SetCompletion(cmd, "clickhouse-id", func() []string {
		return getClusterList(ctx, nil, *req.ProjectId, *req.Region)
	})

	cmd.MarkFlagRequired("clickhouse-id")
	return cmd
}

func describeUClickhouseCluster(client *uclickhousesdk.UClickhouseClient, req *uclickhousesdk.DescribeUClickhouseClusterRequest) (*describeUClickhouseClusterResponse, error) {
	var resp describeUClickhouseClusterResponse
	reqCopier := *req
	err := invokeUClickhouseAction(client, "DescribeUClickhouseCluster", &reqCopier, &resp)
	return &resp, err
}

func describeRows(data describeUClickhouseClusterResponseData) []cli.DescribeRow {
	cluster := data.Cluster
	rows := []cli.DescribeRow{
		{Attribute: "ClusterID", Content: cluster.ClusterId},
		{Attribute: "ClusterName", Content: cluster.ClusterName},
		{Attribute: "Status", Content: cluster.Status},
		{Attribute: "VPCId", Content: cluster.VPCId},
		{Attribute: "SubnetId", Content: cluster.SubnetId},
		{Attribute: "ClickhouseVersion", Content: cluster.ClickhouseVersion},
		{Attribute: "ZookeeperVersion", Content: cluster.ZookeeperVersion},
		{Attribute: "MachineType", Content: cluster.MachineType},
		{Attribute: "ShardCount", Content: fmt.Sprintf("%d", cluster.ShardCount)},
		{Attribute: "ReplicateCount", Content: fmt.Sprintf("%d", cluster.ReplicateCount)},
		{Attribute: "ClickhouseMachineTypeID", Content: cluster.ClickhouseMachineTypeId},
		{Attribute: "ClickhouseMachineTypeName", Content: cluster.ClickhouseMachineTypeName},
		{Attribute: "ClickhouseDataDiskType", Content: cluster.ClickhouseDataDiskType},
		{Attribute: "ClickhouseDataDiskSize", Content: fmt.Sprintf("%dGB", cluster.ClickhouseDataDiskSize)},
		{Attribute: "ClickhouseNodeCPU", Content: fmt.Sprintf("%d", cluster.ClickhouseNodeCPU)},
		{Attribute: "ClickhouseNodeMemory", Content: fmt.Sprintf("%dGB", cluster.ClickhouseNodeMemory)},
		{Attribute: "ZookeeperMachineTypeID", Content: cluster.ZookeeperMachineTypeId},
		{Attribute: "ZookeeperMachineTypeName", Content: cluster.ZookeeperMachineTypeName},
		{Attribute: "ZookeeperDataDiskType", Content: cluster.ZookeeperDataDiskType},
		{Attribute: "ZookeeperDataDiskSize", Content: fmt.Sprintf("%dGB", cluster.ZookeeperDataDiskSize)},
		{Attribute: "ZookeeperNodeCPU", Content: fmt.Sprintf("%d", cluster.ZookeeperNodeCPU)},
		{Attribute: "ZookeeperNodeMemory", Content: fmt.Sprintf("%dGB", cluster.ZookeeperNodeMemory)},
		{Attribute: "IsZookeeperHA", Content: cluster.IsZookeeperHA},
		{Attribute: "IsSecgroup", Content: cluster.IsSecgroup},
		{Attribute: "IsBackup", Content: cluster.IsBackup},
		{Attribute: "IsTieredStorage", Content: cluster.IsTieredStorage},
		{Attribute: "MultiZones", Content: strings.Join(cluster.MultiZones, ",")},
		{Attribute: "CreateTime", Content: formatUnixDate(cluster.CreateTimestamp)},
		{Attribute: "ExpireTime", Content: formatUnixDate(int(cluster.ExpireTimestamp))},
		{Attribute: "Payment.ChargeType", Content: data.Payment.ChargeType},
		{Attribute: "Payment.Price", Content: data.Payment.Price.String()},
		{Attribute: "Payment.OriginalPrice", Content: data.Payment.OriginalPrice.String()},
	}
	if len(data.ClickhouseNodes) > 0 {
		rows = append(rows, cli.DescribeRow{Attribute: "--- ClickhouseNodes ---", Content: fmt.Sprintf("%d nodes", len(data.ClickhouseNodes))})
		for i, node := range data.ClickhouseNodes {
			prefix := fmt.Sprintf("ClickhouseNode[%d]", i)
			rows = append(rows,
				cli.DescribeRow{Attribute: prefix + ".NodeID", Content: node.NodeId},
				cli.DescribeRow{Attribute: prefix + ".NodeName", Content: node.NodeName},
				cli.DescribeRow{Attribute: prefix + ".Zone", Content: node.Zone},
				cli.DescribeRow{Attribute: prefix + ".IPv4", Content: node.IPv4},
				cli.DescribeRow{Attribute: prefix + ".ServiceStatus", Content: node.ServiceStatus},
				cli.DescribeRow{Attribute: prefix + ".ShardGroup", Content: node.ShardGroup},
				cli.DescribeRow{Attribute: prefix + ".MachineType", Content: node.MachineType},
				cli.DescribeRow{Attribute: prefix + ".CPU", Content: fmt.Sprintf("%d", node.CPU)},
				cli.DescribeRow{Attribute: prefix + ".Memory", Content: fmt.Sprintf("%dGB", node.Memory)},
				cli.DescribeRow{Attribute: prefix + ".DataDiskSize", Content: fmt.Sprintf("%dGB", node.DataDiskSize)},
				cli.DescribeRow{Attribute: prefix + ".DataDiskType", Content: node.DataDiskType},
			)
		}
	}
	if len(data.ZookeeperNodes) > 0 {
		rows = append(rows, cli.DescribeRow{Attribute: "--- ZookeeperNodes ---", Content: fmt.Sprintf("%d nodes", len(data.ZookeeperNodes))})
		for i, node := range data.ZookeeperNodes {
			prefix := fmt.Sprintf("ZookeeperNode[%d]", i)
			rows = append(rows,
				cli.DescribeRow{Attribute: prefix + ".NodeID", Content: node.NodeId},
				cli.DescribeRow{Attribute: prefix + ".NodeName", Content: node.NodeName},
				cli.DescribeRow{Attribute: prefix + ".Zone", Content: node.Zone},
				cli.DescribeRow{Attribute: prefix + ".ServiceStatus", Content: node.ServiceStatus},
				cli.DescribeRow{Attribute: prefix + ".MachineType", Content: node.MachineType},
				cli.DescribeRow{Attribute: prefix + ".CPU", Content: fmt.Sprintf("%d", node.CPU)},
				cli.DescribeRow{Attribute: prefix + ".Memory", Content: fmt.Sprintf("%dGB", node.Memory)},
				cli.DescribeRow{Attribute: prefix + ".DataDiskSize", Content: fmt.Sprintf("%dGB", node.DataDiskSize)},
				cli.DescribeRow{Attribute: prefix + ".DataDiskType", Content: node.DataDiskType},
			)
		}
	}
	return rows
}

// describeClusterByID returns the poller's describe func.
func describeClusterByID(ctx *cli.Context) func(clusterID string, commonBase *request.CommonBase) (interface{}, error) {
	return func(clusterID string, commonBase *request.CommonBase) (interface{}, error) {
		client := cli.NewServiceClient(ctx, uclickhousesdk.NewClient)
		req := client.NewDescribeUClickhouseClusterRequest()
		if commonBase != nil {
			req.CommonBase = *commonBase
		}
		req.ClusterId = sdk.String(clusterID)
		resp, err := describeUClickhouseCluster(client, req)
		if err != nil {
			return nil, err
		}
		return &resp.Data.Cluster, nil
	}
}
