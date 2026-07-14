package nlb

import (
	"github.com/spf13/cobra"

	nlbsdk "github.com/ucloud/ucloud-sdk-go/services/nlb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/internal/common"
	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newList implements `nlb list`.
func newList(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, nlbsdk.NewClient)
	req := client.NewDescribeNetworkLoadBalancersRequest()

	var nlbID string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List NLB instances",
		Long:  "List NLB instances in the active region/project.",
		Run: func(c *cobra.Command, args []string) {
			req.ProjectId = sdk.String(ctx.PickResourceID(*req.ProjectId))
			if id := ctx.PickResourceID(nlbID); id != "" {
				req.NLBIds = []string{id}
			}
			req.VPCId = sdk.String(ctx.PickResourceID(*req.VPCId))
			req.SubnetId = sdk.String(ctx.PickResourceID(*req.SubnetId))

			resp, err := client.DescribeNetworkLoadBalancers(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			rows := make([]NLBRow, 0, len(resp.NLBs))
			for _, n := range resp.NLBs {
				rows = append(rows, NLBRow{
					ResourceID:       n.NLBId,
					Name:             n.Name,
					Status:           n.Status,
					VPC:              n.VPCId,
					Subnet:           n.SubnetId,
					IPVersion:        n.IPVersion,
					ForwardingMode:   n.ForwardingMode,
					AutoRenewEnabled: n.AutoRenewEnabled,
					PurchaseValue:    common.FormatDate(n.PurchaseValue),
					Group:            n.Tag,
					CreationTime:     common.FormatDate(n.CreateTime),
				})
			}
			ctx.PrintList(rows)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)

	flags.StringVar(&nlbID, resourceIDFlag, "", "Optional. List only the specified NLB instance.")
	req.VPCId = flags.String("vpc-id", "", "Optional. List only NLB instances in the specified VPC.")
	req.SubnetId = flags.String("subnet-id", "", "Optional. List only NLB instances in the specified subnet.")
	req.Offset = flags.Int("offset", 0, "Optional. Offset.")
	req.Limit = flags.Int("limit", 100, "Optional. Limit.")

	command.SetCompletion(cmd, resourceIDFlag, func() []string {
		return getAllNLBIDNames(ctx, derefStr(req.ProjectId), derefStr(req.Region))
	})
	command.SetCompletion(cmd, "vpc-id", func() []string {
		return getAllVPCIDNames(ctx, derefStr(req.ProjectId), derefStr(req.Region))
	})
	command.SetCompletion(cmd, "subnet-id", func() []string {
		return getAllSubnetIDNames(ctx, derefStr(req.VPCId), derefStr(req.ProjectId), derefStr(req.Region))
	})

	return cmd
}
