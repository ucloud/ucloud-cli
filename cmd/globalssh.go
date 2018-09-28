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
	"github.com/spf13/cobra"
	. "github.com/ucloud/ucloud-cli/util"
)

//NewCmdGssh ucloud gssh
func NewCmdGssh() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "gssh",
		Short: "Create, list, update and delete globalssh instance",
		Long:  `Create, list, update and delete globalssh instance`,
	}
	cmd.AddCommand(NewCmdGsshList())
	cmd.AddCommand(NewCmdGsshCreate())
	cmd.AddCommand(NewCmdGsshDelete())
	cmd.AddCommand(NewCmdGsshModify())
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
	var cmd = &cobra.Command{
		Use:     "list",
		Short:   "List all GlobalSSH instances",
		Long:    `List all GlobalSSH instances`,
		Example: "ucloud gssh ls",
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := BizClient.DescribeGlobalSSHInstance(req)
			if err != nil {
				Cxt.PrintErr(err)
			} else {
				if resp.RetCode == 0 {
					if global.json {
						PrintJSON(resp.InstanceSet)
					} else {
						list := make([]GSSHRow, 0)
						for _, gssh := range resp.InstanceSet {
							row := GSSHRow{}
							row.ResourceID = gssh.InstanceId
							row.SSHServerIP = gssh.TargetIP
							row.AcceleratingDomain = gssh.AcceleratingDomain
							row.SSHServerLocation = gssh.Area
							row.SSHPort = gssh.Port
							row.Remark = gssh.Remark
							list = append(list, row)
						}
						PrintTable(list, []string{"ResourceID", "SSHServerIP", "AcceleratingDomain", "SSHServerLocation", "SSHPort", "Remark"})
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

//NewCmdGsshCreate ucloud gssh create
func NewCmdGsshCreate() *cobra.Command {
	var req = BizClient.NewCreateGlobalSSHInstanceRequest()
	var cmd = &cobra.Command{
		Use:     "create",
		Short:   "Create GlobalSSH instance",
		Long:    "Create GlobalSSH instance",
		Example: "ucloud gssh create --area Washington --target-ip 8.8.8.8",
		Run: func(cmd *cobra.Command, args []string) {
			var areaMap = map[string]string{
				"LosAngeles": "洛杉矶",
				"Singapore":  "新加坡",
				"HongKong":   "香港",
				"Tokyo":      "东京",
				"Washington": "华盛顿",
				"Frankfurt":  "法兰克福",
			}

			port := *req.Port
			if port < 1 || port > 65535 || port == 80 || port == 443 {
				Cxt.Println("The port number should be between 1 and 65535, and cannot be equal to 80 or 443")
				return
			}

			if area, ok := areaMap[*req.Area]; ok {
				*req.Area = area
			} else {
				Cxt.Println("Area should be one of LosAngeles,Singapore,HongKong,Tokyo,Washington,Frankfurt.")
				return
			}
			resp, err := BizClient.CreateGlobalSSHInstance(req)
			if err != nil {
				Cxt.PrintErr(err)
			} else {
				if resp.RetCode == 0 {
					Cxt.Println("Succeed, GlobalSSHInstanceId:", resp.InstanceId)
				} else {
					HandleBizError(resp)
				}
			}
		},
	}
	cmd.Flags().SortFlags = false
	req.Region = cmd.Flags().String("region", ConfigInstance.Region, "Assign region(override default region of your config)")
	req.ProjectId = cmd.Flags().String("project-id", ConfigInstance.ProjectID, "Assign project-id(override default projec-id of your config)")
	req.Area = cmd.Flags().String("area", "", "Location of the source server.Only supports six cities,LosAngeles,Singapore,HongKong,Tokyo,Washington,Frankfurt. Required")
	req.TargetIP = cmd.Flags().String("target-ip", "", "IP of the source server. Required")
	req.Port = cmd.Flags().Int("port", 22, "Port of The SSH service between 1 and 65535. Do not use ports such as 80,443.")
	req.Remark = cmd.Flags().String("remark", "", "Remark of your GlobalSSH.")
	req.CouponId = cmd.Flags().String("coupon-id", "", "Coupon ID, The Coupon can deduct part of the payment,see DescribeCoupon or https://accountv2.ucloud.cn")
	cmd.MarkFlagRequired("area")
	cmd.MarkFlagRequired("target-ip")
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
		Example: "ucloud gssh delete --id uga-xx1  --id uga-xx2",
		Run: func(cmd *cobra.Command, args []string) {
			for _, id := range *gsshIds {
				req.InstanceId = &id
				resp, err := BizClient.DeleteGlobalSSHInstance(req)
				if err != nil {
					Cxt.PrintErr(err)
				} else {
					if resp.RetCode == 0 {
						Cxt.Printf("GlobalSSH(%s) was successfully deleted\n", id)
					} else {
						HandleBizError(resp)
					}
				}
			}
		},
	}
	req.Region = cmd.Flags().String("region", ConfigInstance.Region, "Assign region(override default region of your config)")
	req.ProjectId = cmd.Flags().String("project-id", ConfigInstance.ProjectID, "Assign project-id(override default projec-id of your config)")
	gsshIds = cmd.Flags().StringArray("id", make([]string, 0), "ID of the GlobalSSH instances you want to delete. Multiple values specified by multiple flags. Required")
	cmd.MarkFlagRequired("id")
	return cmd
}

//NewCmdGsshModify ucloud gssh modify
func NewCmdGsshModify() *cobra.Command {
	var gsshModifyPortReq = BizClient.NewModifyGlobalSSHPortRequest()
	var gsshModifyRemarkReq = BizClient.NewModifyGlobalSSHRemarkRequest()
	var region, project string
	var cmd = &cobra.Command{
		Use:     "update",
		Short:   "Update GlobalSSH instance",
		Long:    "Update GlobalSSH instance, including port and remark attribute",
		Example: "ucloud gssh update --id uga-xxx --port 22",
		Run: func(cmd *cobra.Command, args []string) {
			*gsshModifyPortReq.Region = region
			*gsshModifyPortReq.ProjectId = project
			*gsshModifyRemarkReq.Region = region
			*gsshModifyRemarkReq.ProjectId = project

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
				resp, err := BizClient.ModifyGlobalSSHPort(gsshModifyPortReq)
				if err != nil {
					Cxt.Println("Error:", err)
				} else {
					if resp.RetCode == 0 {
						Cxt.Println("Successfully updated")
					} else {
						Cxt.Printf("Something wrong. RetCode:%d, Message: %s\n", resp.RetCode, resp.Message)
					}
				}
			}
			if *gsshModifyRemarkReq.Remark != "" {
				resp, err := BizClient.ModifyGlobalSSHRemark(gsshModifyRemarkReq)
				if err != nil {
					Cxt.Println(err)
				} else {
					if resp.RetCode == 0 {
						Cxt.Println("Successfully updated")
					} else {
						HandleBizError(resp)
					}
				}
			}
		},
	}
	cmd.Flags().StringVar(&region, "region", ConfigInstance.Region, "Assign region(override default region of your config)")
	cmd.Flags().StringVar(&project, "project-id", ConfigInstance.ProjectID, "Assign project-id(override default projec-id of your config)")
	gsshModifyPortReq.Port = cmd.Flags().Int("port", 0, "Port of SSH service.")
	gsshModifyRemarkReq.Remark = cmd.Flags().String("remark", "", "Remark of your GlobalSSH.")
	gsshModifyRemarkReq.InstanceId = cmd.Flags().String("id", "", "InstanceID of your GlobalSSH. Required")
	cmd.MarkFlagRequired("id")
	return cmd
}
