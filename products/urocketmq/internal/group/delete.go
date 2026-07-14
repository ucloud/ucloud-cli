package group

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/urocketmq"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
	"github.com/ucloud/ucloud-cli/products/urocketmq/internal/service"
)

// newDelete ucloud urocketmq group delete
func newDelete(ctx *cli.Context) *cobra.Command {
	var yes bool
	client := cli.NewServiceClient(ctx, urocketmq.NewClient)
	req := client.NewDeleteURocketMQGroupRequest()
	cmd := &cobra.Command{
		Use:          "delete",
		Short:        "Delete a consumer group",
		Long:         "Delete a consumer group",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ok, err := ctx.Confirm(yes, fmt.Sprintf("Are you sure you want to delete group %q?", *req.GroupName))
			if err != nil {
				return err
			}
			if !ok {
				return nil
			}
			_, err = client.DeleteURocketMQGroup(req)
			if err != nil {
				return err
			}
			ctx.EmitResult(cli.OpResultRow{
				ResourceID: *req.GroupName,
				Action:     "delete",
				Status:     "Deleted",
			})
			return nil
		},
	}
	cmd.Flags().SortFlags = false

	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)
	req.ServiceId = cmd.Flags().String("service-id", "", "Required. Service ID")
	req.GroupName = cmd.Flags().String("group-name", "", "Required. Consumer group name")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Optional. Do not prompt for confirmation.")

	command.SetCompletion(cmd, "service-id", func() []string {
		return service.ServiceList(ctx, *req.ProjectId, *req.Region)
	})
	command.SetCompletion(cmd, "group-name", func() []string {
		return GroupList(ctx, *req.ProjectId, *req.Region, *req.ServiceId)
	})
	cmd.MarkFlagRequired("service-id")
	cmd.MarkFlagRequired("group-name")

	return cmd
}
