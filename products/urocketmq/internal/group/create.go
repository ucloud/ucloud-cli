package group

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/urocketmq"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
	"github.com/ucloud/ucloud-cli/products/urocketmq/internal/service"
)

// newCreate ucloud urocketmq group create
func newCreate(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, urocketmq.NewClient)
	req := client.NewCreateURocketMQGroupRequest()
	cmd := &cobra.Command{
		Use:          "create",
		Short:        "Create a consumer group",
		Long:         "Create a consumer group",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := client.CreateURocketMQGroup(req)
			if err != nil {
				return err
			}
			ctx.EmitResult(cli.OpResultRow{
				ResourceID: resp.GroupId,
				Action:     "create",
				Status:     "Created",
			})
			return nil
		},
	}
	cmd.Flags().SortFlags = false

	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)
	req.Name = cmd.Flags().String("name", "", "Required. Consumer group name, 1-36 characters, supports letters, digits, - and _")
	req.ServiceId = cmd.Flags().String("service-id", "", "Required. Service ID")
	req.Remark = cmd.Flags().String("remark", "", "Optional. Group remark")

	command.SetCompletion(cmd, "service-id", func() []string {
		return service.ServiceList(ctx, *req.ProjectId, *req.Region)
	})
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("service-id")

	return cmd
}
