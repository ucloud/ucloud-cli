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
	"strings"

	"github.com/spf13/cobra"

	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/base"
)

//NewCmdSharedBW ucloud shared-bw
func NewCmdSharedBW() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "shared-bw",
		Short: "Create and manipulate shared bandwidth instances",
		Long:  "Create and manipulate shared bandwidth instances",
	}
	cmd.AddCommand(NewCmdSharedBWCreate())
	cmd.AddCommand(NewCmdSharedBWList())
	cmd.AddCommand(NewCmdSharedBWResize())
	cmd.AddCommand(NewCmdSharedBWDelete())
	return cmd
}

//NewCmdSharedBWCreate ucloud shared-bw create
func NewCmdSharedBWCreate() *cobra.Command {
	req := base.BizClient.NewAllocateShareBandwidthRequest()
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create shared bandwidth instance",
		Long:  "Create shared bandwidth instance",
		Run: func(c *cobra.Command, args []string) {
			if *req.ShareBandwidth < 20 || *req.ShareBandwidth > 5000 {
				base.Cxt.Printf("bandwidth should be between 20 and 5000. received %d\n", *req.ShareBandwidth)
				return
			}
			resp, err := base.BizClient.AllocateShareBandwidth(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			base.Cxt.Printf("shared bandwidth[%s] created\n", resp.ShareBandwidthId)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.Name = flags.String("name", "", "Required. Name of the shared bandwidth instance")
	req.ShareBandwidth = flags.Int("bandwidth-mb", 20, "Optional. Unit:Mb. Bandwidth of the shared bandwidth. Range [20,5000]")
	req.Region = flags.String("region", base.ConfigInstance.Region, "Optional. Region, see 'ucloud region'")
	req.ProjectId = flags.String("project-id", base.ConfigInstance.ProjectID, "Optional. Project-id, see 'ucloud project list'")
	req.ChargeType = flags.String("charge-type", "Month", "Optional.'Year',pay yearly;'Month',pay monthly;'Dynamic', pay hourly")
	req.Quantity = flags.Int("quantity", 1, "Optional. The duration of the instance. N years/months.")
	flags.SetFlagValues("charge-type", "Month", "Year", "Dynamic")
	cmd.MarkFlagRequired("name")
	return cmd
}

//SharedBWRow 表格行
type SharedBWRow struct {
	Name           string
	ResourceID     string
	ChargeType     string
	Bandwidth      string
	EIP            string
	ExpirationTime string
}

//NewCmdSharedBWList ucloud shared-bw list
func NewCmdSharedBWList() *cobra.Command {
	req := base.BizClient.NewDescribeShareBandwidthRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List shared bandwidth instances",
		Long:  "List shared bandwidth instances",
		Run: func(c *cobra.Command, args []string) {
			resp, err := base.BizClient.DescribeShareBandwidth(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			list := []SharedBWRow{}
			for _, sb := range resp.DataSet {
				row := SharedBWRow{}
				row.Name = sb.Name
				row.ResourceID = sb.ShareBandwidthId
				row.ChargeType = sb.ChargeType
				row.Bandwidth = strconv.Itoa(sb.ShareBandwidth) + "Mb"
				row.ExpirationTime = base.FormatDate(sb.ExpireTime)
				eipList := []string{}
				for _, eip := range sb.EIPSet {
					eipText := ""
					eipText += eip.EIPId
					for _, ip := range eip.EIPAddr {
						eipText += fmt.Sprintf("/%s/%s", ip.IP, ip.OperatorName)
					}
					eipList = append(eipList, eipText)
				}
				row.EIP = strings.Join(eipList, "\n")
				list = append(list, row)
			}
			if global.json {
				base.PrintJSON(list)
			} else {
				base.PrintTableS(list)
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.Region = flags.String("region", base.ConfigInstance.Region, "Optional. Region, see 'ucloud region'")
	req.ProjectId = flags.String("project-id", base.ConfigInstance.ProjectID, "Optional. Project-id, see 'ucloud project list'")
	flags.StringSliceVar(&req.ShareBandwidthIds, "resource-id", nil, "Resource ID of shared bandwidth instances to list")

	return cmd
}

//NewCmdSharedBWResize ucloud shared-bw resize
func NewCmdSharedBWResize() *cobra.Command {
	req := base.BizClient.NewResizeShareBandwidthRequest()
	cmd := &cobra.Command{
		Use:   "resize",
		Short: "Resize shared bandwidth instance's bandwidth",
		Long:  "Resize shared bandwidth instance's bandwidth",
		Run: func(c *cobra.Command, args []string) {
			if *req.ShareBandwidth < 20 || *req.ShareBandwidth > 5000 {
				base.Cxt.Printf("bandwidth should be between 20 and 5000. received %d\n", *req.ShareBandwidth)
				return
			}
			req.ShareBandwidthId = sdk.String(base.PickResourceID(*req.ShareBandwidthId))
			_, err := base.BizClient.ResizeShareBandwidth(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			base.Cxt.Printf("shared bandwidth[%s] resized to %dMb\n", *req.ShareBandwidthId, *req.ShareBandwidth)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.ShareBandwidthId = flags.String("resource-id", "", "Required. Resource ID of shared bandwidth instance to resize")
	req.ShareBandwidth = flags.Int("bandwidth-mb", 0, "Required. Unit:Mb. resize to bandwidth value")
	req.Region = flags.String("region", base.ConfigInstance.Region, "Optional. Region, see 'ucloud region'")
	req.ProjectId = flags.String("project-id", base.ConfigInstance.ProjectID, "Optional. Project-id, see 'ucloud project list'")
	flags.SetFlagValuesFunc("resource-id", func() []string {
		list, _ := getAllSharedBW(*req.ProjectId, *req.Region)
		return list
	})
	return cmd
}

//NewCmdSharedBWDelete ucloud shared-bw delete
func NewCmdSharedBWDelete() *cobra.Command {
	req := base.BizClient.NewReleaseShareBandwidthRequest()
	ids := []string{}
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete shared bandwidth instance",
		Long:  "Delete shared bandwidth instance",
		Run: func(c *cobra.Command, args []string) {
			for _, id := range ids {
				req.ShareBandwidthId = sdk.String(base.PickResourceID(id))
				_, err := base.BizClient.ReleaseShareBandwidth(req)
				if err != nil {
					base.HandleError(err)
					continue
				}
				base.Cxt.Printf("shared bandwidth[%s] deleted\n", id)
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	flags.StringSliceVar(&ids, "resource-id", nil, "Required. Resource ID of shared bandwidth instances to delete")
	req.EIPBandwidth = flags.Int("eip-bandwidth-mb", 1, "Optional. Bandwidth of the joined EIPs,after deleting the shared bandwidth instance")
	req.PayMode = flags.String("charge-mode", "", "Optional. Charge mode of joined EIPs,after deleting the shared bandwidth")
	req.Region = flags.String("region", base.ConfigInstance.Region, "Optional. Region, see 'ucloud region'")
	req.ProjectId = flags.String("project-id", base.ConfigInstance.ProjectID, "Optional. Project-id, see 'ucloud project list'")
	flags.SetFlagValuesFunc("resource-id", func() []string {
		list, _ := getAllSharedBW(*req.ProjectId, *req.Region)
		return list
	})
	flags.SetFlagValues("charge-mode", "Bandwidth", "Traffic")
	cmd.MarkFlagRequired("resource-id")

	return cmd
}

func getAllSharedBW(project, region string) ([]string, error) {
	req := base.BizClient.NewDescribeShareBandwidthRequest()
	req.ProjectId = &project
	req.Region = &region
	resp, err := base.BizClient.DescribeShareBandwidth(req)
	if err != nil {
		return nil, err
	}
	list := []string{}
	for _, item := range resp.DataSet {
		list = append(list, item.ShareBandwidthId+"/"+item.Name)
	}
	return list, nil
}
