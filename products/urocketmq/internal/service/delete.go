package service

import (
	"fmt"

	"github.com/spf13/cobra"

	urocketmq "github.com/ucloud/ucloud-sdk-go/services/urocketmq"
	uerr "github.com/ucloud/ucloud-sdk-go/ucloud/error"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newDelete ucloud urocketmq service delete
func newDelete(ctx *cli.Context) *cobra.Command {
	var async bool
	var yes bool
	client := cli.NewServiceClient(ctx, urocketmq.NewClient)
	req := client.NewDeleteURocketMQServiceRequest()
	cmd := &cobra.Command{
		Use:          "delete",
		Short:        "Delete a URocketMQ service instance",
		Long:         "Delete a URocketMQ service instance",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			serviceID := *req.ServiceId
			ok, err := ctx.Confirm(yes, fmt.Sprintf("Are you sure you want to delete service %q?", serviceID))
			if err != nil {
				return err
			}
			if !ok {
				return nil
			}

			resp, err := client.DeleteURocketMQService(req)
			if err != nil {
				if serr, ok := err.(uerr.ServerError); ok && serr.Code() == 99539 {
					return fmt.Errorf("please delete all Topics and Groups under the instance before deleting it")
				}
				return err
			}

			prog := ctx.NewProgress()
			block := prog.NewBlock()
			ctx.EmitResult(cli.OpResultRow{ResourceID: serviceID, Action: "delete", Status: "Deleting"})

			_ = resp // delete response contains only Message

			text := fmt.Sprintf("the service[%s] is deleting", serviceID)
			if async {
				block.Append(text)
			} else {
				prog.Sspoll(describeDeletedServiceByID(ctx), serviceID, text,
					// Poll until service no longer exists (describe returns nil) or deletion fails.
					[]string{SERVICE_DELETED, SERVICE_DELETE_FAILED}, block, &req.CommonBase)
			}
			return nil
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.ServiceId = flags.String("service-id", "", "Required. Service ID")
	flags.BoolVar(&async, "async", false, "Optional. Do not wait for the long-running operation to finish.")
	flags.BoolVarP(&yes, "yes", "y", false, "Optional. Do not prompt for confirmation.")

	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)

	command.SetCompletion(cmd, "service-id", func() []string {
		return ServiceList(ctx, req.GetProjectId(), req.GetRegion())
	})

	cmd.MarkFlagRequired("service-id")

	return cmd
}

func describeDeletedServiceByID(ctx *cli.Context) func(string, *request.CommonBase) (interface{}, error) {
	describe := describeServiceByID(ctx)
	return func(id string, commonBase *request.CommonBase) (interface{}, error) {
		inst, err := describe(id, commonBase)
		if err != nil {
			return nil, err
		}
		if inst == nil {
			return &urocketmq.ServiceDetail{State: SERVICE_DELETED}, nil
		}
		return inst, nil
	}
}
