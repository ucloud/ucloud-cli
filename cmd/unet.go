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

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/udpn"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/base"
)

// NewCmdUDPN ucloud udpn
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

// NewCmdUDPNCreate ucloud udpn create
func NewCmdUDPNCreate(out io.Writer) *cobra.Command {
	req := base.BizClient.NewAllocateUDPNRequest()
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create UDPN tunnel",
		Long:  "Create UDPN tunnel",
		Run: func(c *cobra.Command, args []string) {
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

	req.Peer1 = flags.String("peer1", base.ConfigIns.Region, "Required. One end of the tunnel to create")
	req.Peer2 = flags.String("peer2", "", "Required. The other end of the tunnel create")
	req.Bandwidth = flags.Int("bandwidth-mb", 0, "Required. Bandwidth of the tunnel to create. Unit:Mb. Rnange [2,1000]")
	req.ChargeType = flags.String("charge-type", "", "Optional. Enumeration value.'Year',pay yearly;'Month',pay monthly;'Dynamic', pay hourly")
	req.Quantity = cmd.Flags().Int("quantity", 1, "Optional. The duration of the instance. N years/months.")
	req.ProjectId = flags.String("project-id", base.ConfigIns.ProjectID, "Optional. Project-id, see 'ucloud project list'")

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

// UDPNRow 表格行
type UDPNRow struct {
	ResourceID   string
	Peers        string
	Bandwidth    string
	ChargeType   string
	CreationTime string
}

// NewCmdUDPNList ucloud udpn list
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
			base.PrintList(list, out)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.UDPNId = flags.String("udpn-id", "", "Optional. Resource ID of udpn instances to list")
	req.Offset = flags.Int("offset", 0, "Optional. Offset")
	req.Limit = flags.Int("limit", 50, "Optional. Limit")
	req.Region = flags.String("region", base.ConfigIns.Region, "Optional. Region, see 'ucloud region'")
	req.ProjectId = flags.String("project-id", base.ConfigIns.ProjectID, "Optional. Project-id, see 'ucloud project list'")

	flags.SetFlagValuesFunc("region", getRegionList)
	flags.SetFlagValuesFunc("project-id", getRegionList)
	flags.SetFlagValuesFunc("udpn-id", func() []string {
		return getAllUDPNIdNames(*req.ProjectId, *req.Region)
	})

	return cmd
}

// NewCmdUdpnDelete ucloud udpn delete
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
	req.ProjectId = flags.String("project-id", base.ConfigIns.ProjectID, "Optional. Project-id, see 'ucloud project list'")

	flags.SetFlagValuesFunc("project-id", getRegionList)
	flags.SetFlagValuesFunc("udpn-id", func() []string {
		return getAllUDPNIdNames(*req.ProjectId, base.ConfigIns.Region)
	})

	cmd.MarkFlagRequired("udpn-id")

	return cmd
}

// NewCmdUdpnModifyBW ucloud udpn modify-bw
func NewCmdUdpnModifyBW(out io.Writer) *cobra.Command {
	idNames := []string{}
	req := base.BizClient.NewModifyUDPNBandwidthRequest()
	cmd := &cobra.Command{
		Use:   "modify-bw",
		Short: "Modify bandwidth of UDPN tunnel",
		Long:  "Modify bandwidth of UDPN tunnel",
		Run: func(c *cobra.Command, args []string) {
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
	req.Region = flags.String("region", base.ConfigIns.Region, "Optional. Region, see 'ucloud region'")
	req.ProjectId = flags.String("project-id", base.ConfigIns.ProjectID, "Optional. Project-id, see 'ucloud project list'")

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
