package ukafka

import (
	"fmt"

	"github.com/spf13/cobra"

	ukafkasdk "github.com/ucloud/ucloud-sdk-go/services/ukafka"

	"github.com/ucloud/ucloud-cli/internal/common"
	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// ListUKafkaInstanceResponse 自定义响应结构，修复 SDK 类型问题
type ListUKafkaInstanceResponse struct {
	RetCode    int             `json:"RetCode"`
	Action     string          `json:"Action"`
	TotalCount int             `json:"TotalCount"`
	ClusterSet []ClusterSetRaw `json:"ClusterSet"`
	Message    string          `json:"Message"`
}

// ClusterSetRaw 实例信息
type ClusterSetRaw struct {
	Zone                string `json:"Zone"`
	ClusterInstanceId   string `json:"ClusterInstanceId"`
	ClusterInstanceName string `json:"ClusterInstanceName"`
	Framework           string `json:"Framework"`
	FrameworkVersion    string `json:"FrameworkVersion"`
	Remark              string `json:"Remark"`
	CreateTime          int    `json:"CreateTime"`
	RunningTime         int    `json:"RunningTime"`
	ExpireTime          int    `json:"ExpireTime"`
	AutoRenew           string `json:"AutoRenew"`
	ChargeType          string `json:"ChargeType"`
	UHostCount          int    `json:"UHostCount"`
	State               string `json:"State"`
	Tag                 string `json:"Tag"`
	InstanceGroupType   string `json:"InstanceGroupType"`
	VPCId               string `json:"VPCId"`
	SubnetId            string `json:"SubnetId"`
	BusinessId          string `json:"BusinessId"`
}

// newList ucloud ukafka list
func newList(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, ukafkasdk.NewClient)
	req := client.NewListUKafkaInstanceRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List UKafka instances",
		Long:  "List UKafka instances",
		Run: func(cmd *cobra.Command, args []string) {
			// 使用 GenericInvoke 绕过 SDK 类型问题
			genReq := client.Client.NewGenericRequest()
			genReq.SetAction("ListUKafkaInstance")
			genReq.SetRegion(*req.Region)
			genReq.SetZone(*req.Zone)
			if req.ProjectId != nil && *req.ProjectId != "" {
				genReq.SetProjectId(*req.ProjectId)
			}

			// 构建额外参数
			payload := map[string]interface{}{}
			if req.Offset != nil && *req.Offset != "0" {
				payload["Offset"] = *req.Offset
			}
			if req.Limit != nil && *req.Limit != "60" {
				payload["Limit"] = *req.Limit
			}
			if req.VPCId != nil && *req.VPCId != "" {
				payload["VPCId"] = *req.VPCId
			}
			if req.SubnetId != nil && *req.SubnetId != "" {
				payload["SubnetId"] = *req.SubnetId
			}
			if req.BusinessId != nil && *req.BusinessId != "" {
				payload["BusinessId"] = *req.BusinessId
			}
			if len(payload) > 0 {
				genReq.SetPayload(payload)
			}

			genResp, err := client.Client.GenericInvoke(genReq)
			if err != nil {
				ctx.HandleError(err)
				return
			}

			var resp ListUKafkaInstanceResponse
			if err := genResp.Unmarshal(&resp); err != nil {
				ctx.HandleError(fmt.Errorf("parse response: %w", err))
				return
			}

			if resp.RetCode != 0 {
				ctx.HandleError(fmt.Errorf("API error: RetCode=%d, Message=%s", resp.RetCode, resp.Message))
				return
			}

			list := []InstanceRow{}
			for _, ins := range resp.ClusterSet {
				row := InstanceRow{
					InstanceID:   ins.ClusterInstanceId,
					InstanceName: ins.ClusterInstanceName,
					Framework:    ins.Framework,
					Version:      ins.FrameworkVersion,
					Zone:         ins.Zone,
					State:        ins.State,
					NodeCount:    fmt.Sprintf("%d", ins.UHostCount),
					VPCId:        ins.VPCId,
					SubnetId:     ins.SubnetId,
					ChargeType:   ins.ChargeType,
					CreateTime:   common.FormatDate(ins.CreateTime),
					ExpireTime:   common.FormatDate(ins.ExpireTime),
				}
				list = append(list, row)
			}
			ctx.PrintList(list)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	req.Zone = flags.String("zone", "", "Optional. Assign availability zone")
	req.Offset = flags.String("offset", "0", "Optional. Offset")
	req.Limit = flags.String("limit", "60", "Optional. Limit, default 60")
	req.VPCId = flags.String("vpc-id", "", "Optional. VPC ID")
	req.SubnetId = flags.String("subnet-id", "", "Optional. Subnet ID")
	req.BusinessId = flags.String("business-id", "", "Optional. Business group ID")
	return cmd
}
