package topic

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/urocketmq"

	"github.com/ucloud/ucloud-cli/internal/common"
	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
	"github.com/ucloud/ucloud-cli/products/urocketmq/internal/service"
)

// newList ucloud urocketmq topic list
func newList(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, urocketmq.NewClient)
	req := client.NewListURocketMQTopicRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List URocketMQ topics",
		Long:  "List URocketMQ topics",
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := client.ListURocketMQTopic(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			listTopic(ctx, resp.TopicList)
		},
	}
	cmd.Flags().SortFlags = false

	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)
	req.Limit = cmd.Flags().Int("limit", 20, "Optional. Limit default 20, max value 1000")
	req.Offset = cmd.Flags().Int("offset", 0, "Optional. Offset default 0")
	req.ServiceId = cmd.Flags().String("service-id", "", "Required. Service ID. see 'ucloud urocketmq service list'")

	command.SetCompletion(cmd, "service-id", func() []string {
		return service.ServiceList(ctx, *req.ProjectId, *req.Region)
	})
	cmd.MarkFlagRequired("service-id")

	return cmd
}

// listTopic renders the topic list. json/yaml emits full-field topicRow; table mode uses curated columns topicRowDefault.
func listTopic(ctx *cli.Context, topics []urocketmq.TopicInfo) {
	list := make([]topicRow, 0, len(topics))
	for _, t := range topics {
		list = append(list, toTopicRow(t))
	}

	if ctx.Format() != cli.OutputTable {
		ctx.PrintList(list)
		return
	}

	rows := make([]topicRowDefault, 0, len(list))
	for _, r := range list {
		rows = append(rows, topicRowDefault{
			TopicName:   r.TopicName,
			MessageType: r.MessageType,
			Remark:      r.Remark,
			CreateTime:  common.FormatDate(r.CreateTime),
		})
	}
	ctx.PrintList(rows)
}

// toTopicRow maps SDK TopicInfo to a full-field row.
func toTopicRow(t urocketmq.TopicInfo) topicRow {
	return topicRow{
		TopicId:     t.TopicId,
		TopicName:   t.TopicName,
		MessageType: t.MessageType,
		Remark:      t.Remark,
		CreateTime:  t.CreateTime,
	}
}
