package ukafka

import (
	"fmt"

	"github.com/spf13/cobra"

	ukafkasdk "github.com/ucloud/ucloud-sdk-go/services/ukafka"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// GetUKafkaNodeTypeResponse 自定义响应结构
type GetUKafkaNodeTypeResponse struct {
	RetCode     int            `json:"RetCode"`
	Action      string         `json:"Action"`
	NodeTypeSet []InstanceType `json:"NodeTypeSet"`
	TotalCount  int            `json:"TotalCount"`
	Message     string         `json:"Message"`
}

// InstanceType 机型信息
type InstanceType struct {
	NodeTypeName   string    `json:"NodeTypeName"`
	CPU            int       `json:"CPU"`
	Memory         int       `json:"Memory"`
	DiskType       string    `json:"DiskType"`
	DiskSet        []DiskSet `json:"DiskSet"`
	MaxDiskSize    int       `json:"MaxDiskSize"`
	MinDiskSize    int       `json:"MinDiskSize"`
	IsOpenSecGroup bool      `json:"IsOpenSecGroup"`
	UHostFamily    string    `json:"UHostFamily"`
}

// DiskSet 磁盘配置
type DiskSet struct {
	Type string `json:"Type"`
	Size int    `json:"Size"`
}

// newNodeConf ucloud ukafka node-conf
func newNodeConf(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, ukafkasdk.NewClient)
	req := client.NewGetUKafkaNodeTypeRequest()
	cmd := &cobra.Command{
		Use:   "node-conf",
		Short: "List available UKafka node configurations",
		Long:  "List available UKafka node configurations",
		Run: func(cmd *cobra.Command, args []string) {
			// 使用 GenericInvoke 绕过 SDK 类型问题
			genReq := client.Client.NewGenericRequest()
			genReq.SetAction("GetUKafkaNodeType")
			genReq.SetRegion(*req.Region)
			genReq.SetZone(*req.Zone)
			if req.ProjectId != nil && *req.ProjectId != "" {
				genReq.SetProjectId(*req.ProjectId)
			}
			if req.NodeType != nil && *req.NodeType != "" {
				payload := map[string]interface{}{
					"NodeType": *req.NodeType,
				}
				genReq.SetPayload(payload)
			}

			genResp, err := client.Client.GenericInvoke(genReq)
			if err != nil {
				ctx.HandleError(err)
				return
			}

			var resp GetUKafkaNodeTypeResponse
			if err := genResp.Unmarshal(&resp); err != nil {
				ctx.HandleError(fmt.Errorf("parse response: %w", err))
				return
			}

			if resp.RetCode != 0 {
				ctx.HandleError(fmt.Errorf("API error: RetCode=%d, Message=%s", resp.RetCode, resp.Message))
				return
			}

			list := []NodeConfRow{}
			for _, t := range resp.NodeTypeSet {
				row := NodeConfRow{
					NodeType:    t.NodeTypeName,
					CPU:         fmt.Sprintf("%d", t.CPU),
					Memory:      fmt.Sprintf("%dMB", t.Memory),
					DiskType:    t.DiskType,
					MinDiskSize: fmt.Sprintf("%d", t.MinDiskSize),
					MaxDiskSize: fmt.Sprintf("%d", t.MaxDiskSize),
					SecGroup:    fmt.Sprintf("%v", t.IsOpenSecGroup),
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
	req.Zone = flags.String("zone", ctx.DefaultZone(), "Optional. Assign availability zone")
	req.NodeType = flags.String("node-type", "", "Optional. Specify node type")
	return cmd
}
