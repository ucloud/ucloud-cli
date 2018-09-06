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

	"github.com/spf13/cobra"
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

//NewCmdEIPList ucloud eip ls
func NewCmdEIPList() *cobra.Command {
	req := client.NewDescribeEIPRequest()
	var cmd = &cobra.Command{
		Use:     "ls",
		Short:   "List all EIP instances",
		Long:    `List all EIP instances`,
		Example: "ucloud eip ls",
		Run: func(cmd *cobra.Command, args []string) {
			bindGlobalParam(req)
			resp, err := client.DescribeEIP(req)
			if err != nil {
				fmt.Println("Error:", err)
				return
			}
			if resp.RetCode == 0 {
				for _, eip := range resp.EIPSet {
					fmt.Printf("EIPId:%s, ", eip.EIPId)
					for _, ip := range eip.EIPAddr {
						fmt.Printf("IP:%-15s, Line:%s, ", ip.IP, ip.OperatorName)
					}
					fmt.Printf("BindResource: %s \n", eip.Resource.ResourceId)
				}
			} else {
				fmt.Printf("Something wrong. RetCode:%d, Message: %s \n", resp.RetCode, resp.Message)
			}
		},
	}
	return cmd
}

//NewCmdEIPAllocate ucloud eip allocate
func NewCmdEIPAllocate() *cobra.Command {
	var eipAllocateReq = client.NewAllocateEIPRequest()
	var cmd = &cobra.Command{
		Use:     "allocate",
		Short:   "Allocate EIP",
		Long:    "Allocate EIP",
		Example: "ucloud eip allocate --line Bgp --bandwidth 2",
		Run: func(cmd *cobra.Command, args []string) {
			bindGlobalParam(eipAllocateReq)
			resp, err := client.AllocateEIP(eipAllocateReq)
			if err != nil {
				fmt.Println(err)
			} else {
				if resp.RetCode == 0 {
					for _, eip := range resp.EIPSet {
						fmt.Printf("EIPId:%s,", eip.EIPId)
						for _, ip := range eip.EIPAddr {
							fmt.Printf("IP:%s,Line:%s \n", ip.IP, ip.OperatorName)
						}
					}
				} else {
					fmt.Printf("Something wrong. RetCode:%d, Message: %s \n", resp.RetCode, resp.Message)
				}
			}
		},
	}
	cmd.Flags().SortFlags = false
	cmd.Flags().StringVar(&eipAllocateReq.OperatorName, "line", "", "Line 'Bgp' or 'International'. 'Bgp' can be set in region cn-sh1,cn-sh2,cn-gd,cn-bj1 and cn-bj2. 'International' can be set in region hk,us-ca,th-bkk,kr-seoul,us-ws,ge-fra,sg,tw-kh and other oversea regions. Required")
	cmd.Flags().IntVar(&eipAllocateReq.Bandwidth, "bandwidth", 0, "Bandwidth(Unit:Mbps). When paying by traffic, it ranges from 1 to 200; when paying by bandwidth, it ranges from 1 to 800, and when shared bandwidth is used, its value is 0. Required")
	cmd.Flags().StringVar(&eipAllocateReq.PayMode, "pay-mode", "Bandwidth", "pay-mode is an enumeration value. 'Traffic','Bandwidth' or 'ShareBandwidth'")
	cmd.Flags().IntVar(&eipAllocateReq.Quantity, "quantity", 1, "The quantity of EIP")
	cmd.Flags().StringVar(&eipAllocateReq.ChargeType, "charge-type", "Month", "charge-type is an enumeration value. 'Year','Month', 'Dynamic'(Pay by the hour), 'Trial'(Need permission)")
	cmd.Flags().StringVar(&eipAllocateReq.Tag, "tag", "Default", "Tag of your EIP.")
	cmd.Flags().StringVar(&eipAllocateReq.Name, "name", "EIP", "Name of your EIP.")
	cmd.Flags().StringVar(&eipAllocateReq.Remark, "remark", "", "Remark of your EIP.")
	cmd.Flags().StringVar(&eipAllocateReq.CouponId, "coupon-id", "", "Coupon ID, The Coupon can deducte part of the payment")
	cmd.MarkFlagRequired("line")
	cmd.MarkFlagRequired("bandwidth")
	return cmd
}

//NewCmdEIPBind ucloud eip bind
func NewCmdEIPBind() *cobra.Command {
	var eipBindReq = client.NewBindEIPRequest()
	var eipBindCmd = &cobra.Command{
		Use:     "bind",
		Short:   "Bind EIP with uhost",
		Long:    "Bind EIP with uhost",
		Example: "ucloud eip bind --eip-id eip-xxx --resource-id uhost-xxx",
		Run: func(cmd *cobra.Command, args []string) {
			bindGlobalParam(eipBindReq)
			eipBindReq.ResourceType = "uhost"
			resp, err := client.BindEIP(eipBindReq)
			if err != nil {
				fmt.Println(err)
			} else {
				if resp.RetCode == 0 {
					fmt.Printf("EIP: %v bind with %v:%v successfully \n", eipBindReq.EIPId, eipBindReq.ResourceType, eipBindReq.ResourceId)
				} else {
					fmt.Printf("Something wrong. RetCode:%d, Message: %s \n", resp.RetCode, resp.Message)
				}
			}
		},
	}
	eipBindCmd.Flags().SortFlags = false
	eipBindCmd.Flags().StringVar(&eipBindReq.EIPId, "eip-id", "", "EIPId to bind. Required")
	eipBindCmd.Flags().StringVar(&eipBindReq.ResourceId, "resource-id", "", "ResourceID , which is the UHostId of uhost. Required")
	eipBindCmd.MarkFlagRequired("eip-id")
	eipBindCmd.MarkFlagRequired("resource-id")
	return eipBindCmd
}

//NewCmdEIPUnbind ucloud eip unbind
func NewCmdEIPUnbind() *cobra.Command {

	var eipUnBindReq = client.NewUnBindEIPRequest()
	var eipUnBindCmd = &cobra.Command{
		Use:     "unbind",
		Short:   "Unbind EIP with uhost",
		Long:    "Unbind EIP with uhost",
		Example: "ucloud eip unbind --eip-id eip-xxx --resource-id uhost-xxx",
		Run: func(cmd *cobra.Command, args []string) {
			bindGlobalParam(eipUnBindReq)
			eipUnBindReq.ResourceType = "uhost"
			resp, err := client.UnBindEIP(eipUnBindReq)
			if err != nil {
				fmt.Println(err)
			} else {
				if resp.RetCode == 0 {
					fmt.Printf("EIP: %v unbind with %v:%v successfully \n", eipUnBindReq.EIPId, eipUnBindReq.ResourceType, eipUnBindReq.ResourceId)
				} else {
					fmt.Printf("Something wrong. RetCode:%d, Message: %s \n", resp.RetCode, resp.Message)
				}
			}
		},
	}
	eipUnBindCmd.Flags().SortFlags = false
	eipUnBindCmd.Flags().StringVar(&eipUnBindReq.EIPId, "eip-id", "", "EIPId to unbind. Required")
	eipUnBindCmd.Flags().StringVar(&eipUnBindReq.ResourceId, "resource-id", "", "ResourceID , which is the UHostId of uhost. Required")
	eipUnBindCmd.MarkFlagRequired("eip-id")
	eipUnBindCmd.MarkFlagRequired("resource-id")

	return eipUnBindCmd
}

//NewCmdEIPRelease ucloud eip release
func NewCmdEIPRelease() *cobra.Command {
	var ids []string
	var eipReleaseCmd = &cobra.Command{
		Use:     "release",
		Short:   "Release EIP",
		Long:    "Release EIP",
		Example: "ucloud eip release --eip-id eip-xx1 --eip-id eip-xx2",
		Run: func(cmd *cobra.Command, args []string) {
			for _, id := range ids {
				var eipReleaseReq = client.NewReleaseEIPRequest()
				bindGlobalParam(eipReleaseReq)
				eipReleaseReq.EIPId = id
				resp, err := client.ReleaseEIP(eipReleaseReq)
				if err != nil {
					fmt.Println(err)
				} else {
					if resp.RetCode == 0 {
						fmt.Printf("EIP: %v released \n", eipReleaseReq.EIPId)
					} else {
						fmt.Printf("Something wrong. RetCode:%d, Message: %s \n", resp.RetCode, resp.Message)
					}
				}
			}
		},
	}
	eipReleaseCmd.Flags().StringArrayVarP(&ids, "eip-id", "", make([]string, 0), "EIPId of the EIP you want to release. Required")
	eipReleaseCmd.MarkFlagRequired("eip-id")
	eipReleaseCmd.MarkFlagRequired("bandwidth")
	return eipReleaseCmd
}
