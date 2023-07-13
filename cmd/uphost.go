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
	"github.com/ucloud/ucloud-cli/base"
)

// NewCmdUPHost ucloud uphost
func NewCmdUPHost() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uphost",
		Short: "List UPHost instances",
		Long:  `List UPHost instances`,
		Args:  cobra.NoArgs,
	}
	out := base.Cxt.GetWriter()
	cmd.AddCommand(NewCmdUPHostList(out))

	return cmd
}

type uphostRow struct {
	ResourceID string
	Name       string
	PrivateIP  string
	PublicIP   string
	Config     string
	Image      string
	HostType   string
	Status     string
	Group      string
}

// NewCmdUPHostList ucloud uphost list
func NewCmdUPHostList(out io.Writer) *cobra.Command {
	ids := []string{}
	req := base.BizClient.NewDescribePHostRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List UPHost instances",
		Long:  "List UPHost instances",
		Run: func(c *cobra.Command, args []string) {
			resp, err := base.BizClient.DescribePHost(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			list := make([]uphostRow, 0)
			for _, ins := range resp.PHostSet {
				row := uphostRow{
					ResourceID: ins.PHostId,
					Name:       ins.Name,
					Config:     fmt.Sprintf("core:%d memory:%dG", ins.CPUSet.CoreCount, ins.Memory/1024),
					Group:      ins.Tag,
					HostType:   ins.PHostType,
					Status:     ins.PMStatus,
					Image:      ins.ImageName,
				}
				for _, ip := range ins.IPSet {
					if ip.OperatorName == "Private" {
						row.PrivateIP = ip.IPAddr
					} else {
						row.PublicIP = ip.IPAddr + " " + ip.OperatorName
					}
				}
				for _, disk := range ins.DiskSet {
					if disk.Name == "data" {
						row.Config += fmt.Sprintf(" data-disk:%dG %s", disk.Space, disk.Type)
					}
				}
				list = append(list, row)
			}
			base.PrintList(list, out)
		},
	}
	flags := cmd.Flags()
	bindRegion(req, flags)
	bindZoneEmpty(req, flags)
	bindProjectID(req, flags)
	bindOffset(req, flags)
	bindLimit(req, flags)
	flags.StringSliceVar(&ids, "uphost-id", nil, "Optional. Resource ID of uphost instances. List those specified uphost instances")

	return cmd
}
