package token

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/urocketmq"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
	"github.com/ucloud/ucloud-cli/products/urocketmq/internal/service"
)

// newDelete ucloud urocketmq token delete
func newDelete(ctx *cli.Context) *cobra.Command {
	var yes bool
	client := cli.NewServiceClient(ctx, urocketmq.NewClient)
	req := client.NewDeleteURocketMQTokenRequest()
	cmd := &cobra.Command{
		Use:          "delete",
		Short:        "Delete URocketMQ token",
		Long:         "Delete URocketMQ token. Default token cannot be deleted.",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ok, err := ctx.Confirm(yes, fmt.Sprintf("Are you sure you want to delete token %q from service %s?", *req.TokenId, *req.ServiceId))
			if err != nil {
				return err
			}
			if !ok {
				return nil
			}

			_, err = client.DeleteURocketMQToken(req)
			if err != nil {
				return err
			}
			ctx.EmitResult(cli.OpResultRow{ResourceID: *req.TokenId, Action: "delete", Status: "Deleted"})
			return nil
		},
	}
	cmd.Flags().SortFlags = false

	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)
	req.ServiceId = cmd.Flags().String("service-id", "", "Required. Service ID. See 'ucloud urocketmq service list'")
	req.TokenId = cmd.Flags().String("token-id", "", "Required. Token ID")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Optional. Do not prompt for confirmation.")

	command.SetCompletion(cmd, "service-id", func() []string {
		return service.ServiceList(ctx, *req.ProjectId, *req.Region)
	})
	command.SetCompletion(cmd, "token-id", func() []string {
		return TokenList(ctx, *req.ProjectId, *req.Region, *req.ServiceId)
	})
	cmd.MarkFlagRequired("service-id")
	cmd.MarkFlagRequired("token-id")

	return cmd
}
