package message

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/urocketmq"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
	"github.com/ucloud/ucloud-cli/products/urocketmq/internal/service"
	"github.com/ucloud/ucloud-cli/products/urocketmq/internal/topic"
)

// newQueryByID ucloud urocketmq message query-by-id
func newQueryByID(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, urocketmq.NewClient)
	req := client.NewQueryURocketMQMessageByIDRequest()
	cmd := &cobra.Command{
		Use:   "query-by-id",
		Short: "Query a message by ID",
		Long:  `Query a message by ID, returns the full message detail including body.`,
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := client.QueryURocketMQMessageByID(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			rows := toMessageDetailRows(resp.MessageList)
			printMessageRows(ctx, rows)
		},
	}
	cmd.Flags().SortFlags = false

	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)

	req.MsgId = cmd.Flags().String("msg-id", "", "Required. Message ID to query")
	req.ServiceId = cmd.Flags().String("service-id", "", "Required. Service ID")
	req.TopicName = cmd.Flags().String("topic-name", "", "Required. Topic name")
	_ = cmd.MarkFlagRequired("msg-id")
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
