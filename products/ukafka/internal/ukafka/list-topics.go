package ukafka

import (
	"fmt"

	"github.com/spf13/cobra"

	ukafkasdk "github.com/ucloud/ucloud-sdk-go/services/ukafka"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// ListUKafkaTopicsResponse 自定义响应结构
type ListUKafkaTopicsResponse struct {
	RetCode  int        `json:"RetCode"`
	Action   string     `json:"Action"`
	TopicList []TopicInfo `json:"TopicList"`
	Length   int        `json:"Length"`
	Message  string     `json:"Message"`
}

// TopicInfo topic信息
type TopicInfo struct {
	Topic             string `json:"Topic"`
	NumOfPartition    int    `json:"NumOfPartition"`
	NumOfOccupyBroker int    `json:"NumOfOccupyBroker"`
	NumOfReplica      int    `json:"NumOfReplica"`
	Status            string `json:"Status"`
	UnderReplicasPer  string `json:"UnderReplicasPer"`
}

// newListTopics ucloud ukafka list-topics
func newListTopics(ctx *cli.Context) *cobra.Command {
	var instanceID *string

	client := cli.NewServiceClient(ctx, ukafkasdk.NewClient)
	req := client.NewListUKafkaTopicsRequest()

	cmd := &cobra.Command{
		Use:   "list-topics",
		Short: "List Kafka topics in UKafka instance",
		Long:  "List Kafka topics in UKafka instance",
		Run: func(cmd *cobra.Command, args []string) {
			// 使用 GenericInvoke 绕过 SDK 类型问题
			genReq := client.Client.NewGenericRequest()
			genReq.SetAction("ListUKafkaTopics")
			genReq.SetRegion(*req.Region)
			genReq.SetZone(*req.Zone)
			if req.ProjectId != nil && *req.ProjectId != "" {
				genReq.SetProjectId(*req.ProjectId)
			}

			payload := map[string]interface{}{
				"ClusterInstanceId": *instanceID,
			}
			genReq.SetPayload(payload)

			genResp, err := client.Client.GenericInvoke(genReq)
			if err != nil {
				ctx.HandleError(err)
				return
			}

			var resp ListUKafkaTopicsResponse
			if err := genResp.Unmarshal(&resp); err != nil {
				ctx.HandleError(fmt.Errorf("parse response: %w", err))
				return
			}

			if resp.RetCode != 0 {
				ctx.HandleError(fmt.Errorf("API error: RetCode=%d, Message=%s", resp.RetCode, resp.Message))
				return
			}

			list := []TopicRow{}
			for _, t := range resp.TopicList {
				row := TopicRow{
					Topic:             t.Topic,
					NumOfPartition:    fmt.Sprintf("%d", t.NumOfPartition),
					NumOfReplica:      fmt.Sprintf("%d", t.NumOfReplica),
					NumOfOccupyBroker: fmt.Sprintf("%d", t.NumOfOccupyBroker),
					UnderReplicasPer:  t.UnderReplicasPer,
					Status:            t.Status,
				}
				list = append(list, row)
			}
			ctx.PrintList(list)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	instanceID = flags.String("instance-id", "", "Required. Instance ID")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	req.Zone = flags.String("zone", ctx.DefaultZone(), "Optional. Assign availability zone")

	cmd.MarkFlagRequired("instance-id")

	return cmd
}
