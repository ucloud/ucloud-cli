package ukafka

import (
	"fmt"

	"github.com/spf13/cobra"

	ukafkasdk "github.com/ucloud/ucloud-sdk-go/services/ukafka"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// DescribeUKafkaInstanceResponse 自定义响应结构
type DescribeUKafkaInstanceResponse struct {
	RetCode    int           `json:"RetCode"`
	Action     string        `json:"Action"`
	ClusterSet []ClusterInfo `json:"ClusterSet"`
	Message    string        `json:"Message"`
}

// ClusterInfo 实例信息
type ClusterInfo struct {
	Zone               string   `json:"Zone"`
	ClusterInstanceId  string   `json:"ClusterInstanceId"`
	ClusterInstanceName string  `json:"ClusterInstanceName"`
	Remark             string   `json:"Remark"`
	Tag                string   `json:"Tag"`
	Framework          string   `json:"Framework"`
	FrameworkVersion   string   `json:"FrameworkVersion"`
	NetworkId          string   `json:"NetworkId"`
	VPCId              string   `json:"VPCId"`
	SubnetId           string   `json:"SubnetId"`
	BusinessId         string   `json:"BusinessId"`
	UHostSet           []Broker `json:"UHostSet"`
	IsOpenSecgroup     bool     `json:"IsOpenSecgroup"`
	ChargeType         string   `json:"ChargeType"`
	AutoRenew          string   `json:"AutoRenew"`
	ValidBrokerNum     int      `json:"ValidBrokerNum"`
	UHostCount         int      `json:"UHostCount"`
	ExpireTime         int      `json:"ExpireTime"`
	CreateTime         int      `json:"CreateTime"`
	RunningTime        int      `json:"RunningTime"`
	State              string   `json:"State"`
}

// Broker 节点信息
type Broker struct {
	BrokerId          string      `json:"BrokerId"`
	UHostId           string      `json:"UHostId"`
	ResourceId        string      `json:"ResourceId"`
	UHostRole         string      `json:"UHostRole"`
	UHostName         string      `json:"UHostName"`
	DomainName        string      `json:"DomainName"`
	Remark            string      `json:"Remark"`
	CreateTime        int         `json:"CreateTime"`
	ExpireTime        int         `json:"ExpireTime"`
	InstanceGroupType string      `json:"InstanceGroupType"`
	SecurityGroupId   string      `json:"SecurityGroupId"`
	State             string      `json:"State"`
	ZooKeeper         string      `json:"ZooKeeper"`
	UHostConfig       UHostConfig `json:"UHostConfig"`
	IPSet             []IPInfo    `json:"IPSet"`
	KafkaPort         int         `json:"KafkaPort"`
	ZooKeeperPort     int         `json:"ZooKeeperPort"`
}

// UHostConfig 节点配置
type UHostConfig struct {
	CPU          int    `json:"CPU"`
	Memory       int    `json:"Memory"`
	DataDiskSize int    `json:"DataDiskSize"`
	DiskType     string `json:"DiskType"`
}

// IPInfo IP信息
type IPInfo struct {
	Type string `json:"Type"`
	IP   string `json:"IP"`
}

// newDescribe ucloud ukafka describe
func newDescribe(ctx *cli.Context) *cobra.Command {
	var instanceID *string
	client := cli.NewServiceClient(ctx, ukafkasdk.NewClient)
	req := client.NewDescribeUKafkaInstanceRequest()
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Describe UKafka instance details",
		Long:  "Describe UKafka instance details",
		Run: func(cmd *cobra.Command, args []string) {
			id := ctx.PickResourceID(*instanceID)

			// 使用 GenericInvoke 绕过 SDK 类型问题
			genReq := client.Client.NewGenericRequest()
			genReq.SetAction("DescribeUKafkaInstance")
			genReq.SetRegion(*req.Region)
			genReq.SetZone(*req.Zone)
			if req.ProjectId != nil && *req.ProjectId != "" {
				genReq.SetProjectId(*req.ProjectId)
			}

			payload := map[string]interface{}{
				"ClusterInstanceId": id,
			}
			genReq.SetPayload(payload)

			genResp, err := client.Client.GenericInvoke(genReq)
			if err != nil {
				ctx.HandleError(err)
				return
			}

			var resp DescribeUKafkaInstanceResponse
			if err := genResp.Unmarshal(&resp); err != nil {
				ctx.HandleError(fmt.Errorf("parse response: %w", err))
				return
			}

			if resp.RetCode != 0 {
				ctx.HandleError(fmt.Errorf("API error: RetCode=%d, Message=%s", resp.RetCode, resp.Message))
				return
			}

			if len(resp.ClusterSet) == 0 {
				ctx.EmitResult(cli.OpResultRow{ResourceID: id, Action: "describe", Status: "NotFound"})
				return
			}

			cluster := resp.ClusterSet[0]
			rows := []cli.DescribeRow{
				{Attribute: "InstanceID", Content: cluster.ClusterInstanceId},
				{Attribute: "InstanceName", Content: cluster.ClusterInstanceName},
				{Attribute: "Region", Content: *req.Region},
				{Attribute: "Zone", Content: cluster.Zone},
				{Attribute: "State", Content: cluster.State},
				{Attribute: "Framework", Content: cluster.Framework},
				{Attribute: "FrameworkVersion", Content: cluster.FrameworkVersion},
				{Attribute: "VPCId", Content: cluster.VPCId},
				{Attribute: "SubnetId", Content: cluster.SubnetId},
				{Attribute: "BusinessId", Content: cluster.BusinessId},
				{Attribute: "ChargeType", Content: cluster.ChargeType},
				{Attribute: "AutoRenew", Content: cluster.AutoRenew},
				{Attribute: "Remark", Content: cluster.Remark},
			}

			// Add node information
			if len(cluster.UHostSet) > 0 {
				rows = append(rows, cli.DescribeRow{Attribute: "--- Nodes ---", Content: fmt.Sprintf("%d nodes", len(cluster.UHostSet))})
				for i, node := range cluster.UHostSet {
					prefix := fmt.Sprintf("Node[%d]", i)
					var ip string
					if len(node.IPSet) > 0 {
						ip = node.IPSet[0].IP
					}
					rows = append(rows,
						cli.DescribeRow{Attribute: prefix + ".NodeID", Content: node.UHostId},
						cli.DescribeRow{Attribute: prefix + ".NodeName", Content: node.UHostName},
						cli.DescribeRow{Attribute: prefix + ".NodeRole", Content: node.UHostRole},
						cli.DescribeRow{Attribute: prefix + ".State", Content: node.State},
						cli.DescribeRow{Attribute: prefix + ".IP", Content: ip},
						cli.DescribeRow{Attribute: prefix + ".CPU", Content: fmt.Sprintf("%d", node.UHostConfig.CPU)},
						cli.DescribeRow{Attribute: prefix + ".Memory", Content: fmt.Sprintf("%dMB", node.UHostConfig.Memory)},
						cli.DescribeRow{Attribute: prefix + ".DiskSize", Content: fmt.Sprintf("%dGB", node.UHostConfig.DataDiskSize)},
						cli.DescribeRow{Attribute: prefix + ".DiskType", Content: node.UHostConfig.DiskType},
						cli.DescribeRow{Attribute: prefix + ".KafkaPort", Content: fmt.Sprintf("%d", node.KafkaPort)},
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

	cmd.MarkFlagRequired("instance-id")

	return cmd
}

// describeUKafkaInstanceByID returns the poller's describe func
func describeUKafkaInstanceByID(ctx *cli.Context) func(instanceID string, commonBase *request.CommonBase) (interface{}, error) {
	return func(instanceID string, commonBase *request.CommonBase) (interface{}, error) {
		client := cli.NewServiceClient(ctx, ukafkasdk.NewClient)
		genReq := client.Client.NewGenericRequest()
		genReq.SetAction("DescribeUKafkaInstance")
		genReq.SetRegion(ctx.DefaultRegion())
		genReq.SetZone(ctx.DefaultZone())

		payload := map[string]interface{}{
			"ClusterInstanceId": instanceID,
		}
		genReq.SetPayload(payload)

		genResp, err := client.Client.GenericInvoke(genReq)
		if err != nil {
			return nil, err
		}

		var resp DescribeUKafkaInstanceResponse
		if err := genResp.Unmarshal(&resp); err != nil {
			return nil, err
		}
		if len(resp.ClusterSet) == 0 {
			return nil, fmt.Errorf("instance not found")
		}
		return &resp.ClusterSet[0], nil
	}
}
