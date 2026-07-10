package message

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/urocketmq"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
	"github.com/ucloud/ucloud-cli/products/urocketmq/internal/service"
	"github.com/ucloud/ucloud-cli/products/urocketmq/internal/topic"
)

// newQueryByKey ucloud urocketmq message query-by-key
func newQueryByKey(ctx *cli.Context) *cobra.Command {
	var idOnly bool
	client := cli.NewServiceClient(ctx, urocketmq.NewClient)
	req := client.NewQueryURocketMQMessageByKeyRequest()
	cmd := &cobra.Command{
		Use:   "query-by-key",
		Short: "Query messages by key",
		Long:  `Query messages by a custom message key under a specific topic.`,
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := client.QueryURocketMQMessageByKey(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if idOnly {
				listMessageID(ctx, resp.MessageList)
				return
			}
			rows := toMessageBaseInfoRows(resp.MessageList)
			printMessageRows(ctx, rows)
		},
	}
	cmd.Flags().SortFlags = false

	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)

	req.Key = cmd.Flags().String("key", "", "Required. Message key to query")
	req.ServiceId = cmd.Flags().String("service-id", "", "Required. Service ID")
	req.TopicName = cmd.Flags().String("topic-name", "", "Required. Topic name")
	cmd.Flags().BoolVar(&idOnly, "id-only", false, "Optional. Only display message IDs")
	_ = cmd.MarkFlagRequired("key")
	_ = cmd.MarkFlagRequired("service-id")
	_ = cmd.MarkFlagRequired("topic-name")

	command.SetFlagValues(cmd, "id-only", "true", "false")
	command.SetCompletion(cmd, "service-id", func() []string {
		return service.ServiceList(ctx, *req.ProjectId, *req.Region)
	})
	command.SetCompletion(cmd, "topic-name", func() []string {
		return topic.TopicList(ctx, *req.ProjectId, *req.Region, *req.ServiceId)
	})

	return cmd
}

// listMessageID outputs a message ID list (--id-only mode), writing to ctx.Out() to avoid capture by
// ProgressWriter which would yield empty output in non-TTY mode.
func listMessageID(ctx *cli.Context, infos []urocketmq.MessageBaseInfo) {
	ids := make([]string, 0, len(infos))
	for _, info := range infos {
		ids = append(ids, info.MsgId)
	}
	fmt.Fprintln(ctx.Out(), strings.Join(ids, ","))
}
