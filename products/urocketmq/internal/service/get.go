package service

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/urocketmq"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newGet ucloud urocketmq service get
func newGet(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, urocketmq.NewClient)
	req := client.NewGetURocketMQServiceRequest()
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get details of a URocketMQ service instance",
		Long:  "Get details of a URocketMQ service instance",
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := client.GetURocketMQService(req)
			if err != nil {
				return err
			}
			ctx.PrintList(resp.ServiceList)
			return nil
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.ServiceId = flags.String("service-id", "", "Required. Service ID")

	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)

	command.SetCompletion(cmd, "service-id", func() []string {
		return ServiceList(ctx, req.GetProjectId(), req.GetRegion())
	})

	cobra.CheckErr(cmd.MarkFlagRequired("service-id"))

	return cmd
}
