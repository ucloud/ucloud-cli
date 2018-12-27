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
	"strconv"
	"time"

	"github.com/ucloud/ucloud-sdk-go/services/udpn"

	"github.com/spf13/cobra"

	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/base"
	"github.com/ucloud/ucloud-cli/model/status"
)

//NewCmdBandwidthPkg ucloud bw-pkg
func NewCmdBandwidthPkg() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bw-pkg",
		Short: "List, create and delete bandwidth package",
		Long:  "List, create and delete bandwidth package",
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
		Example: "ucloud bw-pkg create --eip-id eip-xxx --bandwidth-mb 20 --start-time 2018-12-15/09:20:00 --end-time 2018-12-16/09:20:00",
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
	req.Region = flags.String("region", base.ConfigInstance.Region, "Optional. Region, see 'ucloud region'")
	req.ProjectId = flags.String("project-id", base.ConfigInstance.ProjectID, "Optional. Project-id, see 'ucloud project list'")
	req.Offset = cmd.Flags().Int("offset", 0, "Optional. Offset")
	req.Limit = cmd.Flags().Int("limit", 50, "Optional. Limit range [0,10000000]")

	return cmd
}

//NewCmdBandwidthPkgDelete ucloud bw-pkg delete
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

//NewCmdUDPN ucloud udpn
func NewCmdUDPN(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "udpn",
		Short: "List and manipulate udpn instances",
		Long:  "List and manipulate udpn instances",
	}

	cmd.AddCommand(NewCmdUDPNCreate(out))
	cmd.AddCommand(NewCmdUDPNList(out))
	cmd.AddCommand(NewCmdUdpnDelete(out))
	cmd.AddCommand(NewCmdUdpnModifyBW(out))

	return cmd
}

//NewCmdUDPNCreate ucloud udpn create
func NewCmdUDPNCreate(out io.Writer) *cobra.Command {
	req := base.BizClient.NewAllocateUDPNRequest()
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create UDPN tunnel",
		Long:  "Create UDPN tunnel",
		Run: func(c *cobra.Command, args []string) {
			if *req.Bandwidth < 2 || *req.Bandwidth > 1000 {
				fmt.Fprintln(out, "Error, bandwidth must be between 2Mb and 1000Mb")
				return
			}
			if *req.Peer1 == *req.Peer2 {
				fmt.Fprintln(out, "Error, flags peer1 and peer2 can't be equal")
				return
			}
			resp, err := base.BizClient.AllocateUDPN(req)
			req.ProjectId = sdk.String(base.PickResourceID(*req.ProjectId))
			if err != nil {
				base.HandleError(err)
				return
			}
			fmt.Fprintf(out, "udpn[%s] created\n", resp.UDPNId)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.Peer1 = flags.String("peer1", base.ConfigInstance.Region, "Required. One end of the tunnel to create")
	req.Peer2 = flags.String("peer2", "", "Required. The other end of the tunnel create")
	req.Bandwidth = flags.Int("bandwidth-mb", 0, "Required. Bandwidth of the tunnel to create. Unit:Mb. Rnange [2,1000]")
	req.ChargeType = flags.String("charge-type", "", "Optional. Enumeration value.'Year',pay yearly;'Month',pay monthly;'Dynamic', pay hourly")
	req.Quantity = cmd.Flags().Int("quantity", 1, "Optional. The duration of the instance. N years/months.")
	req.ProjectId = flags.String("project-id", base.ConfigInstance.ProjectID, "Optional. Project-id, see 'ucloud project list'")

	flags.SetFlagValues("charge-type", "Month", "Year", "Dynamic")
	flags.SetFlagValuesFunc("project-id", getProjectList)
	flags.SetFlagValuesFunc("peer1", getRegionList)
	//peer1和peer2不相等
	flags.SetFlagValuesFunc("peer2", func() []string {
		regions := getRegionList()
		list := []string{}
		for _, r := range regions {
			if r != *req.Peer1 {
				list = append(list, r)
			}
		}
		return list
	})

	cmd.MarkFlagRequired("peer1")
	cmd.MarkFlagRequired("peer2")
	cmd.MarkFlagRequired("bandwidth-mb")

	return cmd
}

//UDPNRow 表格行
type UDPNRow struct {
	ResourceID   string
	Peers        string
	Bandwidth    string
	ChargeType   string
	CreationTime string
}

//NewCmdUDPNList ucloud udpn list
func NewCmdUDPNList(out io.Writer) *cobra.Command {
	req := base.BizClient.NewDescribeUDPNRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List udpn instances",
		Long:  "List udpn instances",
		Run: func(c *cobra.Command, args []string) {
			req.UDPNId = sdk.String(base.PickResourceID(*req.UDPNId))
			resp, err := base.BizClient.DescribeUDPN(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			list := []UDPNRow{}
			for _, udpn := range resp.DataSet {
				row := UDPNRow{}
				row.ResourceID = udpn.UDPNId
				row.Peers = fmt.Sprintf("%s <--> %s", udpn.Peer1, udpn.Peer2)
				row.Bandwidth = fmt.Sprintf("%dMb", udpn.Bandwidth)
				row.ChargeType = udpn.ChargeType
				row.CreationTime = base.FormatDate(udpn.CreateTime)
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

	req.UDPNId = flags.String("udpn-id", "", "Optional. Resource ID of udpn instances to list")
	req.Offset = flags.Int("offset", 0, "Optional. Offset")
	req.Limit = flags.Int("limit", 50, "Optional. Limit")
	req.Region = flags.String("region", base.ConfigInstance.Region, "Optional. Region, see 'ucloud region'")
	req.ProjectId = flags.String("project-id", base.ConfigInstance.ProjectID, "Optional. Project-id, see 'ucloud project list'")

	flags.SetFlagValuesFunc("region", getRegionList)
	flags.SetFlagValuesFunc("project-id", getRegionList)
	flags.SetFlagValuesFunc("udpn-id", func() []string {
		return getAllUDPNIdNames(*req.ProjectId, *req.Region)
	})

	return cmd
}

//NewCmdUdpnDelete ucloud udpn delete
func NewCmdUdpnDelete(out io.Writer) *cobra.Command {
	idNames := []string{}
	req := base.BizClient.NewReleaseUDPNRequest()
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "delete udpn instances",
		Long:  "delete udpn instances",
		Run: func(c *cobra.Command, args []string) {
			for _, idname := range idNames {
				req.UDPNId = sdk.String(base.PickResourceID(idname))
				_, err := base.BizClient.ReleaseUDPN(req)
				if err != nil {
					base.HandleError(err)
					continue
				}
				fmt.Fprintf(out, "udpn[%s] deleted\n", idname)
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&idNames, "udpn-id", nil, "Required. Resource ID of udpn instances to delete")
	req.ProjectId = flags.String("project-id", base.ConfigInstance.ProjectID, "Optional. Project-id, see 'ucloud project list'")

	flags.SetFlagValuesFunc("project-id", getRegionList)
	flags.SetFlagValuesFunc("udpn-id", func() []string {
		return getAllUDPNIdNames(*req.ProjectId, base.ConfigInstance.Region)
	})

	cmd.MarkFlagRequired("udpn-id")

	return cmd
}

//NewCmdUdpnModifyBW ucloud udpn modify-bw
func NewCmdUdpnModifyBW(out io.Writer) *cobra.Command {
	idNames := []string{}
	req := base.BizClient.NewModifyUDPNBandwidthRequest()
	cmd := &cobra.Command{
		Use:   "modify-bw",
		Short: "Modify bandwidth of UDPN tunnel",
		Long:  "Modify bandwidth of UDPN tunnel",
		Run: func(c *cobra.Command, args []string) {
			if *req.Bandwidth < 2 || *req.Bandwidth > 1000 {
				fmt.Fprintln(out, "Error, bandwidth must be between 2Mb and 1000Mb")
				return
			}
			for _, idname := range idNames {
				req.UDPNId = sdk.String(base.PickResourceID(idname))
				_, err := base.BizClient.ModifyUDPNBandwidth(req)
				if err != nil {
					base.HandleError(err)
					return
				}
				fmt.Fprintf(out, "udpn[%s]'s bandwidth modified\n", idname)
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&idNames, "udpn-id", nil, "Required. Resource ID of UDPN to modify bandwidth")
	req.Bandwidth = flags.Int("bandwidth-mb", 0, "Required. Bandwidth of UDPN tunnel. Unit:Mb. Range [2,1000]")
	req.Region = flags.String("region", base.ConfigInstance.Region, "Optional. Region, see 'ucloud region'")
	req.ProjectId = flags.String("project-id", base.ConfigInstance.ProjectID, "Optional. Project-id, see 'ucloud project list'")

	flags.SetFlagValuesFunc("udpn-id", func() []string {
		return getAllUDPNIdNames(*req.ProjectId, *req.Region)
	})

	cmd.MarkFlagRequired("udpn-id")
	cmd.MarkFlagRequired("bandwidth-mb")

	return cmd
}

func getAllUDPNIns(project, region string) ([]udpn.UDPNData, error) {
	req := base.BizClient.NewDescribeUDPNRequest()
	req.ProjectId = sdk.String(project)
	req.Region = sdk.String(region)
	list := make([]udpn.UDPNData, 0)
	for offset, limit := 0, 50; ; offset += limit {
		req.Offset = sdk.Int(offset)
		req.Limit = sdk.Int(limit)
		resp, err := base.BizClient.DescribeUDPN(req)
		if err != nil {
			return nil, err
		}
		for _, u := range resp.DataSet {
			list = append(list, u)
		}
		if offset+limit > resp.TotalCount {
			break
		}
	}
	return list, nil
}

func getAllUDPNIdNames(project, region string) []string {
	udpnInsList, err := getAllUDPNIns(project, region)
	if err != nil {
		return nil
	}
	idNameList := []string{}
	for _, udpn := range udpnInsList {
		idNameList = append(idNameList, fmt.Sprintf("%s/%s:%s", udpn.UDPNId, udpn.Peer1, udpn.Peer2))
	}
	return idNameList
}
