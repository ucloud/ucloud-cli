package topic

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/urocketmq"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
	"github.com/ucloud/ucloud-cli/products/urocketmq/internal/service"
)

// newUpdate ucloud urocketmq topic update
func newUpdate(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, urocketmq.NewClient)
	req := client.NewUpdateURocketMQTopicRequest()
	cmd := &cobra.Command{
		Use:          "update",
		Short:        "Update URocketMQ topic remark",
		Long:         "Update URocketMQ topic remark. Currently only supports updating the topic remark.",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := client.UpdateURocketMQTopic(req)
			if err != nil {
				return err
			}
			ctx.EmitResult(cli.OpResultRow{ResourceID: *req.TopicName, Action: "update", Status: "Success"})
			return nil
		},
	}
	cmd.Flags().SortFlags = false

	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)
	req.ServiceId = cmd.Flags().String("service-id", "", "Required. Service ID. see 'ucloud urocketmq service list'")
	req.TopicName = cmd.Flags().String("topic-name", "", "Required. Topic name")
	req.Remark = cmd.Flags().String("remark", "", "Optional. Topic remark")

	command.SetCompletion(cmd, "service-id", func() []string {
		return service.ServiceList(ctx, *req.ProjectId, *req.Region)
	})
	command.SetCompletion(cmd, "topic-name", func() []string {
		return TopicList(ctx, *req.ProjectId, *req.Region, *req.ServiceId)
	})
	cmd.MarkFlagRequired("service-id")
	cmd.MarkFlagRequired("topic-name")

	return cmd
}
