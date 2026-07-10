package message

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	urocketmq "github.com/ucloud/ucloud-sdk-go/services/urocketmq"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
	"github.com/ucloud/ucloud-cli/products/urocketmq/internal/service"
	"github.com/ucloud/ucloud-cli/products/urocketmq/internal/topic"
)

// newQueryByTopic ucloud urocketmq message query-by-topic
func newQueryByTopic(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, urocketmq.NewClient)
	req := client.NewQueryURocketMQMessageByTopicRequest()
	cmd := &cobra.Command{
		Use:   "query-by-topic",
		Short: "Query messages by topic and time range",
		Long:  `Query messages by topic and time range. Time format: {2006-01-02 15:04:05}`,
		Run: func(cmd *cobra.Command, args []string) {
			beginStr, _ := cmd.Flags().GetString("begin")
			endStr, _ := cmd.Flags().GetString("end")
			const layout = "2006-01-02 15:04:05"
			beginTime, err := time.ParseInLocation(layout, beginStr, time.Local)
			if err != nil {
				ctx.HandleError(fmt.Errorf("invalid begin time %q, expected format: 2006-01-02 15:04:05", beginStr))
				return
			}
			endTime, err := time.ParseInLocation(layout, endStr, time.Local)
			if err != nil {
				ctx.HandleError(fmt.Errorf("invalid end time %q, expected format: 2006-01-02 15:04:05", endStr))
				return
			}
			beginUnix := int(beginTime.Unix())
			endUnix := int(endTime.Unix())
			req.Begin = &beginUnix
			req.End = &endUnix
			resp, err := client.QueryURocketMQMessageByTopic(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			rows := toMessageBaseInfoRows(resp.MessageList)
			printMessageRows(ctx, rows)
		},
	}
	cmd.Flags().SortFlags = false

	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)

	cmd.Flags().String("begin", "", "Required. Begin time (YYYY-MM-DD HH:MM:SS)")
	cmd.Flags().String("end", "", "Required. End time (YYYY-MM-DD HH:MM:SS)")
	req.ServiceId = cmd.Flags().String("service-id", "", "Required. Service ID")
	req.TopicName = cmd.Flags().String("topic-name", "", "Required. Topic name")
	_ = cmd.MarkFlagRequired("begin")
	_ = cmd.MarkFlagRequired("end")
	_ = cmd.MarkFlagRequired("service-id")
	_ = cmd.MarkFlagRequired("topic-name")

	command.SetCompletion(cmd, "service-id", func() []string {
		return service.ServiceList(ctx, *req.ProjectId, *req.Region)
	})
	command.SetCompletion(cmd, "topic-name", func() []string {
		return topic.TopicList(ctx, *req.ProjectId, *req.Region, *req.ServiceId)
	})

	return cmd
}
