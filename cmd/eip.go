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
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/ucloud/ucloud-sdk-go/services/unet"

	"github.com/spf13/cobra"

	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	. "github.com/ucloud/ucloud-cli/base"
)

//NewCmdEIP ucloud eip
func NewCmdEIP() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "eip",
		Short: "List,allocate and release EIP",
		Long:  `Manipulate EIP, such as list,allocate and release`,
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(NewCmdEIPList())
	cmd.AddCommand(NewCmdEIPAllocate())
	cmd.AddCommand(NewCmdEIPRelease())
	cmd.AddCommand(NewCmdEIPBind())
	cmd.AddCommand(NewCmdEIPUnbind())
	return cmd
}

//EIPRow 表格行
type EIPRow struct {
	Name           string
	IP             string
	ResourceID     string
	Group          string
	Billing        string
	Bandwidth      string
	BindResource   string
	Status         string
	ExpirationTime string
}

//NewCmdEIPList ucloud eip list
func NewCmdEIPList() *cobra.Command {
	req := BizClient.NewDescribeEIPRequest()
	fetchAll := sdk.Bool(false)
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List all EIP instances",
		Long:    `List all EIP instances`,
		Example: "ucloud eip list",
		Run: func(cmd *cobra.Command, args []string) {
			var eipList []unet.UnetEIPSet
			if *fetchAll == true {
				list, err := fetchAllEip(*req.ProjectId, *req.Region)
				if err != nil {
					HandleError(err)
					return
				}
				eipList = list
			} else {
				resp, err := BizClient.DescribeEIP(req)
				if err != nil {
					HandleError(err)
					return
				}
				eipList = resp.EIPSet
			}

			if global.json {
				PrintJSON(eipList)
			} else {
				list := make([]EIPRow, 0)
				for _, eip := range eipList {
					row := EIPRow{}
					row.Name = eip.Name
					for _, ip := range eip.EIPAddr {
						row.IP += ip.IP + " " + ip.OperatorName + "   "
					}
					row.ResourceID = eip.EIPId
					row.Group = eip.Tag
					row.Billing = eip.PayMode
					row.Bandwidth = strconv.Itoa(eip.Bandwidth) + "Mb"
					if eip.Resource.ResourceId != "" {
						row.BindResource = fmt.Sprintf("%s|%s(%s)", eip.Resource.ResourceName, eip.Resource.ResourceId, eip.Resource.ResourceType)
					}
					row.Status = eip.Status
					row.ExpirationTime = time.Unix(int64(eip.ExpireTime), 0).Format("2006-01-02")
					list = append(list, row)
				}
				PrintTable(list, []string{"Name", "IP", "ResourceID", "Group", "Billing", "Bandwidth", "BindResource", "Status", "ExpirationTime"})
			}
		},
	}

	req.ProjectId = cmd.Flags().String("project-id", ConfigInstance.ProjectID, "Assign project-id")
	req.Region = cmd.Flags().String("region", ConfigInstance.Region, "Assign region")
	req.Offset = cmd.Flags().Int("offset", 0, "Optional. Offset default 0")
	req.Limit = cmd.Flags().Int("limit", 50, "Optional. Limit default 50, max value 100")
	fetchAll = cmd.Flags().Bool("list-all", false, "List all eip")
	cmd.Flags().SetFlagValues("list-all", "true", "false")

	return cmd
}

func getEIPIDbyIP(ip net.IP, projectID, region string) (string, error) {
	eipList, err := fetchAllEip(projectID, region)
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

func fetchAllEip(projectID, region string) ([]unet.UnetEIPSet, error) {
	req := BizClient.NewDescribeEIPRequest()
	list := []unet.UnetEIPSet{}
	req.ProjectId = sdk.String(projectID)
	req.Region = sdk.String(region)
	for offset, step := 0, 100; ; offset += step {
		req.Offset = &offset
		req.Limit = &step
		resp, err := BizClient.DescribeEIP(req)
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

func getAllEip(states []string, projectID, region string) []string {
	list, err := fetchAllEip(projectID, region)
	if err != nil {
		return nil
	}
	strs := []string{}
	for _, item := range list {
		ips := []string{}
		for _, ip := range item.EIPAddr {
			ips = append(ips, ip.IP)
		}
		strs = append(strs, item.EIPId+"/"+strings.Join(ips, ","))
	}
	return strs
}

//NewCmdEIPAllocate ucloud eip allocate
func NewCmdEIPAllocate() *cobra.Command {
	var count *int
	var req = BizClient.NewAllocateEIPRequest()
	var cmd = &cobra.Command{
		Use:     "allocate",
		Short:   "Allocate EIP",
		Long:    "Allocate EIP",
		Example: "ucloud eip allocate --line BGP --bandwidth 2",
		Run: func(cmd *cobra.Command, args []string) {
			if *req.OperatorName == "BGP" {
				*req.OperatorName = "Bgp"
			}
			for i := 0; i < *count; i++ {
				resp, err := BizClient.AllocateEIP(req)
				if err != nil {
					HandleError(err)
				} else {
					for _, eip := range resp.EIPSet {
						Cxt.Printf("allocate EIP[%s] ", eip.EIPId)
						for _, ip := range eip.EIPAddr {
							Cxt.Printf("IP:%s  Line:%s \n", ip.IP, ip.OperatorName)
						}
					}
				}
			}
		},
	}
	cmd.Flags().SortFlags = false
	req.OperatorName = cmd.Flags().String("line", "", "Required. 'BGP' or 'International'. 'BGP' could be set in China mainland regions, such as cn-bj2 etc. 'International' could be set in the regions beyond mainland, such as hk, tw-kh, us-ws etc.")
	req.Bandwidth = cmd.Flags().Int("bandwidth-mb", 0, "Required. Bandwidth(Unit:Mbps).The range of value related to network charge mode. By traffic [1, 200]; by bandwidth [1,800] (Unit: Mbps); it could be 0 if the eip belong to the shared bandwidth")
	req.ProjectId = cmd.Flags().String("project-id", ConfigInstance.ProjectID, "Optional. Assign project-id")
	req.Region = cmd.Flags().String("region", ConfigInstance.Region, "Optional. Assign region")
	req.PayMode = cmd.Flags().String("charge-mode", "Bandwidth", "Optional. charge-mode is an enumeration value. 'Traffic','Bandwidth' or 'ShareBandwidth'")
	req.ShareBandwidthId = cmd.Flags().String("share-bandwidth-id", "", "Optional. ShareBandwidthId, required only when charge-mode is 'ShareBandwidth'")
	req.Quantity = cmd.Flags().Int("quantity", 1, "Optional. The duration of the instance. N years/months.")
	req.ChargeType = cmd.Flags().String("charge-type", "Month", "Optional. Enumeration value.'Year',pay yearly;'Month',pay monthly;'Dynamic', pay hourly(requires permission),'Trial', free trial(need permission)")
	req.Tag = cmd.Flags().String("group", "Default", "Optional. Group of your EIP.")
	req.Name = cmd.Flags().String("name", "EIP", "Optional. Name of your EIP.")
	req.Remark = cmd.Flags().String("remark", "", "Optional. Remark of your EIP.")
	req.CouponId = cmd.Flags().String("coupon-id", "", "Optional. Coupon ID, The Coupon can deducte part of the payment")
	count = cmd.Flags().Int("count", 1, "Optional. Count of EIP to allocate")

	cmd.Flags().SetFlagValues("line", "BGP", "International")
	cmd.Flags().SetFlagValues("charge-mode", "Bandwidth", "Traffic", "ShareBandwidth")
	cmd.Flags().SetFlagValues("charge-type", "Month", "Year", "Dynamic", "Trial")
	cmd.MarkFlagRequired("line")
	cmd.MarkFlagRequired("bandwidth-mb")
	return cmd
}

//NewCmdEIPBind ucloud eip bind
func NewCmdEIPBind() *cobra.Command {
	var projectID, region, eipID, resourceID, resourceType *string
	cmd := &cobra.Command{
		Use:     "bind",
		Short:   "Bind EIP with uhost",
		Long:    "Bind EIP with uhost",
		Example: "ucloud eip bind --eip-id eip-xxx --resource-id uhost-xxx",
		Run: func(cmd *cobra.Command, args []string) {
			bindEIP(resourceID, resourceType, eipID, projectID, region)
		},
	}
	cmd.Flags().SortFlags = false
	eipID = cmd.Flags().String("eip-id", "", "Required. EIPId to bind")
	resourceID = cmd.Flags().String("resource-id", "", "Required. ResourceID , which is the UHostId of uhost")
	resourceType = cmd.Flags().String("resource-type", "uhost", "Requried. ResourceType, type of resource to bind with eip. 'uhost','vrouter','ulb','upm','hadoophost'.eg..")
	projectID = cmd.Flags().String("project-id", ConfigInstance.ProjectID, "Optional. Assign project-id")
	region = cmd.Flags().String("region", ConfigInstance.Region, "Optional. Assign region")
	cmd.MarkFlagRequired("eip-id")
	cmd.MarkFlagRequired("resource-id")
	cmd.Flags().SetFlagValues("resource-type", "uhost", "vrouter", "ulb", "upm", "hadoophost", "fortresshost", "udockhost", "udhost", "natgw", "udb", "vpngw", "ucdr", "dbaudit")
	return cmd
}

func bindEIP(resourceID, resourceType, eipID, projectID, region *string) {
	req := BizClient.NewBindEIPRequest()
	req.ResourceId = resourceID
	req.ResourceType = resourceType
	req.EIPId = eipID
	req.ProjectId = projectID
	req.Region = region
	_, err := BizClient.BindEIP(req)
	if err != nil {
		HandleError(err)
	} else {
		Cxt.Printf("bind EIP[%s] with %s[%s]\n", *req.EIPId, *req.ResourceType, *req.ResourceId)
	}
}

//NewCmdEIPUnbind ucloud eip unbind
func NewCmdEIPUnbind() *cobra.Command {

	var req = BizClient.NewUnBindEIPRequest()
	var cmd = &cobra.Command{
		Use:     "unbind",
		Short:   "Unbind EIP with uhost",
		Long:    "Unbind EIP with uhost",
		Example: "ucloud eip unbind --eip-id eip-xxx --resource-id uhost-xxx",
		Run: func(cmd *cobra.Command, args []string) {
			req.ResourceType = sdk.String("uhost")
			_, err := BizClient.UnBindEIP(req)
			if err != nil {
				HandleError(err)
			} else {
				Cxt.Printf("unbind EIP[%s] with %s[%s]\n", *req.EIPId, *req.ResourceType, *req.ResourceId)
			}
		},
	}
	cmd.Flags().SortFlags = false
	req.EIPId = cmd.Flags().String("eip-id", "", "Required. EIPId to unbind")
	req.ResourceId = cmd.Flags().String("resource-id", "", "Required. ResourceID , which is the UHostId of uhost")
	req.ProjectId = cmd.Flags().String("project-id", ConfigInstance.ProjectID, "Optional. Assign project-id")
	req.Region = cmd.Flags().String("region", ConfigInstance.Region, "Optional. Assign region")
	cmd.MarkFlagRequired("eip-id")
	cmd.MarkFlagRequired("resource-id")

	return cmd
}

//NewCmdEIPRelease ucloud eip release
func NewCmdEIPRelease() *cobra.Command {
	var ids []string
	var req = BizClient.NewReleaseEIPRequest()
	var cmd = &cobra.Command{
		Use:     "release",
		Short:   "Release EIP",
		Long:    "Release EIP",
		Example: "ucloud eip release --eip-id eip-xx1,eip-xx2",
		Run: func(cmd *cobra.Command, args []string) {
			for _, id := range ids {
				req.EIPId = &id
				_, err := BizClient.ReleaseEIP(req)
				if err != nil {
					HandleError(err)
				} else {
					Cxt.Printf("released EIP[%v]\n", *req.EIPId)
				}
			}
		},
	}
	cmd.Flags().SortFlags = false
	cmd.Flags().StringSliceVarP(&ids, "eip-id", "", make([]string, 0), "Required. EIPIds of the EIP you want to release")
	req.ProjectId = cmd.Flags().String("project-id", ConfigInstance.ProjectID, "Optional. Assign project-id")
	req.Region = cmd.Flags().String("region", ConfigInstance.Region, "Optional. Assign region")
	cmd.MarkFlagRequired("eip-id")
	cmd.MarkFlagRequired("bandwidth")
	return cmd
}
