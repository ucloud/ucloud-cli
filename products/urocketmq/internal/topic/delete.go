package topic

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/urocketmq"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
	"github.com/ucloud/ucloud-cli/products/urocketmq/internal/service"
)

// newDelete ucloud urocketmq topic delete
func newDelete(ctx *cli.Context) *cobra.Command {
	var yes *bool
	client := cli.NewServiceClient(ctx, urocketmq.NewClient)
	req := client.NewDeleteURocketMQTopicRequest()
	cmd := &cobra.Command{
		Use:          "delete",
		Short:        "Delete URocketMQ topic",
		Long:         "Delete URocketMQ topic",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ok, err := ctx.Confirm(*yes, fmt.Sprintf("Are you sure you want to delete topic %q from service %s?", *req.TopicName, *req.ServiceId))
			if err != nil {
				return err
			}
			if !ok {
				return nil
			}
			_, err = client.DeleteURocketMQTopic(req)
			if err != nil {
				return err
			}
			ctx.EmitResult(cli.OpResultRow{ResourceID: *req.TopicName, Action: "delete", Status: "Success"})
			return nil
		},
	}
	cmd.Flags().SortFlags = false

	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)
	req.ServiceId = cmd.Flags().String("service-id", "", "Required. Service ID. see 'ucloud urocketmq service list'")
	req.TopicName = cmd.Flags().String("topic-name", "", "Required. Topic name")
	yes = cmd.Flags().BoolP("yes", "y", false, "Optional. Do not prompt for confirmation.")

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
