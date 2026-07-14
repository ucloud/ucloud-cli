package token

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/urocketmq"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
	"github.com/ucloud/ucloud-cli/products/urocketmq/internal/service"
	"github.com/ucloud/ucloud-cli/products/urocketmq/internal/topic"
)

// newUpdate ucloud urocketmq token update
func newUpdate(ctx *cli.Context) *cobra.Command {
	var allowConsumeTopicList []string
	var allowProduceTopicList []string
	client := cli.NewServiceClient(ctx, urocketmq.NewClient)
	req := client.NewUpdateURocketMQTokenRequest()
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update URocketMQ token configuration",
		Long:  "Update URocketMQ token configuration",
		Run: func(cmd *cobra.Command, args []string) {
			if len(allowConsumeTopicList) > 0 {
				req.AllowConsumeTopicList = sdk.String(strings.Join(allowConsumeTopicList, ","))
			}
			if len(allowProduceTopicList) > 0 {
				req.AllowProduceTopicList = sdk.String(strings.Join(allowProduceTopicList, ","))
			}
			_, err := client.UpdateURocketMQToken(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			ctx.EmitResult(cli.OpResultRow{ResourceID: *req.TokenId, Action: "update", Status: "Success"})
		},
	}
	cmd.Flags().SortFlags = false

	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)
	req.ServiceId = cmd.Flags().String("service-id", "", "Required. Service ID. See 'ucloud urocketmq service list'")
	req.TokenId = cmd.Flags().String("token-id", "", "Required. Token ID")
	req.TopicConsumePerm = cmd.Flags().String("topic-consume-perm", "", "Required. Topic consume permission. Accept values: ALL, NONE, PART")
	req.TopicProducePerm = cmd.Flags().String("topic-produce-perm", "", "Required. Topic produce permission. Accept values: ALL, NONE, PART")
	cmd.Flags().StringSliceVar(&allowConsumeTopicList, "allow-consume-topic-list", nil, "Optional. Allow consume topic name list, multiple values separated by comma")
	cmd.Flags().StringSliceVar(&allowProduceTopicList, "allow-produce-topic-list", nil, "Optional. Allow produce topic name list, multiple values separated by comma")

	command.SetFlagValues(cmd, "topic-consume-perm", "ALL", "NONE", "PART")
	command.SetFlagValues(cmd, "topic-produce-perm", "ALL", "NONE", "PART")
	command.SetCompletion(cmd, "service-id", func() []string {
		return service.ServiceList(ctx, *req.ProjectId, *req.Region)
	})
	command.SetCompletion(cmd, "token-id", func() []string {
		return TokenList(ctx, *req.ProjectId, *req.Region, *req.ServiceId)
	})
	command.SetCompletion(cmd, "allow-consume-topic-list", func() []string {
		return topic.TopicList(ctx, *req.ProjectId, *req.Region, *req.ServiceId)
	})
	command.SetCompletion(cmd, "allow-produce-topic-list", func() []string {
		return topic.TopicList(ctx, *req.ProjectId, *req.Region, *req.ServiceId)
	})
	cmd.MarkFlagRequired("service-id")
	cmd.MarkFlagRequired("token-id")
	cmd.MarkFlagRequired("topic-consume-perm")
	cmd.MarkFlagRequired("topic-produce-perm")

	return cmd
}
