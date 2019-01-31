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
	"time"

	"github.com/spf13/cobra"

	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/base"
	"github.com/ucloud/ucloud-cli/model/status"
)

//NewCmdBandwidth ucloud bw
func NewCmdBandwidth() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bw",
		Short: "Manipulate bandwidth package and shared bandwidth",
		Long:  "Manipulate bandwidth package and shared bandwidth",
	}
	cmd.AddCommand(NewCmdBandwidthPkg())
	cmd.AddCommand(NewCmdSharedBW())
	return cmd
}

//NewCmdSharedBW ucloud shared-bw
func NewCmdSharedBW() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "shared",
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
	req.Region = flags.String("region", base.ConfigIns.Region, "Optional. Region, see 'ucloud region'")
	req.ProjectId = flags.String("project-id", base.ConfigIns.ProjectID, "Optional. Project-id, see 'ucloud project list'")
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

	req.Region = flags.String("region", base.ConfigIns.Region, "Optional. Region, see 'ucloud region'")
	req.ProjectId = flags.String("project-id", base.ConfigIns.ProjectID, "Optional. Project-id, see 'ucloud project list'")
	flags.StringSliceVar(&req.ShareBandwidthIds, "shared-bw-id", nil, "Resource ID of shared bandwidth instances to list")

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

	req.ShareBandwidthId = flags.String("shared-bw-id", "", "Required. Resource ID of shared bandwidth instance to resize")
	req.ShareBandwidth = flags.Int("bandwidth-mb", 0, "Required. Unit:Mb. resize to bandwidth value")
	req.Region = flags.String("region", base.ConfigIns.Region, "Optional. Region, see 'ucloud region'")
	req.ProjectId = flags.String("project-id", base.ConfigIns.ProjectID, "Optional. Project-id, see 'ucloud project list'")

	flags.SetFlagValuesFunc("shared-bw-id", func() []string {
		list, _ := getAllSharedBW(*req.ProjectId, *req.Region)
		return list
	})

	cmd.MarkFlagRequired("shared-bw-id")
	cmd.MarkFlagRequired("bandwidth-mb")

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

	flags.StringSliceVar(&ids, "shared-bw-id", nil, "Required. Resource ID of shared bandwidth instances to delete")
	req.EIPBandwidth = flags.Int("eip-bandwidth-mb", 1, "Optional. Bandwidth of the joined EIPs,after deleting the shared bandwidth instance")
	req.PayMode = flags.String("traffic-mode", "", "Optional. The charge mode of joined EIPs after deleting the shared bandwidth. Accept values:Bandwidth,Traffic")
	req.Region = flags.String("region", base.ConfigIns.Region, "Optional. Region, see 'ucloud region'")
	req.ProjectId = flags.String("project-id", base.ConfigIns.ProjectID, "Optional. Project-id, see 'ucloud project list'")
	flags.SetFlagValuesFunc("shared-bw-id", func() []string {
		list, _ := getAllSharedBW(*req.ProjectId, *req.Region)
		return list
	})
	flags.SetFlagValues("traffic-mode", "Bandwidth", "Traffic")

	cmd.MarkFlagRequired("shared-bw-id")

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

//NewCmdBandwidthPkg ucloud bw-pkg
func NewCmdBandwidthPkg() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pkg",
		Short: "List, create and delete bandwidth package instances",
		Long:  "List, create and delete bandwidth package instances",
	}
	cmd.AddCommand(NewCmdBandwidthPkgCreate())
	cmd.AddCommand(NewCmdBandwidthPkgList())
	cmd.AddCommand(NewCmdBandwidthPkgDelete())
	return cmd
}

//NewCmdBandwidthPkgCreate ucloud bw-pkg create
func NewCmdBandwidthPkgCreate() *cobra.Command {
	var start, end *string
	timeLayout := "2006-01-02/15:04:05"
	ids := []string{}
	req := base.BizClient.NewCreateBandwidthPackageRequest()
	loc, _ := time.LoadLocation("Local")
	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Create bandwidth package",
		Long:    "Create bandwidth package",
		Example: "ucloud bw pkg create --eip-id eip-xxx --bandwidth-mb 20 --start-time 2018-12-15/09:20:00 --end-time 2018-12-16/09:20:00",
		Run: func(c *cobra.Command, args []string) {
			st, err := time.ParseInLocation(timeLayout, *start, loc)
			if err != nil {
				base.HandleError(err)
				return
			}
			et, err := time.ParseInLocation(timeLayout, *end, loc)
			if err != nil {
				base.HandleError(err)
				return
			}
			if st.Sub(time.Now()) < 0 {
				base.Cxt.Println("start-time must be after the current time")
				return
			}
			du := et.Unix() - st.Unix()
			if du <= 0 {
				base.Cxt.Println("end-time must be after the start-time")
				return
			}
			req.EnableTime = sdk.Int(int(st.Unix()))
			req.TimeRange = sdk.Int(int(du))

			for _, id := range ids {
				id = base.PickResourceID(id)
				req.EIPId = &id
				resp, err := base.BizClient.CreateBandwidthPackage(req)
				if err != nil {
					base.HandleError(err)
					continue
				}
				base.Cxt.Printf("bandwidth package[%s] created for eip[%s]\n", resp.BandwidthPackageId, id)
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	flags.StringSliceVar(&ids, "eip-id", nil, "Required. Resource ID of eip to be bound with created bandwidth package")
	start = flags.String("start-time", "", "Required. The time to enable bandwidth package. Local time, for example '2018-12-25/08:30:00'")
	end = flags.String("end-time", "", "Required. The time to disable bandwidth package. Local time, for example '2018-12-26/08:30:00'")
	req.Bandwidth = flags.Int("bandwidth-mb", 0, "Required. bandwidth of the bandwidth package to create.Range [1,800]. Unit:'Mb'.")
	req.Region = flags.String("region", base.ConfigIns.Region, "Optional. Region, see 'ucloud region'")
	req.ProjectId = flags.String("project-id", base.ConfigIns.ProjectID, "Optional. Project-id, see 'ucloud project list'")

	cmd.Flags().SetFlagValuesFunc("eip-id", func() []string {
		return getAllEip(*req.ProjectId, *req.Region, []string{status.EIP_USED}, []string{status.EIP_CHARGE_BANDWIDTH})
	})

	cmd.MarkFlagRequired("eip-id")
	cmd.MarkFlagRequired("start-time")
	cmd.MarkFlagRequired("end-time")
	cmd.MarkFlagRequired("bandwidth-mb")
	return cmd
}

//BandwidthPkgRow 表格行
type BandwidthPkgRow struct {
	ResourceID string
	EIP        string
	Bandwidth  string
	StartTime  string
	EndTime    string
}

//NewCmdBandwidthPkgList ucloud bw-pkg list
func NewCmdBandwidthPkgList() *cobra.Command {
	req := base.BizClient.NewDescribeBandwidthPackageRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List bandwidth packages",
		Long:  "List bandwidth packages",
		Run: func(c *cobra.Command, args []string) {
			resp, err := base.BizClient.DescribeBandwidthPackage(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			list := []BandwidthPkgRow{}
			for _, bp := range resp.DataSets {
				row := BandwidthPkgRow{
					ResourceID: bp.BandwidthPackageId,
					Bandwidth:  strconv.Itoa(bp.Bandwidth) + "MB",
					StartTime:  base.FormatDateTime(bp.EnableTime),
					EndTime:    base.FormatDateTime(bp.DisableTime),
				}
				eip := bp.EIPId
				for _, addr := range bp.EIPAddr {
					eip += "/" + addr.IP + "/" + addr.OperatorName
				}
				row.EIP = eip
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
	req.Region = flags.String("region", base.ConfigIns.Region, "Optional. Region, see 'ucloud region'")
	req.ProjectId = flags.String("project-id", base.ConfigIns.ProjectID, "Optional. Project-id, see 'ucloud project list'")
	req.Offset = cmd.Flags().Int("offset", 0, "Optional. Offset")
	req.Limit = cmd.Flags().Int("limit", 50, "Optional. Limit range [0,10000000]")

	return cmd
}

//NewCmdBandwidthPkgDelete ucloud bw-pkg delete
func NewCmdBandwidthPkgDelete() *cobra.Command {
	ids := []string{}
	req := base.BizClient.NewDeleteBandwidthPackageRequest()
	cmd := &cobra.Command{
		Use:     "delete",
		Short:   "Delete bandwidth packages",
		Long:    "Delete bandwidth packages",
		Example: "ucloud bw pkg delete --resource-id bwpack-xxx",
		Run: func(c *cobra.Command, args []string) {
			for _, id := range ids {
				id := base.PickResourceID(id)
				req.BandwidthPackageId = &id
				_, err := base.BizClient.DeleteBandwidthPackage(req)
				if err != nil {
					base.HandleError(err)
					return
				}
				base.Cxt.Printf("bandwidth package[%s] deleted\n", id)
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	flags.StringSliceVar(&ids, "resource-id", nil, "Required, Resource ID of bandwidth package to delete")
	req.Region = flags.String("region", base.ConfigIns.Region, "Optional. Region, see 'ucloud region'")
	req.ProjectId = flags.String("project-id", base.ConfigIns.ProjectID, "Optional. Project-id, see 'ucloud project list'")

	return cmd
}
