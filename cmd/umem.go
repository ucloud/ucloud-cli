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
	"github.com/ucloud/ucloud-cli/base"
)

//NewCmdUHost ucloud uhost
func NewCmdUMem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "umem",
		Short: "List Umem instance",
		Long:  `List Umem instance`,
		Args:  cobra.NoArgs,
	}
	// writer := base.Cxt.GetWriter()
	cmd.AddCommand(NewCmdRedis())
	cmd.AddCommand(NewCmdMemcache())

	return cmd
}
func NewCmdMemcache() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "memcache",
		Short: "List memcache instance",
		Long:  `List memcache instance`,
		Args:  cobra.NoArgs,
	}
	// writer := base.Cxt.GetWriter()
	cmd.AddCommand(NewCmdMemcacheList())

	return cmd
}

func NewCmdRedis() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "redis",
		Short: "List redis instance",
		Long:  `List redis instance`,
		Args:  cobra.NoArgs,
	}
	// writer := base.Cxt.GetWriter()
	cmd.AddCommand(NewCmdRedisList())

	return cmd
}

//UMemcacheRow UHost表格行
type UMemcacheRow struct {
	Name         string
	ResourceID   string
	Group        string
	VirtualIP    string
	Size         int
	UsedSize     int
	State        string
	CreationTime string
}

func NewCmdMemcacheList() *cobra.Command {
	req := base.BizClient.NewDescribeUMemcacheGroupRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all memcache Instances",
		Long:  `List all memcache Instances`,
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := base.BizClient.DescribeUMemcacheGroup(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			list := make([]UMemcacheRow, 0)
			for _, host := range resp.DataSet {
				row := UMemcacheRow{}
				row.Name = host.Name
				row.ResourceID = host.GroupId
				row.Group = host.Tag
				row.VirtualIP = host.VirtualIP

				row.Size = host.Size
				row.UsedSize = host.UsedSize
				row.CreationTime = base.FormatDate(host.CreateTime)
				row.State = host.State
				list = append(list, row)
			}
			base.PrintList(list)
		},
	}
	cmd.Flags().SortFlags = false
	req.ProjectId = cmd.Flags().String("project-id", base.ConfigIns.ProjectID, "Optional. Assign project-id")
	req.Region = cmd.Flags().String("region", base.ConfigIns.Region, "Optional. Assign region")
	req.Zone = cmd.Flags().String("zone", "", "Optional. Assign availability zone")
	req.Offset = cmd.Flags().Int("offset", 0, "Optional. Offset default 0")
	req.Limit = cmd.Flags().Int("limit", 50, "Optional. Limit default 50, max value 100")
	//bindGroup(req, cmd.Flags())

	return cmd
}

//UHostRow UHost表格行
type URedisRow struct {
	Name         string
	ResourceID   string
	Group        string
	VirtualIP    string
	Size         int
	UsedSize     int
	State        string
	CreationTime string
}

//NewCmdUHostList [ucloud uhost list]
func NewCmdRedisList() *cobra.Command {
	req := base.BizClient.NewDescribeURedisGroupRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all UHost Instances",
		Long:  `List all UHost Instances`,
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := base.BizClient.DescribeURedisGroup(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			list := make([]URedisRow, 0)
			for _, host := range resp.DataSet {
				row := URedisRow{}
				row.Name = host.Name
				row.ResourceID = host.GroupId
				row.Group = host.Tag
				row.VirtualIP = host.VirtualIP

				row.Size = host.Size
				row.UsedSize = host.UsedSize
				row.CreationTime = base.FormatDate(host.CreateTime)
				row.State = host.State
				list = append(list, row)
			}
			base.PrintList(list)
		},
	}
	cmd.Flags().SortFlags = false
	req.ProjectId = cmd.Flags().String("project-id", base.ConfigIns.ProjectID, "Optional. Assign project-id")
	req.Region = cmd.Flags().String("region", base.ConfigIns.Region, "Optional. Assign region")
	req.Zone = cmd.Flags().String("zone", "", "Optional. Assign availability zone")
	req.Offset = cmd.Flags().Int("offset", 0, "Optional. Offset default 0")
	req.Limit = cmd.Flags().Int("limit", 50, "Optional. Limit default 50, max value 100")

	return cmd
}
