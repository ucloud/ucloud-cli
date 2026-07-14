package service

import (
	"github.com/spf13/cobra"

	urocketmq "github.com/ucloud/ucloud-sdk-go/services/urocketmq"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newUpdateRemark ucloud urocketmq service update-remark
func newUpdateRemark(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, urocketmq.NewClient)
	req := client.NewUpdateURocketMQServiceRemarkRequest()
	cmd := &cobra.Command{
		Use:   "update-remark",
		Short: "Update URocketMQ service instance remark",
		Long:  "Update URocketMQ service instance remark",
		RunE: func(cmd *cobra.Command, args []string) error {
			serviceID := *req.ServiceId
			_, err := client.UpdateURocketMQServiceRemark(req)
			if err != nil {
				return err
			}
			ctx.EmitResult(cli.OpResultRow{ResourceID: serviceID, Action: "update-remark", Status: "Updated"})
			return nil
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.ServiceId = flags.String("service-id", "", "Required. Service ID")
	req.Remark = flags.String("remark", "", "Optional. New remark for the service")

	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)

	command.SetCompletion(cmd, "service-id", func() []string {
		return ServiceList(ctx, req.GetProjectId(), req.GetRegion())
	})

	cmd.MarkFlagRequired("service-id")

	return cmd
}
