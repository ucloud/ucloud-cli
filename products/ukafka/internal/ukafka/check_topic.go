package ukafka

import (
	"fmt"

	"github.com/spf13/cobra"

	ukafkasdk "github.com/ucloud/ucloud-sdk-go/services/ukafka"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// IsUKafkaTopicNameExistResponse 自定义响应结构
type IsUKafkaTopicNameExistResponse struct {
	RetCode int    `json:"RetCode"`
	Action  string `json:"Action"`
	IsExist string `json:"IsExist"`
	Message string `json:"Message"`
}

// newCheckTopic ucloud ukafka check-topic
func newCheckTopic(ctx *cli.Context) *cobra.Command {
	var instanceID *string
	var topicName *string

	client := cli.NewServiceClient(ctx, ukafkasdk.NewClient)
	req := client.NewIsUKafkaTopicNameExistRequest()

	cmd := &cobra.Command{
		Use:   "check-topic",
		Short: "Check if a topic name exists in UKafka instance",
		Long:  "Check if a topic name exists in UKafka instance",
		Run: func(cmd *cobra.Command, args []string) {
			// 使用 GenericInvoke 调用 API
			genReq := client.Client.NewGenericRequest()
			genReq.SetAction("IsUKafkaTopicNameExist")
			genReq.SetRegion(*req.Region)
			genReq.SetZone(*req.Zone)
			if req.ProjectId != nil && *req.ProjectId != "" {
				genReq.SetProjectId(*req.ProjectId)
			}

			payload := map[string]interface{}{
				"ClusterInstanceId": *instanceID,
				"TopicName":         *topicName,
			}
			genReq.SetPayload(payload)

			genResp, err := client.Client.GenericInvoke(genReq)
			if err != nil {
				ctx.HandleError(err)
				return
			}

			var resp IsUKafkaTopicNameExistResponse
			if err := genResp.Unmarshal(&resp); err != nil {
				ctx.HandleError(fmt.Errorf("parse response: %w", err))
				return
			}

			if resp.RetCode != 0 {
				ctx.HandleError(fmt.Errorf("API error: RetCode=%d, Message=%s", resp.RetCode, resp.Message))
				return
			}

			// 输出结构化结果
			rows := []cli.DescribeRow{
				{Attribute: "InstanceID", Content: *instanceID},
				{Attribute: "TopicName", Content: *topicName},
				{Attribute: "Exists", Content: resp.IsExist},
			}
			ctx.PrintList(rows)

			// 同时输出事件结果
			ctx.EmitResult(cli.OpResultRow{
				ResourceID: *instanceID,
				Action:     "check-topic",
				Status:     resp.IsExist,
			})
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	instanceID = flags.String("ukafka-id", "", "Required. Instance ID")
	topicName = flags.String("topic-name", "", "Required. Topic name to check")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	req.Zone = flags.String("zone", ctx.DefaultZone(), "Optional. Assign availability zone")

	cmd.MarkFlagRequired("ukafka-id")
	cmd.MarkFlagRequired("topic-name")

	return cmd
}
