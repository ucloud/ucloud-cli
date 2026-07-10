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
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/cmd/internal/platform"
	"github.com/ucloud/ucloud-cli/model/status"
	"github.com/ucloud/ucloud-cli/pkg/command"
	"github.com/ucloud/ucloud-sdk-go/services/uhost"
	"github.com/ucloud/ucloud-sdk-go/services/unet"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"
)

// NewCmdExt ucloud ext
func NewCmdExt() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ext",
		Short: "extended commands of UCloud CLI",
		Long:  "extended commands of UCloud CLI",
	}
	cmd.AddCommand(NewCmdExtUHost())
	return cmd
}

// NewCmdExtUHost ucloud ext uhost
func NewCmdExtUHost() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uhost",
		Short: "extended uhost commands",
		Long:  "extended uhost commands",
	}
	cmd.AddCommand(NewCmdExtUHostSwitchEIP())
	return cmd
}

// NewCmdExtUHostSwitchEIP ucloud ext uhost switch-eip
func NewCmdExtUHostSwitchEIP() *cobra.Command {
	var project, region, zone, chargeType, trafficMode, shareBandwidthID string
	var uhostIDs, eipAddrs []string
	var eipBandwidth, quntity int
	var unbind, release bool

	cmd := &cobra.Command{
		Use:     "switch-eip",
		Short:   "Switch EIP for UHost instances",
		Long:    "Switch EIP for UHost instances",
		Example: "ucloud ext uhost switch-eip --uhost-id uhost-1n1sxx2,uhost-li4jxx1 --create-eip-bandwidth-mb 2",
		Run: func(c *cobra.Command, args []string) {
			unetClient := newServiceClient(unet.NewClient)
			project = platform.PickResourceID(project)
			eipAddrMap := make(map[string]bool)
			for _, addr := range eipAddrs {
				eipAddrMap[addr] = true
			}
			logs := make([]string, 0)
			for _, idname := range uhostIDs {
				uhostID := platform.PickResourceID(idname)
				logs = append(logs, fmt.Sprintf("describe uhost instance by uhostID %s", uhostID))
				ins, err := extDescribeUHostByID(uhostID, project, region, zone)
				if err != nil {
					errStr := fmt.Sprintf("describe uhost %s failed: %v", uhostID, err)
					platform.HandleError(errors.New(errStr))
					logs = append(logs, errStr)
					continue
				}
				uhostIns, ok := ins.(*uhost.UHostInstanceSet)
				if !ok {
					errStr := fmt.Sprintf("uhost %s does not exist", uhostID)
					platform.HandleError(errors.New(errStr))
					logs = append(logs, errStr)
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
					req := unetClient.NewAllocateEIPRequest()
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
							errStr := "create-eip-share-bandwidth-id should not be empty when create-eip-traffic-mode is assigned 'ShareBandwidth'"
							logs = append(logs, errStr)
							platform.HandleError(errors.New(errStr))
							return
						}
					}
					logs = append(logs, fmt.Sprintf("api AllocateEIP, request:%v", platform.ToQueryMap(req)))
					resp, err := unetClient.AllocateEIP(req)
					if err != nil {
						errStr := fmt.Sprintf("allocate EIP failed: %v", err)
						logs = append(logs, errStr)
						platform.HandleError(errors.New(errStr))
						continue
					}
					if len(resp.EIPSet) != 1 {
						errStr := fmt.Sprintf("allocate EIP failed, length of eip set is not 1")
						platform.HandleError(errors.New(errStr))
						logs = append(logs, errStr)
						continue
					}
					eipID := resp.EIPSet[0].EIPId
					eipRet := fmt.Sprintf("allocated new eip %s|%s", eipID, resp.EIPSet[0].EIPAddr[0].IP)
					logs = append(logs, eipRet)
					fmt.Println(eipRet)

					//绑定新EIP
					slogs, err2 := extAttachEIPWithLogs(&uhostID, sdk.String("uhost"), &eipID, &project, &region)
					logs = append(logs, slogs...)
					if err2 != nil {
						platform.HandleError(fmt.Errorf("bind new eip %s failed: %v", eipID, err2))
						continue
					}
					fmt.Printf("bound eip %s with uhost %s\n", eipID, uhostID)

					if unbind {
						slogs, err := extDetachEIPWithLogs(uhostID, "uhost", ip.IPId, project, region)
						logs = append(logs, slogs...)
						if err != nil {
							platform.HandleError(fmt.Errorf("unbind eip %s failed: %v", ip.IPId, err))
							continue
						}
						fmt.Printf("unbound eip %s|%s with uhost %s\n", ip.IPId, ip.IP, uhostID)
					}

					if release {
						req := unetClient.NewReleaseEIPRequest()
						req.ProjectId = &project
						req.Region = &region
						req.EIPId = sdk.String(ip.IPId)
						logs = append(logs, fmt.Sprintf("api ReleaseEIP, request:%v", platform.ToQueryMap(req)))
						_, err := unetClient.ReleaseEIP(req)
						if err != nil {
							errStr := fmt.Sprintf("release eip %s failed: %v", ip.IPId, err)
							logs = append(logs, errStr)
							platform.HandleError(errors.New(errStr))
							continue
						}
						releaseRet := fmt.Sprintf("released eip %s|%s", ip.IPId, ip.IP)
						logs = append(logs, releaseRet)
						fmt.Println(releaseRet)
					}
					platform.LogInfo(logs...)
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

	command.SetFlagValues(cmd, "create-eip-traffic-mode", "Bandwidth", "Traffic", "ShareBandwidth")
	command.SetFlagValues(cmd, "create-eip-charge-type", "Month", "Year", "Dynamic", "Trial")

	bindProjectIDS(&project, cmd)
	bindRegionS(&region, cmd)
	bindZoneEmptyS(&zone, &region, cmd)

	command.SetCompletion(cmd, "uhost-id", func() []string {
		return extUHostList([]string{status.HOST_RUNNING, status.HOST_STOPPED, status.HOST_FAIL}, project, region, zone)
	})

	cmd.MarkFlagRequired("uhost-id")

	return cmd
}

func extDescribeUHostByID(uhostID, projectID, region, zone string) (interface{}, error) {
	client := newServiceClient(uhost.NewClient)
	req := client.NewDescribeUHostInstanceRequest()
	req.UHostIds = []string{uhostID}
	req.ProjectId = &projectID
	req.Region = &region
	req.Zone = &zone

	resp, err := client.DescribeUHostInstance(req)
	if err != nil {
		return nil, err
	}
	if len(resp.UHostSet) < 1 {
		return nil, fmt.Errorf("uhost [%s] does not exist", uhostID)
	}

	return &resp.UHostSet[0], nil
}

func extUHostList(states []string, project, region, zone string) []string {
	client := newServiceClient(uhost.NewClient)
	req := client.NewDescribeUHostInstanceRequest()
	req.ProjectId = sdk.String(project)
	req.Region = sdk.String(region)
	req.Zone = sdk.String(zone)
	req.Limit = sdk.Int(50)
	resp, err := client.DescribeUHostInstance(req)
	if err != nil {
		return nil
	}
	list := []string{}
	for _, host := range resp.UHostSet {
		if states != nil {
			for _, s := range states {
				if host.State == s {
					list = append(list, host.UHostId+"/"+strings.Replace(host.Name, " ", "-", -1))
				}
			}
		} else {
			list = append(list, host.UHostId+"/"+strings.Replace(host.Name, " ", "-", -1))
		}
	}
	return list
}

func extEIPIDByIP(ip net.IP, projectID, region string) (string, error) {
	eipList, err := extFetchAllEIP(projectID, region)
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

func extFetchAllEIP(projectID, region string) ([]unet.UnetEIPSet, error) {
	client := newServiceClient(unet.NewClient)
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

func extAttachEIPWithLogs(resourceID, resourceType, eipID, projectID, region *string) ([]string, error) {
	logs := make([]string, 0)
	ip := net.ParseIP(*eipID)
	if ip != nil {
		id, err := extEIPIDByIP(ip, *projectID, *region)
		if err != nil {
			platform.HandleError(err)
		} else {
			*eipID = id
		}
	}
	client := newServiceClient(unet.NewClient)
	req := client.NewBindEIPRequest()
	req.ResourceId = resourceID
	req.ResourceType = resourceType
	req.EIPId = sdk.String(platform.PickResourceID(*eipID))
	req.ProjectId = sdk.String(platform.PickResourceID(*projectID))
	req.Region = region
	logs = append(logs, fmt.Sprintf("api: BindEIP, request: %v", platform.ToQueryMap(req)))
	_, err := client.BindEIP(req)
	if err != nil {
		logs = append(logs, fmt.Sprintf("bind eip failed: %v", err))
		return logs, err
	}
	logs = append(logs, fmt.Sprintf("bind eip[%s] with %s[%s] successfully", *req.EIPId, *req.ResourceType, *req.ResourceId))
	return logs, nil
}

func extDetachEIPWithLogs(resourceID, resourceType, eipID, projectID, region string) ([]string, error) {
	logs := make([]string, 0)
	eipID = platform.PickResourceID(eipID)
	ip := net.ParseIP(eipID)
	if ip != nil {
		id, err := extEIPIDByIP(ip, projectID, region)
		if err != nil {
			platform.HandleError(err)
		} else {
			eipID = id
		}
	}
	client := newServiceClient(unet.NewClient)
	req := client.NewUnBindEIPRequest()
	req.ResourceId = &resourceID
	req.ResourceType = &resourceType
	req.EIPId = &eipID
	req.ProjectId = sdk.String(platform.PickResourceID(projectID))
	req.Region = &region
	logs = append(logs, fmt.Sprintf("api: UnBindEIP, request: %v", platform.ToQueryMap(req)))
	_, err := client.UnBindEIP(req)
	if err != nil {
		logs = append(logs, fmt.Sprintf("unbind eip failed: %v", err))
		return logs, err
	}
	logs = append(logs, fmt.Sprintf("unbind eip[%s] with %s[%s] successfully", *req.EIPId, *req.ResourceType, *req.ResourceId))
	return logs, nil
}
