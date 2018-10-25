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
	"strings"

	"github.com/spf13/cobra"
	. "github.com/ucloud/ucloud-cli/base"
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
	UGroup         string
	AffiliatedVPC  string
	NetworkSegment string
	CreationTime   string
}

//NewCmdSubnetList ucloud subnet list
func NewCmdSubnetList() *cobra.Command {
	req := BizClient.NewDescribeSubnetRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List subnet",
		Long:  `List subnet`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := BizClient.DescribeSubnet(req)
			if err != nil {
				HandleError(err)
				return
			}
			if global.json {
				PrintJSON(resp.DataSet)
			} else {
				list := make([]SubnetRow, 0)
				for _, sn := range resp.DataSet {
					row := SubnetRow{}
					row.SubnetName = sn.SubnetName
					row.ResourceID = sn.SubnetId
					row.UGroup = sn.Tag
					row.AffiliatedVPC = fmt.Sprintf("%s/%s", sn.VPCName, sn.VPCId)
					row.NetworkSegment = fmt.Sprintf("%s/%s", sn.Subnet, sn.Netmask)
					row.CreationTime = FormatDate(sn.CreateTime)
					list = append(list, row)
				}
				PrintTable(list, []string{"SubnetName", "ResourceID", "UGroup", "AffiliatedVPC", "NetworkSegment", "CreationTime"})
			}
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	req.Region = flags.String("region", ConfigInstance.Region, "Optional. Region, see 'ucloud region'")
	req.ProjectId = flags.String("project-id", ConfigInstance.ProjectID, "Optional. Project-id, see 'ucloud project list'")
	flags.StringSliceVar(&req.SubnetIds, "subnet-id", []string{}, "Optional. Multiple values separated by commas")
	req.VPCId = flags.String("vpc-id", "", "Optional. ResourceID of VPC")
	req.Tag = flags.String("ugroup", "", "Optional. UGroup")
	req.Offset = flags.Int("offset", 0, "Optional. offset default 0")
	req.Limit = flags.Int("limit", 50, "Optional. max count")

	return cmd
}

//NewCmdVPC  ucloud vpc
func NewCmdVPC() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vpc",
		Short: "List vpc",
		Long:  `List vpc`,
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(NewCmdVPCList())

	return cmd
}

//VPCRow 表格行
type VPCRow struct {
	VPCName        string
	ResourceID     string
	UGroup         string
	NetworkSegment string
	SubnetCount    int
	CreationTime   string
}

//NewCmdVPCList ucloud vpc list
func NewCmdVPCList() *cobra.Command {
	req := BizClient.NewDescribeVPCRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List vpc",
		Long:  "List vpc",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := BizClient.DescribeVPC(req)
			if err != nil {
				HandleError(err)
				return
			}
			if global.json {
				PrintJSON(resp.DataSet)
			} else {
				list := []VPCRow{}
				for _, vpc := range resp.DataSet {
					row := VPCRow{}
					row.VPCName = vpc.Name
					row.ResourceID = vpc.VPCId
					row.UGroup = vpc.Tag
					row.NetworkSegment = strings.Join(vpc.Network, ",")
					row.SubnetCount = vpc.SubnetCount
					row.CreationTime = FormatDate(vpc.CreateTime)
					list = append(list, row)
				}
				PrintTable(list, []string{"VPCName", "ResourceID", "UGroup", "NetworkSegment", "SubnetCount", "CreationTime"})
			}

		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.Region = flags.String("region", ConfigInstance.Region, "Optional. Region, see 'ucloud region'")
	req.ProjectId = flags.String("project-id", ConfigInstance.ProjectID, "Optional. Project-id, see 'ucloud project list'")
	req.Tag = flags.String("ugroup", "", "Optional. UGroup")
	flags.StringSliceVar(&req.VPCIds, "vpc-id", []string{}, "Optional. Multiple values separated by commas")

	return cmd
}

//NewCmdFirewall  ucloud firewall
func NewCmdFirewall() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "firewall",
		Short: "List extranet firewall",
		Long:  `List extranet firewall`,
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(NewCmdFirewallList())

	return cmd
}

//FirewallRow 表格行
type FirewallRow struct {
	ResourceID          string
	FirewallName        string
	Remark              string
	UGroup              string
	RuleAmount          int
	BoundResourceAmount int
	CreationTime        string
}

//NewCmdFirewallList ucloud firewall list
func NewCmdFirewallList() *cobra.Command {
	req := BizClient.NewDescribeFirewallRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List extranet firewall",
		Long:  `List extranet firewall`,
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := BizClient.DescribeFirewall(req)
			if err != nil {
				HandleError(err)
				return
			}
			if global.json {
				PrintJSON(resp.DataSet)
			} else {
				list := []FirewallRow{}
				for _, fw := range resp.DataSet {
					row := FirewallRow{}
					row.ResourceID = fw.FWId
					row.FirewallName = fw.Name
					row.Remark = fw.Remark
					row.UGroup = fw.Tag
					row.RuleAmount = len(fw.Rule)
					row.BoundResourceAmount = fw.ResourceCount
					row.CreationTime = FormatDate(fw.CreateTime)
					list = append(list, row)
				}
				PrintTable(list, []string{"ResourceID", "FirewallName", "Remark", "UGroup", "RuleAmount", "BoundResourceAmount", "CreationTime"})
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.Region = flags.String("region", ConfigInstance.Region, "Optional. Region, see 'ucloud region'")
	req.ProjectId = flags.String("project-id", ConfigInstance.ProjectID, "Optional. Project-id, see 'ucloud project list'")
	req.FWId = flags.String("firewall-id", "", "Optional. The Resource ID of firewall. Return all firewalls by default.")
	req.ResourceType = flags.String("bound-resource-type", "", "Optional. The type of resource bound on the firewall")
	req.ResourceId = flags.String("bound-resource-id", "", "Optional. The resource ID of resource bound on the firewall")
	req.Offset = flags.String("offset", "0", "Optional. offset default 0")
	req.Limit = flags.String("limit", "50", "Optional. max count")
	return cmd
}
