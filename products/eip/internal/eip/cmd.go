package eip

import (
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/unet"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/model/status"
	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// NewCommand builds the `eip` root command and mounts the 9 subcommands.
// Mirrors cmd/eip.go NewCmdEIP (same AddCommand order).
func NewCommand(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "eip",
		Short: "List,allocate and release EIP",
		Long:  `Manipulate EIP, such as list,allocate and release`,
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(newList(ctx))
	cmd.AddCommand(newAllocate(ctx))
	cmd.AddCommand(newRelease(ctx))
	cmd.AddCommand(newBind(ctx))
	cmd.AddCommand(newUnbind(ctx))
	cmd.AddCommand(newModifyBandwidth(ctx))
	cmd.AddCommand(newSetChargeMode(ctx))
	cmd.AddCommand(newJoinSharedBW(ctx))
	cmd.AddCommand(newLeaveSharedBW(ctx))
	return cmd
}

// newList ucloud eip list
func newList(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, unet.NewClient)
	req := client.NewDescribeEIPRequest()
	fetchAll := false
	pageOff := false
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List all EIP instances",
		Long:    `List all EIP instances`,
		Example: "ucloud eip list",
		Run: func(cmd *cobra.Command, args []string) {
			var eipList []unet.UnetEIPSet
			if fetchAll || pageOff {
				list, err := fetchAllEip(ctx, *req.ProjectId, *req.Region)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				eipList = list
			} else {
				resp, err := client.DescribeEIP(req)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				eipList = resp.EIPSet
			}

			list := make([]EIPRow, 0)
			for _, eip := range eipList {
				row := EIPRow{}
				row.Name = eip.Name
				for _, ip := range eip.EIPAddr {
					row.IP += ip.IP + " " + ip.OperatorName + "   "
				}
				row.ResourceID = eip.EIPId
				row.Group = eip.Tag
				row.ChargeMode = eip.PayMode
				row.Bandwidth = strconv.Itoa(eip.Bandwidth) + "Mb"
				if eip.Resource.ResourceID != "" {
					row.BindResource = fmt.Sprintf("%s|%s(%s)", eip.Resource.ResourceName, eip.Resource.ResourceID, eip.Resource.ResourceType)
				}
				row.Status = eip.Status
				row.ExpirationTime = time.Unix(int64(eip.ExpireTime), 0).Format("2006-01-02")
				list = append(list, row)
			}
			ctx.PrintList(list)
		},
	}

	flags := cmd.Flags()
	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)
	req.Offset = flags.Int("offset", 0, "Optional. Offset default 0")
	req.Limit = flags.Int("limit", 50, "Optional. Limit default 50, max value 100")
	flags.BoolVar(&fetchAll, "list-all", false, "List all eip")
	flags.BoolVar(&pageOff, "page-off", false, "Optional. Paging or not. Accept values: true or false")
	command.SetFlagValues(cmd, "list-all", "true", "false")
	flags.MarkDeprecated("list-all", "please use '--page-off' instead")

	return cmd
}

// newAllocate ucloud eip allocate
func newAllocate(ctx *cli.Context) *cobra.Command {
	var count *int
	client := cli.NewServiceClient(ctx, unet.NewClient)
	req := client.NewAllocateEIPRequest()
	cmd := &cobra.Command{
		Use:     "allocate",
		Short:   "Allocate EIP",
		Long:    "Allocate EIP",
		Example: "ucloud eip allocate --line BGP --bandwidth-mb 2",
		Run: func(cmd *cobra.Command, args []string) {
			if *req.OperatorName == "" {
				*req.OperatorName = getEIPLine(*req.Region)
			}
			results := []cli.OpResultRow{}
			for i := 0; i < *count; i++ {
				resp, err := client.AllocateEIP(req)
				if err != nil {
					ctx.HandleError(err)
					continue
				}
				for _, eip := range resp.EIPSet {
					fmt.Fprintf(ctx.ProgressWriter(), "allocate EIP[%s] ", eip.EIPId)
					for _, ip := range eip.EIPAddr {
						fmt.Fprintf(ctx.ProgressWriter(), "IP:%s  Line:%s \n", ip.IP, ip.OperatorName)
					}
					results = append(results, cli.OpResultRow{ResourceID: eip.EIPId, Action: "allocate", Status: "Allocated"})
				}
			}
			ctx.EmitResult(results...)
		},
	}
	cmd.Flags().SortFlags = false
	req.Bandwidth = cmd.Flags().Int("bandwidth-mb", 0, "Required. Bandwidth(Unit:Mbps).The range of value related to network charge mode. By traffic [1, 200]; by bandwidth [1,800] (Unit: Mbps); it could be 0 if the eip belong to the shared bandwidth")
	req.OperatorName = cmd.Flags().String("line", "", "Optional. 'BGP' or 'International'. 'BGP' could be set in China mainland regions, such as cn-bj2 etc. 'International' could be set in the regions beyond mainland, such as hk, tw-kh, us-ws etc.")
	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)
	req.PayMode = cmd.Flags().String("traffic-mode", "Bandwidth", "Optional. traffic-mode is an enumeration value. 'Traffic','Bandwidth' or 'ShareBandwidth'")
	req.ShareBandwidthId = cmd.Flags().String("share-bandwidth-id", "", "Optional. ShareBandwidthId, required only when traffic-mode is 'ShareBandwidth'")
	req.Quantity = cmd.Flags().Int("quantity", 1, "Optional. The duration of the instance. N years/months.")
	req.ChargeType = cmd.Flags().String("charge-type", "Month", "Optional. Enumeration value.'Year',pay yearly;'Month',pay monthly;'Dynamic', pay hourly(requires permission),'Trial', free trial(need permission)")
	req.Tag = cmd.Flags().String("group", "Default", "Optional. Group of your EIP.")
	req.Name = cmd.Flags().String("name", "EIP", "Optional. Name of your EIP.")
	req.Remark = cmd.Flags().String("remark", "", "Optional. Remark of your EIP.")
	count = cmd.Flags().Int("count", 1, "Optional. Count of EIP to allocate")

	command.SetFlagValues(cmd, "line", "BGP", "International")
	command.SetFlagValues(cmd, "traffic-mode", "Bandwidth", "Traffic", "ShareBandwidth")
	command.SetFlagValues(cmd, "charge-type", "Month", "Year", "Dynamic", "Trial")
	cmd.MarkFlagRequired("bandwidth-mb")
	return cmd
}

// newRelease ucloud eip release
func newRelease(ctx *cli.Context) *cobra.Command {
	var ids []string
	client := cli.NewServiceClient(ctx, unet.NewClient)
	req := client.NewReleaseEIPRequest()
	cmd := &cobra.Command{
		Use:     "release",
		Short:   "Release EIP",
		Long:    "Release EIP",
		Example: "ucloud eip release --eip-id eip-xx1,eip-xx2",
		Run: func(cmd *cobra.Command, args []string) {
			req.ProjectId = sdk.String(ctx.PickResourceID(*req.ProjectId))
			results := []cli.OpResultRow{}
			for _, id := range ids {
				req.EIPId = sdk.String(ctx.PickResourceID(id))
				_, err := client.ReleaseEIP(req)
				if err != nil {
					ctx.HandleError(err)
				} else {
					fmt.Fprintf(ctx.ProgressWriter(), "eip[%s] released\n", *req.EIPId)
					results = append(results, cli.OpResultRow{ResourceID: *req.EIPId, Action: "release", Status: "Released"})
				}
			}
			ctx.EmitResult(results...)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	flags.StringSliceVarP(&ids, "eip-id", "", nil, "Required. Resource ID of the EIPs you want to release")
	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)
	cmd.MarkFlagRequired("eip-id")
	command.SetCompletion(cmd, "eip-id", func() []string {
		return getAllEip(ctx, *req.ProjectId, *req.Region, []string{status.EIP_FREE}, nil)
	})

	return cmd
}

// newBind ucloud eip bind
func newBind(ctx *cli.Context) *cobra.Command {
	var projectID, region, resourceID, resourceType *string
	var eipIDs []string
	cmd := &cobra.Command{
		Use:     "bind",
		Short:   "Bind EIP with uhost",
		Long:    "Bind EIP with uhost",
		Example: "ucloud eip bind --eip-id eip-xxx --resource-id uhost-xxx",
		Run: func(cmd *cobra.Command, args []string) {
			results := []cli.OpResultRow{}
			for _, eipID := range eipIDs {
				if err := bindEIP(ctx, resourceID, resourceType, &eipID, projectID, region); err == nil {
					results = append(results, cli.OpResultRow{ResourceID: ctx.PickResourceID(eipID), Action: "bind", Status: "Bound"})
				}
			}
			ctx.EmitResult(results...)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	cmd.Flags().StringSliceVar(&eipIDs, "eip-id", nil, "Required. EIPId to bind")
	resourceID = cmd.Flags().String("resource-id", "", "Required. ResourceID , which is the UHostId of uhost")
	resourceType = cmd.Flags().String("resource-type", "uhost", "Requried. ResourceType, type of resource to bind with eip. 'uhost','vrouter','ulb','upm','hadoophost'.eg..")
	projectID = cmd.Flags().String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	region = cmd.Flags().String("region", ctx.DefaultRegion(), "Optional. Assign region")

	command.SetFlagValues(cmd, "resource-type", "uhost", "vrouter", "ulb", "upm", "hadoophost", "fortresshost", "udockhost", "udhost", "natgw", "udb", "vpngw", "ucdr", "dbaudit")
	command.SetCompletion(cmd, "eip-id", func() []string {
		return getAllEip(ctx, *projectID, *region, []string{status.EIP_FREE}, nil)
	})

	cmd.MarkFlagRequired("eip-id")
	cmd.MarkFlagRequired("resource-id")

	return cmd
}

// newUnbind ucloud eip unbind
func newUnbind(ctx *cli.Context) *cobra.Command {
	eipIDs := []string{}
	client := cli.NewServiceClient(ctx, unet.NewClient)
	req := client.NewUnBindEIPRequest()
	cmd := &cobra.Command{
		Use:     "unbind",
		Short:   "Unbind EIP with uhost",
		Long:    "Unbind EIP with uhost",
		Example: "ucloud eip unbind --eip-id eip-xxx",
		Run: func(cmd *cobra.Command, args []string) {
			req.ProjectId = sdk.String(ctx.PickResourceID(*req.ProjectId))
			results := []cli.OpResultRow{}
			for _, eip := range eipIDs {
				eipIns, err := getEIP(ctx, ctx.PickResourceID(eip))
				if err != nil {
					ctx.HandleError(err)
					return
				}
				req.EIPId = sdk.String(ctx.PickResourceID(eip))
				req.ResourceId = sdk.String(eipIns.Resource.ResourceID)
				req.ResourceType = sdk.String(eipIns.Resource.ResourceType)
				_, err = client.UnBindEIP(req)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				fmt.Fprintf(ctx.ProgressWriter(), "unbind EIP[%s] with %s[%s]\n", *req.EIPId, *req.ResourceType, *req.ResourceId)
				results = append(results, cli.OpResultRow{ResourceID: *req.EIPId, Action: "unbind", Status: "Unbound"})
			}
			ctx.EmitResult(results...)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&eipIDs, "eip-id", nil, "Required. Resource ID of eips to unbind with some resource")
	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)

	cmd.MarkFlagRequired("eip-id")
	command.SetCompletion(cmd, "eip-id", func() []string {
		return getAllEip(ctx, *req.ProjectId, *req.Region, []string{status.EIP_USED}, nil)
	})

	return cmd
}

// newModifyBandwidth ucloud eip modify-bw
func newModifyBandwidth(ctx *cli.Context) *cobra.Command {
	ids := []string{}
	client := cli.NewServiceClient(ctx, unet.NewClient)
	req := client.NewModifyEIPBandwidthRequest()
	cmd := &cobra.Command{
		Use:     "modify-bw",
		Short:   "Modify bandwith of EIP instances",
		Long:    "Modify bandwith of EIP instances",
		Example: "ucloud eip modify-bw --eip-id eip-xx1,eip-xx2 --bandwidth-mb 20",
		// Deprecated: "use 'ucloud eip modiy'",
		Run: func(cmd *cobra.Command, args []string) {
			results := []cli.OpResultRow{}
			for _, id := range ids {
				id = ctx.PickResourceID(id)
				req.EIPId = &id
				_, err := client.ModifyEIPBandwidth(req)
				if err != nil {
					ctx.HandleError(err)
				} else {
					fmt.Fprintf(ctx.ProgressWriter(), "eip[%s]'s bandwidth modified\n", id)
					results = append(results, cli.OpResultRow{ResourceID: id, Action: "modify-bw", Status: "Modified"})
				}
			}
			ctx.EmitResult(results...)
		},
	}
	cmd.Flags().SortFlags = false
	cmd.Flags().StringSliceVarP(&ids, "eip-id", "", nil, "Required, Resource ID of EIPs to modify bandwidth")
	req.Bandwidth = cmd.Flags().Int("bandwidth-mb", 0, "Required. Bandwidth of EIP after modifed. Charge by traffic, range [1,300]; charge by bandwidth, range [1,800]")
	req.ProjectId = cmd.Flags().String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = cmd.Flags().String("region", ctx.DefaultRegion(), "Optional. Assign region")
	command.SetCompletion(cmd, "eip-id", func() []string {
		return getAllEip(ctx, *req.ProjectId, *req.Region, nil, nil)
	})
	cmd.MarkFlagRequired("eip-id")
	cmd.MarkFlagRequired("bandwidth-mb")
	return cmd
}

// newSetChargeMode ucloud eip modify-traffic-mode
func newSetChargeMode(ctx *cli.Context) *cobra.Command {
	ids := []string{}
	client := cli.NewServiceClient(ctx, unet.NewClient)
	req := client.NewSetEIPPayModeRequest()
	cmd := &cobra.Command{
		Use:     "modify-traffic-mode",
		Short:   "Modify charge mode of EIP instances",
		Long:    "Modify charge mode of EIP instances",
		Example: "ucloud eip modify-traffic-mode --eip-id eip-xx1,eip-xx2 --traffic-mode Traffic",
		Run: func(cmd *cobra.Command, args []string) {
			results := []cli.OpResultRow{}
			for _, id := range ids {
				id = ctx.PickResourceID(id)
				req.EIPId = &id
				eipIns, err := getEIP(ctx, id)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				req.Bandwidth = sdk.Int(eipIns.Bandwidth)
				_, err = client.SetEIPPayMode(req)
				if err != nil {
					ctx.HandleError(err)
				} else {
					fmt.Fprintf(ctx.ProgressWriter(), "eip[%s]'s charge mode was modified to %s\n", id, *req.PayMode)
					results = append(results, cli.OpResultRow{ResourceID: id, Action: "modify-traffic-mode", Status: "Modified"})
				}
			}
			ctx.EmitResult(results...)
		},
	}

	cmd.Flags().SortFlags = false
	cmd.Flags().StringSliceVarP(&ids, "eip-id", "", nil, "Required, Resource ID of EIPs to modify charge mode")
	req.PayMode = cmd.Flags().String("traffic-mode", "", "Required, Charge mode of eip, 'Traffic','Bandwidth' or 'PostAccurateBandwidth'")
	req.ProjectId = cmd.Flags().String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = cmd.Flags().String("region", ctx.DefaultRegion(), "Optional. Assign region")
	command.SetFlagValues(cmd, "traffic-mode", "Bandwidth", "Traffic", "PostAccurateBandwidth")
	command.SetCompletion(cmd, "eip-id", func() []string {
		return getAllEip(ctx, *req.ProjectId, *req.Region, nil, nil)
	})
	cmd.MarkFlagRequired("eip-id")
	cmd.MarkFlagRequired("traffic-mode")
	return cmd
}

// newJoinSharedBW ucloud eip join-shared-bw
func newJoinSharedBW(ctx *cli.Context) *cobra.Command {
	eipIDs := []string{}
	client := cli.NewServiceClient(ctx, unet.NewClient)
	req := client.NewAssociateEIPWithShareBandwidthRequest()
	cmd := &cobra.Command{
		Use:     "join-shared-bw",
		Short:   "Join shared bandwidth",
		Long:    "Join shared bandwidth",
		Example: "ucloud eip join-shared-bw --eip-id eip-xxx --shared-bw-id bwshare-xxx",
		Run: func(c *cobra.Command, args []string) {
			for _, eip := range eipIDs {
				req.EIPIds = append(req.EIPIds, ctx.PickResourceID(eip))
			}
			req.ShareBandwidthId = sdk.String(ctx.PickResourceID(*req.ShareBandwidthId))
			_, err := client.AssociateEIPWithShareBandwidth(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "eip%v joined shared bandwidth[%s]\n", req.EIPIds, *req.ShareBandwidthId)
			results := []cli.OpResultRow{}
			for _, eipID := range req.EIPIds {
				results = append(results, cli.OpResultRow{ResourceID: eipID, Action: "join-shared-bw", Status: "Joined"})
			}
			ctx.EmitResult(results...)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	flags.StringSliceVar(&eipIDs, "eip-id", nil, "Required. Resource ID of EIPs to join shared bandwdith")
	req.ShareBandwidthId = flags.String("shared-bw-id", "", "Required. Resource ID of shared bandwidth to be joined")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Region, see 'ucloud region'")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Project-id, see 'ucloud project list'")
	command.SetCompletion(cmd, "eip-id", func() []string {
		return getAllEip(ctx, *req.ProjectId, *req.Region, nil, []string{status.EIP_CHARGE_BANDWIDTH, status.EIP_CHARGE_TRAFFIC})
	})
	command.SetCompletion(cmd, "shared-bw-id", func() []string {
		list, _ := getAllSharedBW(ctx, *req.ProjectId, *req.Region)
		return list
	})
	cmd.MarkFlagRequired("eip-id")
	cmd.MarkFlagRequired("shared-bw-id")

	return cmd
}

// newLeaveSharedBW ucloud eip leave-shared-bw
func newLeaveSharedBW(ctx *cli.Context) *cobra.Command {
	eipIDs := []string{}
	client := cli.NewServiceClient(ctx, unet.NewClient)
	req := client.NewDisassociateEIPWithShareBandwidthRequest()
	cmd := &cobra.Command{
		Use:     "leave-shared-bw",
		Short:   "Leave shared bandwidth",
		Long:    "Leave shared bandwidth",
		Example: "ucloud eip leave-shared-bw --eip-id eip-b2gvu3",
		Run: func(c *cobra.Command, args []string) {
			results := []cli.OpResultRow{}
			if *req.ShareBandwidthId == "" {
				for _, eipID := range eipIDs {
					eipIns, err := getEIP(ctx, ctx.PickResourceID(eipID))
					if err != nil {
						ctx.HandleError(err)
						continue
					}
					sharedBWID := eipIns.ShareBandwidthSet.ShareBandwidthId
					if sharedBWID == "" {
						fmt.Fprintf(ctx.ProgressWriter(), "eip[%s] doesn't join any shared bandwidth\n", eipID)
						continue
					}
					req.ShareBandwidthId = sdk.String(sharedBWID)
					req.EIPIds = []string{ctx.PickResourceID(eipID)}
					_, err = client.DisassociateEIPWithShareBandwidth(req)
					if err != nil {
						ctx.HandleError(err)
						continue
					}
					fmt.Fprintf(ctx.ProgressWriter(), "eip[%s] left shared bandwidth[%s]\n", eipID, sharedBWID)
					results = append(results, cli.OpResultRow{ResourceID: ctx.PickResourceID(eipID), Action: "leave-shared-bw", Status: "Left"})
				}
			} else {
				for _, id := range eipIDs {
					req.EIPIds = append(req.EIPIds, ctx.PickResourceID(id))
				}
				*req.ShareBandwidthId = ctx.PickResourceID(*req.ShareBandwidthId)
				_, err := client.DisassociateEIPWithShareBandwidth(req)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				fmt.Fprintf(ctx.ProgressWriter(), "eip%v left shared bandwidth[%s]\n", eipIDs, *req.ShareBandwidthId)
				for _, eipID := range req.EIPIds {
					results = append(results, cli.OpResultRow{ResourceID: eipID, Action: "leave-shared-bw", Status: "Left"})
				}
			}
			ctx.EmitResult(results...)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	flags.StringSliceVar(&eipIDs, "eip-id", nil, "Required. Resource ID of EIPs to leave shared bandwidth")
	req.Bandwidth = flags.Int("bandwidth-mb", 1, "Required. Bandwidth of EIP after leaving shared bandwidth, ranging [1,300] for 'Traffic' charge mode, ranging [1,800] for 'Bandwidth' charge mode. Unit:Mb")
	req.PayMode = flags.String("traffic-mode", "Bandwidth", "Optional. Charge mode of the EIP after leaving shared bandwidth, 'Bandwidth' or 'Traffic'")
	req.ShareBandwidthId = flags.String("shared-bw-id", "", "Optional. Resource ID of shared bandwidth instance, assign this flag to make the operation faster")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Region, see 'ucloud region'")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Project-id, see 'ucloud project list'")

	command.SetFlagValues(cmd, "traffic-mode", "Bandwidth", "Traffic")
	command.SetCompletion(cmd, "eip-id", func() []string {
		return getAllEip(ctx, *req.ProjectId, *req.Region, nil, []string{status.EIP_CHARGE_SHARE})
	})
	command.SetCompletion(cmd, "shared-bw-id", func() []string {
		list, _ := getAllSharedBW(ctx, *req.ProjectId, *req.Region)
		return list
	})

	// L2 prebug preserved verbatim: the flag is named "bandwidth-mb" (above), so
	// MarkFlagRequired("bandwidth") is a silent no-op. Matches cmd/eip.go ~:649.
	cmd.MarkFlagRequired("bandwidth")
	cmd.MarkFlagRequired("eip-id")
	return cmd
}
