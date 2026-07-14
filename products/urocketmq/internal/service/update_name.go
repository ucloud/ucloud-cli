package service

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/urocketmq"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newUpdateName ucloud urocketmq service update-name
func newUpdateName(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, urocketmq.NewClient)
	req := client.NewUpdateURocketMQServiceNameRequest()
	cmd := &cobra.Command{
		Use:   "update-name",
		Short: "Update URocketMQ service instance name",
		Long:  "Update URocketMQ service instance name",
		RunE: func(cmd *cobra.Command, args []string) error {
			serviceID := *req.ServiceId
			_, err := client.UpdateURocketMQServiceName(req)
			if err != nil {
				return err
			}
			ctx.EmitResult(cli.OpResultRow{ResourceID: serviceID, Action: "update-name", Status: "Updated"})
			return nil
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.ServiceId = flags.String("service-id", "", "Required. Service ID")
	req.Name = flags.String("name", "", "Required. New service name. Regex: ^[a-zA-Z0-9-_]{1,36}$")

	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)

	command.SetCompletion(cmd, "service-id", func() []string {
		return ServiceList(ctx, req.GetProjectId(), req.GetRegion())
	})

	cmd.MarkFlagRequired("service-id")
	cmd.MarkFlagRequired("name")

	return cmd
}
