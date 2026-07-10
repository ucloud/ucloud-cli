package topic

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/urocketmq"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
	"github.com/ucloud/ucloud-cli/products/urocketmq/internal/service"
)

// newCreate ucloud urocketmq topic create
func newCreate(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, urocketmq.NewClient)
	req := client.NewCreateURocketMQTopicRequest()
	cmd := &cobra.Command{
		Use:          "create",
		Short:        "Create URocketMQ topic",
		Long:         "Create URocketMQ topic",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := client.CreateURocketMQTopic(req)
			if err != nil {
				return err
			}
			ctx.EmitResult(cli.OpResultRow{ResourceID: resp.TopicId, Action: "create", Status: "Success"})
			return nil
		},
	}
	cmd.Flags().SortFlags = false

	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)
	req.MessageType = cmd.Flags().String("message-type", "", "Required. Message type. Accept values: Normal, PartitionSequence, GlobalSequence, Transaction, Delay")
	req.Name = cmd.Flags().String("name", "", "Required. Topic name, supports letters, digits, hyphens and underscores, length 1~36")
	req.ServiceId = cmd.Flags().String("service-id", "", "Required. Service ID. see 'ucloud urocketmq service list'")
	req.Remark = cmd.Flags().String("remark", "", "Optional. Topic remark, max length 36")

	command.SetFlagValues(cmd, "message-type", "Normal", "PartitionSequence", "GlobalSequence", "Transaction", "Delay")
	command.SetCompletion(cmd, "service-id", func() []string {
		return service.ServiceList(ctx, *req.ProjectId, *req.Region)
	})
	cmd.MarkFlagRequired("message-type")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("service-id")

	return cmd
}
