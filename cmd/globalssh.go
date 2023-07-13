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
	"io"
	"net"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/pathx"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/base"
)

// NewCmdGssh ucloud gssh
func NewCmdGssh() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gssh",
		Short: "Create,list,update and delete globalssh instance",
		Long:  `Create,list,update and delete globalssh instance`,
	}
	out := base.Cxt.GetWriter()
	cmd.AddCommand(NewCmdGsshList(out))
	cmd.AddCommand(NewCmdGsshCreate())
	cmd.AddCommand(NewCmdGsshDelete())
	cmd.AddCommand(NewCmdGsshModify())
	cmd.AddCommand(NewCmdGsshArea())
	return cmd
}

// GSSHRow gssh表格行
type GSSHRow struct {
	ResourceID         string
	SSHServerIP        string
	AcceleratingDomain string
	SSHServerLocation  string
	SSHPort            int
	GlobalSSHPort      int
	Remark             string
	InstanceType       string
}

// NewCmdGsshList ucloud gssh list
func NewCmdGsshList(out io.Writer) *cobra.Command {
	req := base.BizClient.NewDescribeGlobalSSHInstanceRequest()
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
				"拉各斯":  "Lagos",
			}

			resp, err := base.BizClient.DescribeGlobalSSHInstance(req)
			if err != nil {
				base.HandleError(err)
			} else {
				list := make([]GSSHRow, 0)
				for _, gssh := range resp.InstanceSet {
					row := GSSHRow{}
					row.ResourceID = gssh.InstanceId
					row.SSHServerIP = gssh.TargetIP
					row.AcceleratingDomain = gssh.AcceleratingDomain
					row.SSHPort = gssh.Port
					row.GlobalSSHPort = gssh.GlobalSSHPort
					row.Remark = gssh.Remark
					row.InstanceType = gssh.InstanceType
					if val, ok := areaMap[gssh.Area]; ok {
						row.SSHServerLocation = val
					} else {
						row.SSHServerLocation = gssh.Area
					}
					list = append(list, row)
				}
				base.PrintList(list, out)
			}
		},
	}
	cmd.Flags().SortFlags = false
	req.Region = cmd.Flags().String("region", base.ConfigIns.Region, "Optional. Assign region")
	req.ProjectId = cmd.Flags().String("project-id", base.ConfigIns.ProjectID, "Optional. Assign project-id")
	return cmd
}

// NewCmdGsshArea ucloud gssh area
func NewCmdGsshArea() *cobra.Command {
	req := base.BizClient.NewDescribeGlobalSSHAreaRequest()
	cmd := &cobra.Command{
		Use:   "location",
		Short: "List SSH server locations and covered areas",
		Long:  "List SSH server locations and covered areas",
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := base.BizClient.DescribeGlobalSSHArea(req)
			if err != nil {
				base.HandleError(err)
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
					regionLabels = append(regionLabels, base.RegionLabel[region])
				}
				row.CoveredArea = strings.Join(regionLabels, ",")
				list = append(list, row)
			}

			base.PrintTable(list, []string{"AirportCode", "SSHServerLocation", "CoveredArea"})
		},
	}
	return cmd
}

// GsshLocation 服务地点和覆盖区域
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
	"LOS": "Lagos",
}

// NewCmdGsshCreate ucloud gssh create
func NewCmdGsshCreate() *cobra.Command {
	var targetIP *net.IP
	req := base.BizClient.NewCreateGlobalSSHInstanceRequest()
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
			if port < 1 || port > 65535 || port == 80 || port == 443 || port == 65123 {
				base.Cxt.Println("The port number should be between 1 and 65535, and cannot be 80, 443 or 65123")
				return
			}
			req.TargetIP = sdk.String(targetIP.String())
			resp, err := base.BizClient.CreateGlobalSSHInstance(req)
			if err != nil {
				base.HandleError(err)
			} else {
				base.Cxt.Printf("gssh[%s] created\n", resp.InstanceId)
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	req.AreaCode = cmd.Flags().String("location", "", "Required. Location of the source server. See 'ucloud gssh location'")
	targetIP = cmd.Flags().IP("target-ip", nil, "Required. IP of the source server. Required")
	bindProjectID(req, flags)
	req.Port = cmd.Flags().Int("port", 22, "Optional. Port of The SSH service between 1 and 65535. Do not use ports such as 80, 443 or 65123.")
	req.Remark = cmd.Flags().String("remark", "", "Optional. Remark of your GlobalSSH.")
	req.ChargeType = cmd.Flags().String("charge-type", "Month", "Optional.'Year',pay yearly;'Month',pay monthly;'Dynamic', pay hourly(requires access)")
	req.Quantity = cmd.Flags().Int("quantity", 1, "Optional. The duration of the instance. N years/months.")
	req.InstanceType = cmd.Flags().String("instance-type", "", "Optional. Possible values: 'Ultimate','Enterprise', 'Basic', 'Free'(Default value)")
	req.ForwardRegion = cmd.Flags().String("forward-region", "", "Optional. You can select one of 'cn-bj2','cn-sh2','cn-gd' When instance-type is 'Basic'")
	req.BandwidthPackage = cmd.Flags().Int("bandwidth-package", 0, "Optional. You can set one of 0, 20, 40 When instance-type is 'Ultimate'")
	cmd.MarkFlagRequired("location")
	cmd.MarkFlagRequired("target-ip")
	cmd.Flags().SetFlagValues("location", "LosAngeles", "Singapore", "Lagos", "HongKong", "Tokyo", "Washington", "Frankfurt")
	cmd.Flags().SetFlagValues("charge-type", "Month", "Year", "Dynamic", "Trial")
	cmd.Flags().SetFlagValues("bandwidth-package", "0", "20", "40")
	cmd.Flags().SetFlagValues("forward-region", "cn-bj2", "cn-sh2", "cn-gd")
	cmd.Flags().SetFlagValues("instance-type", "Free", "Basic", "Enterprise", "Ultimate")
	cmd.Flags().SetFlagValuesFunc("target-ip", func() []string {
		eips := getAllEip(*req.ProjectId, base.ConfigIns.Region, nil, nil)
		for idx, eip := range eips {
			eips[idx] = strings.SplitN(eip, "/", 2)[1]
		}
		return eips
	})
	return cmd
}

// NewCmdGsshDelete ucloud gssh delete
func NewCmdGsshDelete() *cobra.Command {
	var req = base.BizClient.NewDeleteGlobalSSHInstanceRequest()
	var gsshIds *[]string
	var cmd = &cobra.Command{
		Use:     "delete",
		Short:   "Delete GlobalSSH instance",
		Long:    "Delete GlobalSSH instance",
		Example: "ucloud gssh delete --gssh-id uga-xx1  --id uga-xx2",
		Run: func(cmd *cobra.Command, args []string) {
			req.ProjectId = sdk.String(base.PickResourceID(*req.ProjectId))
			for _, id := range *gsshIds {
				req.InstanceId = sdk.String(base.PickResourceID(id))
				_, err := base.BizClient.DeleteGlobalSSHInstance(req)
				if err != nil {
					base.HandleError(err)
				} else {
					base.Cxt.Printf("gssh[%s] deleted\n", id)
				}
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	gsshIds = cmd.Flags().StringSlice("gssh-id", make([]string, 0), "Required. ID of the GlobalSSH instances you want to delete. Multiple values specified by multiple commas")
	bindProjectID(req, flags)
	cmd.MarkFlagRequired("gssh-id")
	cmd.Flags().SetFlagValuesFunc("gssh-id", func() []string {
		return getAllGsshIDNames(*req.ProjectId)
	})
	return cmd
}

// NewCmdGsshModify ucloud gssh modify
func NewCmdGsshModify() *cobra.Command {
	gsshModifyPortReq := base.BizClient.NewModifyGlobalSSHPortRequest()
	gsshModifyRemarkReq := base.BizClient.NewModifyGlobalSSHRemarkRequest()
	project := base.ConfigIns.ProjectID
	gsshIDs := []string{}
	cmd := &cobra.Command{
		Use:     "update",
		Short:   "Update GlobalSSH instance",
		Long:    "Update GlobalSSH instance, including port and remark attribute",
		Example: "ucloud gssh update --gssh-id uga-xxx --port 22",
		Run: func(cmd *cobra.Command, args []string) {
			gsshModifyPortReq.ProjectId = sdk.String(project)
			gsshModifyRemarkReq.ProjectId = sdk.String(project)
			if *gsshModifyPortReq.Port == 0 && *gsshModifyRemarkReq.Remark == "" {
				base.Cxt.Println("Error, port or remark required")
			}
			if *gsshModifyPortReq.Port != 0 {
				port := *gsshModifyPortReq.Port
				if port <= 1 || port >= 65535 || port == 80 || port == 443 || port == 65123 {
					base.Cxt.Println("The port number should be between 1 and 65535, and cannot be equal to 80, 443 or 65123")
					return
				}
				for _, idname := range gsshIDs {
					gsshModifyPortReq.InstanceId = sdk.String(base.PickResourceID(idname))
					_, err := base.BizClient.ModifyGlobalSSHPort(gsshModifyPortReq)
					if err != nil {
						base.HandleError(err)
					} else {
						base.Cxt.Printf("gssh[%s]'s port updated\n", *gsshModifyPortReq.InstanceId)
					}
				}
			}
			if *gsshModifyRemarkReq.Remark != "" {
				for _, idname := range gsshIDs {
					gsshModifyRemarkReq.InstanceId = sdk.String(base.PickResourceID(idname))
					_, err := base.BizClient.ModifyGlobalSSHRemark(gsshModifyRemarkReq)
					if err != nil {
						base.HandleError(err)
					} else {
						base.Cxt.Printf("gssh[%s]'s remark updated\n", *gsshModifyRemarkReq.InstanceId)
					}
				}
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&gsshIDs, "gssh-id", nil, "Required. ResourceID of your GlobalSSH instances")
	bindProjectIDS(&project, flags)
	gsshModifyPortReq.Port = cmd.Flags().Int("port", 0, "Optional. Port of SSH service.")
	gsshModifyRemarkReq.Remark = cmd.Flags().String("remark", "", "Optional. Remark of your GlobalSSH.")
	cmd.MarkFlagRequired("gssh-id")
	cmd.Flags().SetFlagValuesFunc("gssh-id", func() []string {
		return getAllGsshIDNames(project)
	})
	return cmd
}

func getAllGssh(project string) ([]pathx.GlobalSSHInfo, error) {
	req := base.BizClient.NewDescribeGlobalSSHInstanceRequest()
	req.ProjectId = &project
	resp, err := base.BizClient.DescribeGlobalSSHInstance(req)
	if err != nil {
		return nil, err
	}
	return resp.InstanceSet, nil
}

func getAllGsshIDNames(project string) []string {
	gsshs, err := getAllGssh(project)
	if err != nil {
		return nil
	}
	list := []string{}
	for _, gssh := range gsshs {
		list = append(list, fmt.Sprintf("%s/%s", gssh.InstanceId, gssh.TargetIP))
	}
	return list
}
