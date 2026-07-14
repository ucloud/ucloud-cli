package ext

import (
	"fmt"
	"net"
	"strings"

	"github.com/spf13/cobra"
	uhostsdk "github.com/ucloud/ucloud-sdk-go/services/uhost"
	"github.com/ucloud/ucloud-sdk-go/services/unet"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newUHostSwitchEIP builds `ucloud ext uhost switch-eip`.
func newUHostSwitchEIP(ctx *cli.Context) *cobra.Command {
	var eipAddrs []string
	var eipBandwidth, quantity int
	var chargeType, trafficMode, shareBandwidthID string
	var uhostIDs []string
	var unbind, release bool

	uhostClient := cli.NewServiceClient(ctx, uhostsdk.NewClient)
	describeReq := uhostClient.NewDescribeUHostInstanceRequest()

	cmd := &cobra.Command{
		Use:     "switch-eip",
		Short:   "Switch EIP for UHost instances",
		Long:    "Switch EIP for UHost instances",
		Example: "ucloud ext uhost switch-eip --uhost-id uhost-1n1sxx2,uhost-li4jxx1 --create-eip-bandwidth-mb 2",
		Run: func(c *cobra.Command, args []string) {
			project := ctx.PickResourceID(*describeReq.ProjectId)
			region := *describeReq.Region
			zone := *describeReq.Zone
			unetClient := cli.NewServiceClient(ctx, unet.NewClient)
			eipAddrMap := make(map[string]bool)
			for _, addr := range eipAddrs {
				eipAddrMap[addr] = true
			}
			results := []cli.OpResultRow{}

			for _, idName := range uhostIDs {
				uhostID := ctx.PickResourceID(idName)
				logs := []string{fmt.Sprintf("describe uhost instance by uhostID %s", uhostID)}
				uhostIns, err := describeUHostByID(ctx, uhostID, project, region, zone)
				if err != nil {
					errStr := fmt.Sprintf("describe uhost %s failed: %v", uhostID, err)
					ctx.HandleError(fmt.Errorf("%s", errStr))
					ctx.LogInfo(append(logs, errStr)...)
					continue
				}

				for _, ip := range uhostIns.IPSet {
					if ip.IPId == "" {
						continue
					}
					if len(eipAddrs) > 0 && !eipAddrMap[ip.IP] {
						continue
					}

					req := unetClient.NewAllocateEIPRequest()
					req.Region = &region
					req.ProjectId = &project
					req.OperatorName = sdk.String(defaultEIPLine(region))
					req.Bandwidth = &eipBandwidth
					req.ChargeType = &chargeType
					req.Quantity = &quantity
					req.PayMode = &trafficMode
					if trafficMode == "ShareBandwidth" {
						if shareBandwidthID == "" {
							errStr := "create-eip-share-bandwidth-id should not be empty when create-eip-traffic-mode is assigned 'ShareBandwidth'"
							ctx.HandleError(fmt.Errorf("%s", errStr))
							ctx.LogInfo(append(logs, errStr)...)
							return
						}
						req.ShareBandwidthId = &shareBandwidthID
					}

					resp, err := unetClient.AllocateEIP(req)
					if err != nil {
						errStr := fmt.Sprintf("allocate EIP failed: %v", err)
						ctx.HandleError(fmt.Errorf("%s", errStr))
						ctx.LogInfo(append(logs, errStr)...)
						continue
					}
					if len(resp.EIPSet) != 1 {
						errStr := "allocate EIP failed, length of eip set is not 1"
						ctx.HandleError(fmt.Errorf("%s", errStr))
						ctx.LogInfo(append(logs, errStr)...)
						continue
					}

					eipID := resp.EIPSet[0].EIPId
					eipIP := ""
					if len(resp.EIPSet[0].EIPAddr) > 0 {
						eipIP = resp.EIPSet[0].EIPAddr[0].IP
					}
					allocRet := fmt.Sprintf("allocated new eip %s|%s", eipID, eipIP)
					logs = append(logs, allocRet)
					fmt.Fprintln(ctx.ProgressWriter(), allocRet)
					results = append(results, cli.OpResultRow{ResourceID: eipID, Action: "allocate", Status: "Allocated"})

					bindLogs, bindErr := bindEIPWithLogs(ctx, &uhostID, sdk.String("uhost"), &eipID, &project, &region)
					logs = append(logs, bindLogs...)
					if bindErr != nil {
						ctx.HandleError(fmt.Errorf("bind new eip %s failed: %v", eipID, bindErr))
						ctx.LogInfo(logs...)
						continue
					}
					fmt.Fprintf(ctx.ProgressWriter(), "bound eip %s with uhost %s\n", eipID, uhostID)
					results = append(results, cli.OpResultRow{ResourceID: eipID, Action: "bind", Status: "Bound"})

					if unbind {
						unbindLogs, err := unbindEIPWithLogs(ctx, uhostID, "uhost", ip.IPId, project, region)
						logs = append(logs, unbindLogs...)
						if err != nil {
							ctx.HandleError(fmt.Errorf("unbind eip %s failed: %v", ip.IPId, err))
							ctx.LogInfo(logs...)
							continue
						}
						fmt.Fprintf(ctx.ProgressWriter(), "unbound eip %s|%s with uhost %s\n", ip.IPId, ip.IP, uhostID)
						results = append(results, cli.OpResultRow{ResourceID: ip.IPId, Action: "unbind", Status: "Unbound"})
					}

					if release {
						req := unetClient.NewReleaseEIPRequest()
						req.ProjectId = &project
						req.Region = &region
						req.EIPId = sdk.String(ip.IPId)
						_, err := unetClient.ReleaseEIP(req)
						if err != nil {
							errStr := fmt.Sprintf("release eip %s failed: %v", ip.IPId, err)
							ctx.HandleError(fmt.Errorf("%s", errStr))
							ctx.LogInfo(append(logs, errStr)...)
							continue
						}
						releaseRet := fmt.Sprintf("released eip %s|%s", ip.IPId, ip.IP)
						logs = append(logs, releaseRet)
						fmt.Fprintln(ctx.ProgressWriter(), releaseRet)
						results = append(results, cli.OpResultRow{ResourceID: ip.IPId, Action: "release", Status: "Released"})
					}
					ctx.LogInfo(logs...)
				}
			}
			ctx.EmitResult(results...)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	flags.StringSliceVar(&uhostIDs, "uhost-id", nil, "Required. Resource ID of uhost instances to switch EIP")
	flags.StringSliceVar(&eipAddrs, "eip-addr", nil, "Optional. Address of EIP instances to be replaced. if eip-id is empty, replace all of the EIPs bound with the uhost ")
	flags.BoolVar(&unbind, "unbind-all", true, "Optional. Unbind all EIP instances that has been replaced. Accept values:true or false")
	flags.BoolVar(&release, "release-all", true, "Optional. Release all EIP instances that has been replaced. Accept values:true or false")
	flags.IntVar(&eipBandwidth, "create-eip-bandwidth-mb", 1, "Optional. Bandwidth of EIP instance to be create with. Unit:Mb")
	flags.StringVar(&trafficMode, "create-eip-traffic-mode", "Bandwidth", "Optional. traffic-mode is an enumeration value. 'Traffic','Bandwidth' or 'ShareBandwidth'")
	flags.StringVar(&shareBandwidthID, "create-eip-share-bandwidth-id", "", "Optional. ShareBandwidthId, required only when traffic-mode is 'ShareBandwidth'")
	flags.StringVar(&chargeType, "create-eip-charge-type", "Month", "Optional. Enumeration value.'Year',pay yearly;'Month',pay monthly;'Dynamic', pay hourly")
	flags.IntVar(&quantity, "create-eip-quantity", 1, "Optional. The duration of the instance. N years/months.")

	command.SetFlagValues(cmd, "create-eip-traffic-mode", "Bandwidth", "Traffic", "ShareBandwidth")
	command.SetFlagValues(cmd, "create-eip-charge-type", "Month", "Year", "Dynamic", "Trial")
	ctx.BindProjectID(cmd, describeReq)
	ctx.BindRegion(cmd, describeReq)
	ctx.BindZoneEmpty(cmd, describeReq)
	command.SetCompletion(cmd, "uhost-id", func() []string {
		return listUHostIDs(ctx, []string{hostRunning, hostStopped, hostFail}, *describeReq.ProjectId, *describeReq.Region, *describeReq.Zone)
	})
	cmd.MarkFlagRequired("uhost-id")

	return cmd
}

func defaultEIPLine(region string) string {
	if strings.HasPrefix(region, "cn") {
		return "BGP"
	}
	return "International"
}

func getEIPIDByIP(ctx *cli.Context, ip net.IP, projectID, region string) (string, error) {
	eipList, err := fetchAllEIP(ctx, projectID, region)
	if err != nil {
		return "", err
	}
	for _, eip := range eipList {
		for _, addr := range eip.EIPAddr {
			if addr.IP == ip.String() {
				return eip.EIPId, nil
			}
		}
	}
	return "", fmt.Errorf("IP[%s] not exist", ip.String())
}

func fetchAllEIP(ctx *cli.Context, projectID, region string) ([]unet.UnetEIPSet, error) {
	client := cli.NewServiceClient(ctx, unet.NewClient)
	req := client.NewDescribeEIPRequest()
	list := []unet.UnetEIPSet{}
	req.ProjectId = sdk.String(projectID)
	req.Region = sdk.String(region)
	for offset, step := 0, 100; ; offset += step {
		req.Offset = &offset
		req.Limit = &step
		resp, err := client.DescribeEIP(req)
		if err != nil {
			return nil, err
		}
		for i, size := 0, len(resp.EIPSet); i < size; i++ {
			list = append(list, resp.EIPSet[i])
		}
		if resp.TotalCount <= offset+step {
			break
		}
	}
	return list, nil
}

func bindEIPWithLogs(ctx *cli.Context, resourceID, resourceType, eipID, projectID, region *string) ([]string, error) {
	logs := make([]string, 0)
	ip := net.ParseIP(*eipID)
	if ip != nil {
		id, err := getEIPIDByIP(ctx, ip, *projectID, *region)
		if err != nil {
			ctx.HandleError(err)
		} else {
			*eipID = id
		}
	}
	client := cli.NewServiceClient(ctx, unet.NewClient)
	req := client.NewBindEIPRequest()
	req.ResourceId = resourceID
	req.ResourceType = resourceType
	req.EIPId = sdk.String(ctx.PickResourceID(*eipID))
	req.ProjectId = sdk.String(ctx.PickResourceID(*projectID))
	req.Region = region
	_, err := client.BindEIP(req)
	if err != nil {
		logs = append(logs, fmt.Sprintf("bind eip failed: %v", err))
		return logs, err
	}
	logs = append(logs, fmt.Sprintf("bind eip[%s] with %s[%s] successfully", *req.EIPId, *req.ResourceType, *req.ResourceId))
	return logs, nil
}

func unbindEIPWithLogs(ctx *cli.Context, resourceID, resourceType, eipID, projectID, region string) ([]string, error) {
	logs := make([]string, 0)
	eipID = ctx.PickResourceID(eipID)
	ip := net.ParseIP(eipID)
	if ip != nil {
		id, err := getEIPIDByIP(ctx, ip, projectID, region)
		if err != nil {
			ctx.HandleError(err)
		} else {
			eipID = id
		}
	}
	client := cli.NewServiceClient(ctx, unet.NewClient)
	req := client.NewUnBindEIPRequest()
	req.ResourceId = &resourceID
	req.ResourceType = &resourceType
	req.EIPId = &eipID
	req.ProjectId = sdk.String(ctx.PickResourceID(projectID))
	req.Region = &region
	_, err := client.UnBindEIP(req)
	if err != nil {
		logs = append(logs, fmt.Sprintf("unbind eip failed: %v", err))
		return logs, err
	}
	logs = append(logs, fmt.Sprintf("unbind eip[%s] with %s[%s] successfully", *req.EIPId, *req.ResourceType, *req.ResourceId))
	return logs, nil
}
