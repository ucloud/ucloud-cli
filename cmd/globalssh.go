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
	"net"
	"strings"

	"github.com/spf13/cobra"

	. "github.com/ucloud/ucloud-cli/base"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"
)

//NewCmdGssh ucloud gssh
func NewCmdGssh() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gssh",
		Short: "Create,list,update and delete globalssh instance",
		Long:  `Create,list,update and delete globalssh instance`,
	}
	cmd.AddCommand(NewCmdGsshList())
	cmd.AddCommand(NewCmdGsshCreate())
	cmd.AddCommand(NewCmdGsshDelete())
	cmd.AddCommand(NewCmdGsshModify())
	cmd.AddCommand(NewCmdGsshArea())
	return cmd
}

//GSSHRow gssh表格行
type GSSHRow struct {
	ResourceID         string
	SSHServerIP        string
	AcceleratingDomain string
	SSHServerLocation  string
	SSHPort            int
	Remark             string
}

//NewCmdGsshList ucloud gssh list
func NewCmdGsshList() *cobra.Command {
	req := BizClient.NewDescribeGlobalSSHInstanceRequest()
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List all GlobalSSH instances",
		Long:    `List all GlobalSSH instances`,
		Example: "ucloud gssh list",
		Run: func(cmd *cobra.Command, args []string) {
			var areaMap = map[string]string{
				"洛杉矶":  "LosAngeles",
				"新加坡":  "Singapore",
				"香港":   "HongKong",
				"东京":   "Tokyo",
				"华盛顿":  "Washington",
				"法兰克福": "Frankfurt",
			}

			resp, err := BizClient.DescribeGlobalSSHInstance(req)
			if err != nil {
				HandleError(err)
			} else {
				if global.json {
					PrintJSON(resp.InstanceSet)
				} else {
					list := make([]GSSHRow, 0)
					for _, gssh := range resp.InstanceSet {
						row := GSSHRow{}
						row.ResourceID = gssh.InstanceId
						row.SSHServerIP = gssh.TargetIP
						row.AcceleratingDomain = gssh.AcceleratingDomain
						row.SSHPort = gssh.Port
						row.Remark = gssh.Remark
						if val, ok := areaMap[gssh.Area]; ok {
							row.SSHServerLocation = val
						} else {
							row.SSHServerLocation = gssh.Area
						}
						list = append(list, row)
					}
					PrintTable(list, []string{"ResourceID", "SSHServerIP", "AcceleratingDomain", "SSHServerLocation", "SSHPort", "Remark"})
				}
			}
		},
	}
	cmd.Flags().SortFlags = false
	req.Region = cmd.Flags().String("region", ConfigInstance.Region, "Optional. Assign region")
	req.ProjectId = cmd.Flags().String("project-id", ConfigInstance.ProjectID, "Optional. Assign project-id")
	return cmd
}

//NewCmdGsshArea ucloud gssh area
func NewCmdGsshArea() *cobra.Command {
	req := BizClient.NewDescribeGlobalSSHAreaRequest()
	cmd := &cobra.Command{
		Use:   "location",
		Short: "List SSH server locations and covered areas",
		Long:  "List SSH server locations and covered areas",
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := BizClient.DescribeGlobalSSHArea(req)
			if err != nil {
				HandleError(err)
				return
			}
			list := make([]GsshLocation, 0)
			for _, item := range resp.AreaSet {
				row := GsshLocation{
					AirportCode:       item.AreaCode,
					SSHServerLocation: areaCodeMap[item.AreaCode],
				}
				regionLabels := make([]string, 0)
				for _, region := range item.RegionSet {
					regionLabels = append(regionLabels, RegionLabel[region])
				}
				row.CoveredArea = strings.Join(regionLabels, ",")
				list = append(list, row)
			}

			PrintTable(list, []string{"AirportCode", "SSHServerLocation", "CoveredArea"})
		},
	}
	return cmd
}

//GsshLocation 服务地点和覆盖区域
type GsshLocation struct {
	AirportCode       string
	SSHServerLocation string
	CoveredArea       string
}

var areaCodeMap = map[string]string{
	"LAX": "LosAngeles",
	"SIN": "Singapore",
	"HKG": "HongKong",
	"HND": "Tokyo",
	"IAD": "Washington",
	"FRA": "Frankfurt",
}

//NewCmdGsshCreate ucloud gssh create
func NewCmdGsshCreate() *cobra.Command {
	var targetIP *net.IP
	req := BizClient.NewCreateGlobalSSHInstanceRequest()
	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Create GlobalSSH instance",
		Long:    "Create GlobalSSH instance",
		Example: "ucloud gssh create --location Washington --target-ip 8.8.8.8",
		Run: func(cmd *cobra.Command, args []string) {
			port := *req.Port
			for code, area := range areaCodeMap {
				if area == *req.AreaCode {
					*req.AreaCode = code
				}
			}
			if port < 1 || port > 65535 || port == 80 || port == 443 {
				Cxt.Println("The port number should be between 1 and 65535, and cannot be 80 or 443")
				return
			}
			req.TargetIP = sdk.String(targetIP.String())
			resp, err := BizClient.CreateGlobalSSHInstance(req)
			if err != nil {
				HandleError(err)
			} else {
				Cxt.Printf("gssh[%s] created\n", resp.InstanceId)
			}
		},
	}
	cmd.Flags().SortFlags = false
	req.AreaCode = cmd.Flags().String("location", "", "Required. Location of the source server. See 'ucloud gssh location'")
	targetIP = cmd.Flags().IP("target-ip", nil, "Required. IP of the source server. Required")
	req.Region = cmd.Flags().String("region", "", "Optional. Assign region")
	req.ProjectId = cmd.Flags().String("project-id", ConfigInstance.ProjectID, "Optional. Assign project-id")
	req.Port = cmd.Flags().Int("port", 22, "Optional. Port of The SSH service between 1 and 65535. Do not use ports such as 80,443.")
	req.Remark = cmd.Flags().String("remark", "", "Optional. Remark of your GlobalSSH.")
	req.ChargeType = cmd.Flags().String("charge-type", "Month", "Optional.'Year',pay yearly;'Month',pay monthly;'Dynamic', pay hourly(requires access)")
	req.Quantity = cmd.Flags().Int("quantity", 1, "Optional. The duration of the instance. N years/months.")
	req.CouponId = cmd.Flags().String("coupon-id", "", "Optional. Coupon ID, The Coupon can deduct part of the payment,see DescribeCoupon or https://accountv2.ucloud.cn")
	cmd.MarkFlagRequired("location")
	cmd.MarkFlagRequired("target-ip")
	cmd.Flags().SetFlagValues("location", "LosAngeles", "Singapore", "HongKong", "Tokyo", "Washington", "Frankfurt")
	cmd.Flags().SetFlagValues("charge-type", "Month", "Year", "Dynamic", "Trial")
	return cmd
}

//NewCmdGsshDelete ucloud gssh delete
func NewCmdGsshDelete() *cobra.Command {
	var req = BizClient.NewDeleteGlobalSSHInstanceRequest()
	var gsshIds *[]string
	var cmd = &cobra.Command{
		Use:     "delete",
		Short:   "Delete GlobalSSH instance",
		Long:    "Delete GlobalSSH instance",
		Example: "ucloud gssh delete --resource-id uga-xx1  --id uga-xx2",
		Run: func(cmd *cobra.Command, args []string) {
			for _, id := range *gsshIds {
				req.InstanceId = &id
				_, err := BizClient.DeleteGlobalSSHInstance(req)
				if err != nil {
					HandleError(err)
				} else {
					Cxt.Printf("gssh[%s] deleted\n", id)
				}
			}
		},
	}
	cmd.Flags().SortFlags = false
	gsshIds = cmd.Flags().StringArray("resource-id", make([]string, 0), "Required. ID of the GlobalSSH instances you want to delete. Multiple values specified by multiple flags")
	req.ProjectId = cmd.Flags().String("project-id", ConfigInstance.ProjectID, "Optional. Assign project-id")
	req.Region = cmd.Flags().String("region", ConfigInstance.Region, "Optional. Assign region")
	cmd.MarkFlagRequired("resource-id")
	return cmd
}

//NewCmdGsshModify ucloud gssh modify
func NewCmdGsshModify() *cobra.Command {
	var gsshModifyPortReq = BizClient.NewModifyGlobalSSHPortRequest()
	var gsshModifyRemarkReq = BizClient.NewModifyGlobalSSHRemarkRequest()
	region := ConfigInstance.Region
	project := ConfigInstance.ProjectID
	var cmd = &cobra.Command{
		Use:     "update",
		Short:   "Update GlobalSSH instance",
		Long:    "Update GlobalSSH instance, including port and remark attribute",
		Example: "ucloud gssh update --resource-id uga-xxx --port 22",
		Run: func(cmd *cobra.Command, args []string) {
			gsshModifyPortReq.Region = sdk.String(region)
			gsshModifyPortReq.ProjectId = sdk.String(project)
			gsshModifyRemarkReq.Region = sdk.String(region)
			gsshModifyRemarkReq.ProjectId = sdk.String(project)
			if *gsshModifyPortReq.Port == 0 && *gsshModifyRemarkReq.Remark == "" {
				Cxt.Println("port or remark required")
			}
			if *gsshModifyPortReq.Port != 0 {
				port := *gsshModifyPortReq.Port
				if port <= 1 || port >= 65535 || port == 80 || port == 443 {
					Cxt.Println("The port number should be between 1 and 65535, and cannot be equal to 80 or 443")
					return
				}
				gsshModifyPortReq.InstanceId = gsshModifyRemarkReq.InstanceId
				_, err := BizClient.ModifyGlobalSSHPort(gsshModifyPortReq)
				if err != nil {
					HandleError(err)
				} else {
					Cxt.Printf("gssh[%s] updated\n", *gsshModifyPortReq.InstanceId)
				}
			}
			if *gsshModifyRemarkReq.Remark != "" {
				_, err := BizClient.ModifyGlobalSSHRemark(gsshModifyRemarkReq)
				if err != nil {
					HandleError(err)
				} else {
					Cxt.Printf("gssh[%s] updated\n", *gsshModifyRemarkReq.InstanceId)
				}
			}
		},
	}
	cmd.Flags().SortFlags = false
	gsshModifyRemarkReq.InstanceId = cmd.Flags().String("resource-id", "", "Required. InstanceID of your GlobalSSH")
	cmd.Flags().StringVar(&region, "region", ConfigInstance.Region, "Optional. Assign region")
	cmd.Flags().StringVar(&project, "project-id", ConfigInstance.ProjectID, "Optional. Assign project-id")
	gsshModifyPortReq.Port = cmd.Flags().Int("port", 0, "Optional. Port of SSH service.")
	gsshModifyRemarkReq.Remark = cmd.Flags().String("remark", "", "Optional. Remark of your GlobalSSH.")
	cmd.MarkFlagRequired("resource-id")
	return cmd
}
