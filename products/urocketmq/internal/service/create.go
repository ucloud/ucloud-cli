package service

import (
	"fmt"

	"github.com/spf13/cobra"

	urocketmq "github.com/ucloud/ucloud-sdk-go/services/urocketmq"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newCreate ucloud urocketmq service create
func newCreate(ctx *cli.Context) *cobra.Command {
	var async bool
	client := cli.NewServiceClient(ctx, urocketmq.NewClient)
	req := client.NewCreateURocketMQServiceRequest()
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a URocketMQ service instance",
		Long:  "Create a URocketMQ service instance",
		RunE: func(cmd *cobra.Command, args []string) error {
			req.VPCId = sdk.String(ctx.PickResourceID(*req.VPCId))
			req.SubnetId = sdk.String(ctx.PickResourceID(*req.SubnetId))
			if *req.Storage <= 0 || *req.Storage%100 != 0 {
				return fmt.Errorf("--storage-gb must be a positive multiple of 100")
			}
			if sdk.StringValue(req.ChargeType) == "Dynamic" {
				req.Quantity = sdk.Int(0)
			}
			resp, err := client.CreateURocketMQService(req)
			if err != nil {
				return err
			}

			serviceID := resp.ServiceId
			prog := ctx.NewProgress()
			block := prog.NewBlock()
			ctx.EmitResult(cli.OpResultRow{ResourceID: serviceID, Action: "create", Status: "Initializing"})

			text := fmt.Sprintf("the service[%s] is initializing", serviceID)
			if async {
				block.Append(text)
			} else {
				prog.Sspoll(describeServiceByID(ctx), serviceID, text,
					[]string{SERVICE_AVAILABLE, SERVICE_CREATE_FAILED}, block, &req.CommonBase)
			}
			return nil
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.ChargeType = flags.String("charge-type", "Month", "Required. Charge type. Enum: Year, Month, Dynamic")
	req.Edition = flags.String("edition", "Enterprise", "Required. Edition. Unique value: Enterprise")
	req.Mode = flags.String("mode", "PrivateNet", "Required. Network mode. Unique value: PrivateNet")
	req.Name = flags.String("name", "", "Required. Service name. Regex: ^[a-zA-Z0-9-_]{1,36}$")
	req.PublicVersion = flags.String("public-version", "", "Cluster version. Options vary by region, see doc for supported values: https://github.com/UCloudDoc-Team/rocketmq/blob/master/price/index.md, e.g. v4, v5 (each region only support one version)")
	req.Storage = flags.Int("storage-gb", 0, "Required. Storage space in GB. Check the doc first to determine available values: https://github.com/UCloudDoc-Team/rocketmq/blob/master/price/index.md")
	req.SubnetId = flags.String("subnet-id", "", "Required. Subnet ID. Default to current region's default subnet")
	req.Tps = flags.String("tps", "", "Required. Transactions per second. Enum: 10000, 20000, 50000, 100000, 200000. Note: v4 supports 20000, 50000, 100000, 200000; v5 currently supports only 10000, 20000.")
	req.VPCId = flags.String("vpc-id", "", "Required. VPC ID. Default to current region's default VPC")
	req.FileReservedTime = flags.String("file-reserved-time", "3", "Optional. Message reserved time in days, default 3")
	req.Quantity = flags.Int("quantity", 1, "Optional. Purchase duration in months. Month: 1-9(month), 0=until end of current month; Dynamic: ignore; Year: use --quantity as years")
	req.Remark = flags.String("remark", "", "Optional. Remark")
	req.Tag = flags.String("group", "Default", "Optional. Business group tag")

	flags.BoolVar(&async, "async", false, "Optional. Do not wait for the long-running operation to finish.")

	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)

	command.SetFlagValues(cmd, "charge-type", "Year", "Month", "Dynamic")
	command.SetFlagValues(cmd, "edition", "Enterprise")
	command.SetFlagValues(cmd, "mode", "PrivateNet")
	command.SetFlagValues(cmd, "tps", "10000", "20000", "50000", "100000", "200000")

	command.SetCompletion(cmd, "vpc-id", func() []string {
		return getAllVPCIdNames(ctx, req.GetProjectId(), req.GetRegion())
	})
	command.SetCompletion(cmd, "subnet-id", func() []string {
		return getAllSubnetIDNames(ctx, *req.VPCId, req.GetProjectId(), req.GetRegion())
	})

	cmd.MarkFlagRequired("charge-type")
	cmd.MarkFlagRequired("edition")
	cmd.MarkFlagRequired("mode")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("storage-gb")
	cmd.MarkFlagRequired("subnet-id")
	cmd.MarkFlagRequired("tps")
	cmd.MarkFlagRequired("vpc-id")

	return cmd
}

// describeServiceByID returns the describe function used by Sspoll, polling via GetURocketMQService
func describeServiceByID(ctx *cli.Context) func(string, *request.CommonBase) (interface{}, error) {
	return func(id string, commonBase *request.CommonBase) (interface{}, error) {
		client := cli.NewServiceClient(ctx, urocketmq.NewClient)
		req := client.NewGetURocketMQServiceRequest()
		if commonBase != nil {
			req.CommonBase = *commonBase
		}
		req.ServiceId = sdk.String(id)
		resp, err := client.GetURocketMQService(req)
		if err != nil {
			return nil, err
		}
		if len(resp.ServiceList) < 1 {
			return nil, nil
		}
		return &resp.ServiceList[0], nil
	}
}
