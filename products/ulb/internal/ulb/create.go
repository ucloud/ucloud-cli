package ulb

import (
	"fmt"

	"github.com/spf13/cobra"

	ulbsdk "github.com/ucloud/ucloud-sdk-go/services/ulb"
	"github.com/ucloud/ucloud-sdk-go/services/unet"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newCreate returns ucloud ulb create.
func newCreate(ctx *cli.Context) *cobra.Command {
	var bindEipID *string
	mode := "outer"
	client := cli.NewServiceClient(ctx, ulbsdk.NewClient)
	unetClient := cli.NewServiceClient(ctx, unet.NewClient)
	req := client.NewCreateULBRequest()
	eipReq := unetClient.NewAllocateEIPRequest()
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create ULB instance",
		Long:  "Create ULB instance",
		Run: func(c *cobra.Command, args []string) {
			req.ProjectId = sdk.String(ctx.PickResourceID(*req.ProjectId))
			if mode == "outer" {
				if *bindEipID == "" && *eipReq.Bandwidth == 0 {
					fmt.Fprintln(ctx.ProgressWriter(), "Outer mode ULB need a eip to bind, please assign eip by flag 'bind-eip' or create eip by 'create-eip-bandwidth-mb'")
					return
				}
				if *eipReq.OperatorName == "" {
					*eipReq.OperatorName = getEIPLine(*req.Region)
				}
				req.OuterMode = sdk.String("Yes")
			} else if mode == "inner" {
				req.InnerMode = sdk.String("Yes")
			} else {
				fmt.Fprintln(ctx.ProgressWriter(), "Error, flag mode should be 'outer' or 'inner'")
				return
			}
			req.VPCId = sdk.String(ctx.PickResourceID(*req.VPCId))
			req.SubnetId = sdk.String(ctx.PickResourceID(*req.SubnetId))
			resp, err := client.CreateULB(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "ulb[%s] created\n", resp.ULBId)
			ctx.EmitResult(cli.OpResultRow{ResourceID: resp.ULBId, Action: "create", Status: "Created"})
			if mode == "inner" {
				return
			}
			bindEipID = sdk.String(ctx.PickResourceID(*bindEipID))
			if *bindEipID != "" {
				_ = bindEIP(ctx, sdk.String(resp.ULBId), sdk.String("ulb"), bindEipID, req.ProjectId, req.Region)
				return
			}
			if *eipReq.OperatorName != "" && *eipReq.Bandwidth != 0 {
				eipReq.ChargeType = req.ChargeType
				eipReq.Tag = req.Tag
				eipReq.Region = req.Region
				eipReq.ProjectId = req.ProjectId
				eipResp, err := unetClient.AllocateEIP(eipReq)

				if err != nil {
					ctx.HandleError(err)
					return
				}

				for _, eip := range eipResp.EIPSet {
					fmt.Fprintf(ctx.ProgressWriter(), "allocate EIP[%s] ", eip.EIPId)
					for _, ip := range eip.EIPAddr {
						fmt.Fprintf(ctx.ProgressWriter(), "IP:%s  Line:%s \n", ip.IP, ip.OperatorName)
					}
					_ = bindEIP(ctx, sdk.String(resp.ULBId), sdk.String("ulb"), sdk.String(eip.EIPId), req.ProjectId, req.Region)
				}
			}
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.ULBName = flags.String("name", "", "Required. Name of ULB instance to create")
	flags.StringVar(&mode, "mode", "outer", "Required. Network mode of ULB instance, outer or inner.")
	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)
	req.VPCId = flags.String("vpc-id", "", "Optional. Resource ID of VPC which the ULB to create belong to. See 'ucloud vpc list'")
	req.SubnetId = flags.String("subnet-id", "", "Optional. Resource ID of subnet. This flag will be discarded when you are creating an outter mode ULB. See 'ucloud subnet list'")
	req.ChargeType = flags.String("charge-type", "Month", "Optional.'Year',pay yearly;'Month',pay monthly;'Dynamic', pay hourly")
	req.Tag = flags.String("group", "Default", "Optional. Business group")
	req.Remark = flags.String("remark", "", "Optional. Remark of instance to create.")
	bindEipID = flags.String("bind-eip", "", "Optional. Resource ID or IP Address of eip that will be bound to the new created outer mode ulb")
	eipReq.Bandwidth = cmd.Flags().Int("create-eip-bandwidth-mb", 0, "Optional. Required if you want to create new EIP. Bandwidth(Unit:Mbps).The range of value related to network charge mode. By traffic [1, 300]; by bandwidth [1,800] (Unit: Mbps); it could be 0 if the eip belong to the shared bandwidth")
	eipReq.OperatorName = flags.String("create-eip-line", "", "Optional. Line of created eip to bind with the new created outer mode ulb")
	eipReq.PayMode = cmd.Flags().String("create-eip-traffic-mode", "Bandwidth", "Optional. 'Traffic','Bandwidth' or 'ShareBandwidth'")
	eipReq.Name = flags.String("create-eip-name", "", "Optional. Name of created eip to bind with the new created outer mode ulb")
	eipReq.Remark = cmd.Flags().String("create-eip-remark", "", "Optional. Remark of your EIP.")

	command.SetFlagValues(cmd, "mode", "outer", "inner")
	command.SetFlagValues(cmd, "charge-type", "Month", "Year", "Dynamic")
	command.SetFlagValues(cmd, "create-eip-line", "BGP", "International")
	command.SetFlagValues(cmd, "create-eip-traffic-mode", "Bandwidth", "Traffic", "ShareBandwidth")
	command.SetCompletion(cmd, "bind-eip", func() []string {
		return getAllEip(ctx, *req.ProjectId, *req.Region, []string{EIP_FREE}, nil)
	})
	command.SetCompletion(cmd, "vpc-id", func() []string {
		return getAllVPCIdNames(ctx, *req.ProjectId, *req.Region)
	})
	command.SetCompletion(cmd, "subnet-id", func() []string {
		return getAllSubnetIDNames(ctx, *req.VPCId, *req.ProjectId, *req.Region)
	})

	cmd.MarkFlagRequired("mode")
	cmd.MarkFlagRequired("name")

	return cmd
}
