// Copyright © 2018 NAME HERE tony.li@ucloud.cn
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/base"
	"github.com/ucloud/ucloud-cli/model/status"
	"github.com/ucloud/ucloud-sdk-go/services/uhost"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"
)

//NewCmdExt ucloud ext
func NewCmdExt() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ext",
		Short: "extended commands of UCloud CLI",
		Long:  "extended commands of UCloud CLI",
	}
	cmd.AddCommand(NewCmdExtUHost())
	return cmd
}

//NewCmdExtUHost ucloud ext uhost
func NewCmdExtUHost() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uhost",
		Short: "extended uhost commands",
		Long:  "extended uhost commands",
	}
	cmd.AddCommand(NewCmdExtUHostSwitchEIP())
	return cmd
}

//NewCmdExtUHostSwitchEIP ucloud ext uhost switch-eip
func NewCmdExtUHostSwitchEIP() *cobra.Command {
	var project, region, zone, chargeType, trafficMode, shareBandwidthID string
	var uhostIDs, eipAddrs []string
	var eipBandwidth, quntity int
	var unbind, release bool

	cmd := &cobra.Command{
		Use:   "switch-eip",
		Short: "Switch EIP for UHost instances",
		Long:  "Switch EIP for UHost instances",
		Run: func(c *cobra.Command, args []string) {
			project = base.PickResourceID(project)
			eipAddrMap := make(map[string]bool)
			for _, addr := range eipAddrs {
				eipAddrMap[addr] = true
			}
			for _, idname := range uhostIDs {
				uhostID := base.PickResourceID(idname)
				ins, err := describeUHostByID(uhostID, project, region, zone)
				if err != nil {
					base.LogError(fmt.Sprintf("get uhost %s failed: %v", uhostID, err))
					continue
				}
				uhostIns, ok := ins.(*uhost.UHostInstanceSet)
				if !ok {
					base.HandleError(fmt.Errorf("get uhost %s failed", uhostID))
					continue
				}
				for _, ip := range uhostIns.IPSet {
					if ip.IPId == "" {
						continue
					}
					if len(eipAddrs) > 0 && eipAddrMap[ip.IP] == false {
						continue
					}
					//申请EIP
					req := base.BizClient.NewAllocateEIPRequest()
					req.Region = &region
					req.ProjectId = &project
					if strings.HasPrefix(region, "cn") {
						req.OperatorName = sdk.String("BGP")
					} else {
						req.OperatorName = sdk.String("International")
					}
					req.Bandwidth = &eipBandwidth
					req.ChargeType = &chargeType
					req.Quantity = &quntity
					req.PayMode = &trafficMode
					if trafficMode == "ShareBandwidth" {
						if shareBandwidthID != "" {
							req.ShareBandwidthId = &shareBandwidthID
						} else {
							base.HandleError(fmt.Errorf("create-eip-share-bandwidth-id should not be empty when create-eip-traffic-mode is assigned 'ShareBandwidth'"))
							return
						}
					}
					resp, err := base.BizClient.AllocateEIP(req)
					if err != nil {
						base.HandleError(err)
						continue
					}
					if len(resp.EIPSet) != 1 {
						base.HandleError(fmt.Errorf("allocate EIP failed, length of eip set is not 1"))
						continue
					}
					eipID := resp.EIPSet[0].EIPId
					fmt.Printf("allocated new eip %s|%s\n", eipID, resp.EIPSet[0].EIPAddr[0].IP)

					//绑定新EIP
					logs, err2 := sbindEIP(&uhostID, sdk.String("uhost"), &eipID, &project, &region)
					if err2 != nil {
						base.HandleError(fmt.Errorf("bind new eip %s failed: %v", eipID, err2))
						continue
					}
					fmt.Printf("bound eip %s witch uhost %s\n", eipID, uhostID)
					base.LogInfo(logs...)

					if unbind {
						logs, err := unbindEIP(uhostID, "uhost", ip.IPId, project, region)
						base.LogInfo(logs...)
						if err != nil {
							base.HandleError(fmt.Errorf("unbind eip %s failed: %v", ip.IPId, err))
							continue
						}
						fmt.Printf("unbound eip %s|%s witch uhost %s\n", ip.IPId, ip.IP, uhostID)
					}

					if release {
						req := base.BizClient.NewReleaseEIPRequest()
						req.ProjectId = &project
						req.Region = &region
						req.EIPId = sdk.String(ip.IPId)
						_, err := base.BizClient.ReleaseEIP(req)
						if err != nil {
							base.HandleError(fmt.Errorf("release eip %s failed: %v", ip.IPId, err))
							continue
						}
						fmt.Printf("released eip %s|%s \n", ip.IPId, ip.IP)
					}
				}
			}
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
	flags.IntVar(&quntity, "create-eip-quantity", 1, "Optional. The duration of the instance. N years/months.")

	flags.SetFlagValues("create-eip-traffic-mode", "Bandwidth", "Traffic", "ShareBandwidth")
	flags.SetFlagValues("create-eip-charge-type", "Month", "Year", "Dynamic", "Trial")

	bindProjectIDS(&project, flags)
	bindRegionS(&region, flags)
	bindZoneEmptyS(&zone, &region, flags)

	flags.SetFlagValuesFunc("uhost-id", func() []string {
		return getUhostList([]string{status.HOST_RUNNING, status.HOST_STOPPED, status.HOST_FAIL}, project, region, zone)
	})

	cmd.MarkFlagRequired("uhost-id")

	return cmd
}
