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
	"strconv"
	"time"

	"github.com/ucloud/ucloud-sdk-go/sdk"

	"github.com/spf13/cobra"
	. "github.com/ucloud/ucloud-cli/util"
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
	UGroup         string
	Billing        string
	Bandwidth      string
	BindResource   string
	Status         string
	ExpirationTime string
}

//NewCmdEIPList ucloud eip ls
func NewCmdEIPList() *cobra.Command {
	req := BizClient.NewDescribeEIPRequest()
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List all EIP instances",
		Long:    `List all EIP instances`,
		Example: "ucloud eip ls",
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := BizClient.DescribeEIP(req)
			if err != nil {
				Cxt.PrintErr(err)
			} else {
				if resp.RetCode == 0 {
					if global.json {
						PrintJSON(resp.EIPSet)
					} else {
						list := make([]EIPRow, 0)
						for _, eip := range resp.EIPSet {
							row := EIPRow{}
							row.Name = eip.Name
							for _, ip := range eip.EIPAddr {
								row.IP += ip.IP + " " + ip.OperatorName + "   "
							}
							row.ResourceID = eip.EIPId
							row.UGroup = eip.Tag
							row.Billing = eip.PayMode
							row.Bandwidth = strconv.Itoa(eip.Bandwidth) + "Mb"
							row.BindResource = fmt.Sprintf("%s(%s)", eip.Resource.ResourceName, eip.Resource.ResourceType)
							row.Status = eip.Status
							row.ExpirationTime = time.Unix(int64(eip.ExpireTime), 0).Format("2006-01-02")
							list = append(list, row)
						}
						PrintTable(list, []string{"Name", "IP", "ResourceID", "UGroup", "Billing", "Bandwidth", "BindResource", "Status", "ExpirationTime"})
					}
				} else {
					HandleBizError(resp)
				}
			}
		},
	}
	req.Region = cmd.Flags().String("region", ConfigInstance.Region, "Assign region(override default region of your config)")
	req.ProjectId = cmd.Flags().String("project-id", ConfigInstance.ProjectID, "Assign project-id(override default projec-id of your config)")
	return cmd
}

//NewCmdEIPAllocate ucloud eip allocate
func NewCmdEIPAllocate() *cobra.Command {
	var req = BizClient.NewAllocateEIPRequest()
	var cmd = &cobra.Command{
		Use:     "allocate",
		Short:   "Allocate EIP",
		Long:    "Allocate EIP",
		Example: "ucloud eip allocate --line Bgp --bandwidth 2",
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := BizClient.AllocateEIP(req)
			if err != nil {
				Cxt.PrintErr(err)
			} else {
				if resp.RetCode == 0 {
					for _, eip := range resp.EIPSet {
						Cxt.Printf("EIPId:%s,", eip.EIPId)
						for _, ip := range eip.EIPAddr {
							Cxt.Printf("IP:%s,Line:%s \n", ip.IP, ip.OperatorName)
						}
					}
				} else {
					HandleBizError(resp)
				}
			}
		},
	}
	cmd.Flags().SortFlags = false
	req.Region = cmd.Flags().String("region", ConfigInstance.Region, "Assign region(override default region of your config)")
	req.ProjectId = cmd.Flags().String("project-id", ConfigInstance.ProjectID, "Assign project-id(override default projec-id of your config)")
	req.OperatorName = cmd.Flags().String("line", "", "Line 'Bgp' or 'International'. 'Bgp' can be set in region cn-sh1,cn-sh2,cn-gd,cn-bj1 and cn-bj2. 'International' can be set in region hk,us-ca,th-bkk,kr-seoul,us-ws,ge-fra,sg,tw-kh and other oversea regions. Required")
	req.Bandwidth = cmd.Flags().Int("bandwidth", 0, "Bandwidth(Unit:Mbps). When paying by traffic, it ranges from 1 to 200; when paying by bandwidth, it ranges from 1 to 800, and when shared bandwidth is used, its value is 0. Required")
	req.PayMode = cmd.Flags().String("pay-mode", "Bandwidth", "pay-mode is an enumeration value. 'Traffic','Bandwidth' or 'ShareBandwidth'")
	req.Quantity = cmd.Flags().Int("quantity", 1, "The quantity of EIP")
	req.ChargeType = cmd.Flags().String("charge-type", "Month", "charge-type is an enumeration value. 'Year','Month', 'Dynamic'(Pay by the hour), 'Trial'(Need permission)")
	req.Tag = cmd.Flags().String("tag", "Default", "Tag of your EIP.")
	req.Name = cmd.Flags().String("name", "EIP", "Name of your EIP.")
	req.Remark = cmd.Flags().String("remark", "", "Remark of your EIP.")
	req.CouponId = cmd.Flags().String("coupon-id", "", "Coupon ID, The Coupon can deducte part of the payment")
	cmd.MarkFlagRequired("line")
	cmd.MarkFlagRequired("bandwidth")
	return cmd
}

//NewCmdEIPBind ucloud eip bind
func NewCmdEIPBind() *cobra.Command {
	var req = BizClient.NewBindEIPRequest()
	var cmd = &cobra.Command{
		Use:     "bind",
		Short:   "Bind EIP with uhost",
		Long:    "Bind EIP with uhost",
		Example: "ucloud eip bind --eip-id eip-xxx --resource-id uhost-xxx",
		Run: func(cmd *cobra.Command, args []string) {
			req.ResourceType = sdk.String("uhost")
			resp, err := BizClient.BindEIP(req)
			if err != nil {
				Cxt.PrintErr(err)
			} else {
				if resp.RetCode == 0 {
					fmt.Printf("EIP: %s bind with %s:%s successfully \n", *req.EIPId, *req.ResourceType, *req.ResourceId)
				} else {
					HandleBizError(resp)
				}
			}
		},
	}
	cmd.Flags().SortFlags = false
	req.Region = cmd.Flags().String("region", ConfigInstance.Region, "Assign region(override default region of your config)")
	req.ProjectId = cmd.Flags().String("project-id", ConfigInstance.ProjectID, "Assign project-id(override default projec-id of your config)")
	req.EIPId = cmd.Flags().String("eip-id", "", "EIPId to bind. Required")
	req.ResourceId = cmd.Flags().String("resource-id", "", "ResourceID , which is the UHostId of uhost. Required")
	cmd.MarkFlagRequired("eip-id")
	cmd.MarkFlagRequired("resource-id")
	return cmd
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
			resp, err := BizClient.UnBindEIP(req)
			if err != nil {
				Cxt.PrintErr(err)
			} else {
				if resp.RetCode == 0 {
					Cxt.Printf("EIP: %s unbind with %s:%s successfully \n", *req.EIPId, *req.ResourceType, *req.ResourceId)
				} else {
					HandleBizError(resp)
				}
			}
		},
	}
	cmd.Flags().SortFlags = false
	req.Region = cmd.Flags().String("region", ConfigInstance.Region, "Assign region(override default region of your config)")
	req.ProjectId = cmd.Flags().String("project-id", ConfigInstance.ProjectID, "Assign project-id(override default projec-id of your config)")
	req.EIPId = cmd.Flags().String("eip-id", "", "EIPId to unbind. Required")
	req.ResourceId = cmd.Flags().String("resource-id", "", "ResourceID , which is the UHostId of uhost. Required")
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
		Example: "ucloud eip release --eip-id eip-xx1 --eip-id eip-xx2",
		Run: func(cmd *cobra.Command, args []string) {
			for _, id := range ids {
				req.EIPId = &id
				resp, err := BizClient.ReleaseEIP(req)
				if err != nil {
					Cxt.PrintErr(err)
				} else {
					if resp.RetCode == 0 {
						Cxt.Printf("EIP: %v released \n", req.EIPId)
					} else {
						HandleBizError(resp)
					}
				}
			}
		},
	}
	req.Region = cmd.Flags().String("region", ConfigInstance.Region, "Assign region(override default region of your config)")
	req.ProjectId = cmd.Flags().String("project-id", ConfigInstance.ProjectID, "Assign project-id(override default projec-id of your config)")
	cmd.Flags().StringArrayVarP(&ids, "eip-id", "", make([]string, 0), "EIPId of the EIP you want to release. Required")
	cmd.MarkFlagRequired("eip-id")
	cmd.MarkFlagRequired("bandwidth")
	return cmd
}
