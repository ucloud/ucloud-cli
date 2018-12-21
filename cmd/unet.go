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

	"github.com/spf13/cobra"
	"github.com/ucloud/ucloud-cli/base"
	"github.com/ucloud/ucloud-cli/model/status"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"
)

//NewCmdSubnet  ucloud subnet
func NewCmdSubnet() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "subnet",
		Short: "List subnet",
		Long:  `List subnet`,
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(NewCmdSubnetList())

	return cmd
}

//SubnetRow 表格行
type SubnetRow struct {
	SubnetName     string
	ResourceID     string
	Group          string
	AffiliatedVPC  string
	NetworkSegment string
	CreationTime   string
}

//NewCmdSubnetList ucloud subnet list
func NewCmdSubnetList() *cobra.Command {
	req := base.BizClient.NewDescribeSubnetRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List subnet",
		Long:  `List subnet`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := base.BizClient.DescribeSubnet(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			if global.json {
				base.PrintJSON(resp.DataSet)
			} else {
				list := make([]SubnetRow, 0)
				for _, sn := range resp.DataSet {
					row := SubnetRow{}
					row.SubnetName = sn.SubnetName
					row.ResourceID = sn.SubnetId
					row.Group = sn.Tag
					row.AffiliatedVPC = fmt.Sprintf("%s/%s", sn.VPCName, sn.VPCId)
					row.NetworkSegment = fmt.Sprintf("%s/%s", sn.Subnet, sn.Netmask)
					row.CreationTime = base.FormatDate(sn.CreateTime)
					list = append(list, row)
				}
				base.PrintTable(list, []string{"SubnetName", "ResourceID", "Group", "AffiliatedVPC", "NetworkSegment", "CreationTime"})
			}
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	req.Region = flags.String("region", base.ConfigInstance.Region, "Optional. Region, see 'ucloud region'")
	req.ProjectId = flags.String("project-id", base.ConfigInstance.ProjectID, "Optional. Project-id, see 'ucloud project list'")
	flags.StringSliceVar(&req.SubnetIds, "subnet-id", []string{}, "Optional. Multiple values separated by commas")
	req.VPCId = flags.String("vpc-id", "", "Optional. ResourceID of VPC")
	req.Tag = flags.String("group", "", "Optional. Group")
	req.Offset = flags.Int("offset", 0, "Optional. offset default 0")
	req.Limit = flags.Int("limit", 50, "Optional. max count")

	return cmd
}

//NewCmdBandwidthPkg ucloud bandwidth-pkg
func NewCmdBandwidthPkg() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bandwidth-pkg",
		Short: "List, create and delete bandwidth package",
		Long:  "List, create and delete bandwidth package",
	}
	cmd.AddCommand(NewCmdBandwidthPkgCreate())
	cmd.AddCommand(NewCmdBandwidthPkgList())
	cmd.AddCommand(NewCmdBandwidthPkgDelete())
	return cmd
}

//NewCmdBandwidthPkgCreate ucloud bandwidth-pkg create
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
		Example: "ucloud bandwidth-pkg create --eip-id eip-xxx --bandwidth-mb 20 --start-time 2018-12-15/09:20:00 --end-time 2018-12-16/09:20:00",
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
	req.Region = flags.String("region", base.ConfigInstance.Region, "Optional. Region, see 'ucloud region'")
	req.ProjectId = flags.String("project-id", base.ConfigInstance.ProjectID, "Optional. Project-id, see 'ucloud project list'")
	flags.StringSliceVar(&ids, "eip-id", nil, "Resource ID of eip to be bound with created bandwidth package")
	start = flags.String("start-time", "", "The time to enable bandwidth package. Local time, for example '2018-12-25/08:30:00'")
	end = flags.String("end-time", "", "The time to disable bandwidth package. Local time, for example '2018-12-26/08:30:00'")
	req.Bandwidth = flags.Int("bandwidth-mb", 0, "Optional, bandwidth of the bandwidth package to create, unit:'Mb'")
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

//NewCmdBandwidthPkgList ucloud bandwidth-pkg list
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
	req.Region = flags.String("region", base.ConfigInstance.Region, "Optional. Region, see 'ucloud region'")
	req.ProjectId = flags.String("project-id", base.ConfigInstance.ProjectID, "Optional. Project-id, see 'ucloud project list'")
	req.Offset = cmd.Flags().Int("offset", 0, "Optional. Offset")
	req.Limit = cmd.Flags().Int("limit", 50, "Optional. Limit range [0,10000000]")

	return cmd
}

//NewCmdBandwidthPkgDelete ucloud bandwidth-pkg delete
func NewCmdBandwidthPkgDelete() *cobra.Command {
	ids := []string{}
	req := base.BizClient.NewDeleteBandwidthPackageRequest()
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete bandwidth packages",
		Long:  "Delete bandwidth packages",
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
	req.Region = flags.String("region", base.ConfigInstance.Region, "Optional. Region, see 'ucloud region'")
	req.ProjectId = flags.String("project-id", base.ConfigInstance.ProjectID, "Optional. Project-id, see 'ucloud project list'")

	return cmd
}
