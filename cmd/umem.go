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
	"strings"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/umem"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/base"
	"github.com/ucloud/ucloud-cli/model/status"
	"github.com/ucloud/ucloud-cli/ux"
)

// NewCmdRedis ucloud redis
func NewCmdRedis() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "redis",
		Short: "List and manipulate redis instances",
		Long:  "List and manipulate redis instances",
	}
	out := base.Cxt.GetWriter()
	cmd.AddCommand(NewCmdRedisList(out))
	cmd.AddCommand(NewCmdRedisCreate(out))
	cmd.AddCommand(NewCmdRedisDelete(out))
	cmd.AddCommand(NewCmdRedisRestart(out))
	return cmd
}

// NewCmdMemcache ucloud memcache
func NewCmdMemcache() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "memcache",
		Short: "List and manipulate memcache instances",
		Long:  "List and manipulate memcache instances",
	}
	out := base.Cxt.GetWriter()
	cmd.AddCommand(NewCmdMemcacheList(out))
	cmd.AddCommand(NewCmdMemcacheCreate(out))
	cmd.AddCommand(NewCmdMemcacheDelete(out))
	cmd.AddCommand(NewCmdMemcacheRestart(out))
	return cmd
}

// UMemRedisRow 表格行
type UMemRedisRow struct {
	ResourceID string
	Name       string
	Role       string
	Type       string
	Address    string
	Size       string
	UsedSize   string
	State      string
	Group      string
	Zone       string
	CreateTime string
}

var redisTypeMap = map[string]string{
	"single":      "master-replica",
	"distributed": "distributed",
}

// NewCmdRedisList ucloud redis list
func NewCmdRedisList(out io.Writer) *cobra.Command {
	req := base.BizClient.NewDescribeUMemRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List redis instances",
		Long:  "List redis instances",
		Run: func(c *cobra.Command, args []string) {
			resp, err := base.BizClient.DescribeUMem(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			list := []UMemRedisRow{}
			for _, ins := range resp.DataSet {
				row := UMemRedisRow{
					ResourceID: ins.ResourceId,
					Name:       ins.Name,
					Role:       ins.Role,
					Type:       redisTypeMap[ins.ResourceType],
					Group:      ins.Tag,
					Size:       fmt.Sprintf("%dGB", ins.Size),
					UsedSize:   fmt.Sprintf("%dMB", ins.UsedSize),
					State:      ins.State,
					Zone:       ins.Zone,
					CreateTime: base.FormatDate(ins.CreateTime),
				}
				addrs := []string{}
				for _, addr := range ins.Address {
					addrs = append(addrs, fmt.Sprintf("%s:%d", addr.IP, addr.Port))
				}
				row.Address = strings.Join(addrs, "|")
				list = append(list, row)
				for _, slave := range ins.DataSet {
					srow := UMemRedisRow{
						ResourceID: slave.GroupId,
						Name:       slave.Name,
						Role:       fmt.Sprintf("\u2b91 %s", slave.Role),
						Type:       redisTypeMap[slave.ResourceType],
						Group:      slave.Tag,
						Size:       fmt.Sprintf("%dGB", slave.Size),
						UsedSize:   fmt.Sprintf("%dMB", slave.UsedSize),
						State:      slave.State,
						Zone:       slave.Zone,
						Address:    fmt.Sprintf("%s:%d", slave.VirtualIP, slave.Port),
						CreateTime: base.FormatDate(slave.CreateTime),
					}
					list = append(list, srow)
				}
			}
			base.PrintList(list, out)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.ResourceId = flags.String("umem-id", "", "Optional. Resource ID of the redis to list")
	bindRegion(req, flags)
	bindZoneEmpty(req, flags)
	bindProjectID(req, flags)
	bindOffset(req, flags)
	bindLimit(req, flags)
	req.Protocol = sdk.String("redis")

	flags.SetFlagValuesFunc("umem-id", func() []string {
		return getRedisIDList(*req.ProjectId, *req.Region)
	})

	return cmd
}

// NewCmdRedisCreate ucloud redis create
func NewCmdRedisCreate(out io.Writer) *cobra.Command {
	req := base.BizClient.NewCreateURedisGroupRequest()
	req.HighAvailability = sdk.String("enable")
	var redisType, password string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create redis instance",
		Long:  "Create redis instance",
		Run: func(c *cobra.Command, args []string) {
			if l := len(*req.Name); l < 6 || l > 63 {
				fmt.Fprintln(out, "length of name should be between 6 and 63")
				return
			}
			if password != "" {
				req.Password = &password
			}
			if redisType == "master-replica" {
				resp, err := base.BizClient.CreateURedisGroup(req)
				if err != nil {
					base.HandleError(err)
					return
				}
				fmt.Printf("redis[%s] created\n", resp.GroupId)
			} else if redisType == "distributed" {
				dreq := base.BizClient.NewCreateUMemSpaceRequest()
				dreq.Region = req.Region
				dreq.Zone = req.Zone
				dreq.ProjectId = req.ProjectId
				dreq.Name = req.Name
				dreq.Size = req.Size
				if *req.Size == 1 {
					dreq.Size = sdk.Int(16)
				}
				dreq.ChargeType = req.ChargeType
				dreq.Quantity = req.Quantity
				dreq.Tag = req.Tag
				dreq.Password = req.Password
				resp, err := base.BizClient.CreateUMemSpace(dreq)
				if err != nil {
					base.HandleError(err)
					return
				}
				fmt.Printf("redis[%s] created\n", resp.SpaceId)
			} else {
				fmt.Printf("unknow redis type[%s], it's should be 'master-replica' or 'distributed'\n", redisType)
			}
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.Name = flags.String("name", "", "Required. Name of the redis to create. Range of the password length is [6,63] and the password can only contain letters and numbers")
	flags.StringVar(&redisType, "type", "", "Required. Type of the redis. Accept values:'master-replica','distributed'")
	req.Size = flags.Int("size-gb", 1, "Optional. Memory size. Default value 1GB(for master-replica redis type) or 16GB(for distributed redis type). Unit GB")
	req.Version = flags.String("version", "3.2", "Optional. Version of redis")
	req.VPCId = flags.String("vpc-id", "", "Optional. VPC ID. This field is required under VPC2.0. See 'ucloud vpc list'")
	req.SubnetId = flags.String("subnet-id", "", "Optional. Subnet ID. This field is required under VPC2.0. See 'ucloud subnet list'")
	flags.StringVar(&password, "password", "", "Optional. Password of redis to create")

	bindRegion(req, flags)
	bindZone(req, flags)
	bindProjectID(req, flags)
	bindGroup(req, flags)
	bindChargeType(req, flags)
	bindQuantity(req, flags)

	flags.SetFlagValues("version", "3.0", "3.2", "4.0")
	flags.SetFlagValues("type", "master-replica", "distributed")
	flags.SetFlagValuesFunc("vpc-id", func() []string {
		return getAllVPCIdNames(*req.ProjectId, *req.Region)
	})
	flags.SetFlagValuesFunc("subnet-id", func() []string {
		return getAllSubnetIDNames(*req.VPCId, *req.ProjectId, *req.Region)
	})

	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("type")

	return cmd
}

// NewCmdRedisDelete ucloud redis delete
func NewCmdRedisDelete(out io.Writer) *cobra.Command {
	var idNames []string
	req := base.BizClient.NewDeleteURedisGroupRequest()
	cmd := &cobra.Command{
		Use:     "delete",
		Short:   "Delete redis instances",
		Long:    "Delete redis instances",
		Example: "ucloud redis delete --umem-id uredis-rl5xuxx/testcli1,uredis-xsdfa/testcli2",
		Run: func(c *cobra.Command, args []string) {
			for _, idname := range idNames {
				id := base.PickResourceID(idname)
				if strings.HasPrefix(id, "uredis") {
					req.GroupId = &id
					_, err := base.BizClient.DeleteURedisGroup(req)
					if err != nil {
						base.HandleError(err)
						continue
					}
				} else if strings.HasPrefix(id, "umem") {
					_req := base.BizClient.NewDeleteUMemSpaceRequest()
					_req.Region = req.Region
					_req.Zone = req.Zone
					_req.ProjectId = req.ProjectId
					_req.SpaceId = &id
					_, err := base.BizClient.DeleteUMemSpace(_req)
					if err != nil {
						base.HandleError(err)
						continue
					}
				}
				fmt.Fprintf(out, "redis[%s] deleted\n", idname)
			}
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&idNames, "umem-id", nil, "Required. Resource ID of redis intances to delete")
	bindProjectID(req, flags)
	bindRegion(req, flags)
	bindZone(req, flags)

	cmd.MarkFlagRequired("umem-id")

	flags.SetFlagValuesFunc("umem-id", func() []string {
		return getRedisIDList(*req.ProjectId, *req.Region)
	})

	return cmd
}

// NewCmdRedisRestart ucloud redis restart
func NewCmdRedisRestart(out io.Writer) *cobra.Command {
	idNames := make([]string, 0)
	req := base.BizClient.UMemClient.NewRestartURedisGroupRequest()
	cmd := &cobra.Command{
		Use:   "restart",
		Short: "Restart redis instances of master-replica type",
		Long:  "Restart redis instances of master-replica type",
		Run: func(c *cobra.Command, args []string) {
			reqs := make([]request.Common, len(idNames))
			for idx, idname := range idNames {
				id := base.PickResourceID(idname)
				_req := *req
				_req.GroupId = &id
				reqs[idx] = &_req
			}
			coAction := newConcurrentAction(reqs, 10, restartRedis)
			coAction.Do()
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&idNames, "umem-id", nil, "Required. Resource ID of redis instances to restart")
	bindProjectID(req, flags)
	bindRegion(req, flags)
	bindZone(req, flags)

	cmd.MarkFlagRequired("umem-id")
	flags.SetFlagValuesFunc("umem-id", func() []string {
		return getRedisIDList(*req.ProjectId, *req.Region)
	})

	return cmd
}

func restartRedis(creq request.Common) (bool, []string) {
	req := creq.(*umem.RestartURedisGroupRequest)
	block := ux.NewBlock()
	ux.Doc.Append(block)
	logs := make([]string, 0)
	logs = append(logs, fmt.Sprintf("api:RestartURedisGroup, request:%v", base.ToQueryMap(req)))
	_, err := base.BizClient.UMemClient.RestartURedisGroup(req)
	if err != nil {
		block.Append(base.ParseError(err))
		logs = append(logs, fmt.Sprintf("restart redis[%s] failed: %s", *req.GroupId, base.ParseError(err)))
		return false, logs
	}
	poller := base.NewSpoller(describeRedisByID, base.Cxt.GetWriter())
	text := fmt.Sprintf("redis[%s] is restarting", *req.GroupId)
	ret := poller.Sspoll(*req.GroupId, text, []string{status.UMEM_RUNNING, status.UMEM_FAIL}, block)
	if ret.Err != nil {
		block.Append(base.ParseError(err))
		logs = append(logs, ret.Err.Error())
	}
	if ret.Timeout {
		logs = append(logs, "poll redis[%s] timeout", *req.GroupId)
	}
	return ret.Done, logs
}

func getRedisIDList(project, region string) []string {
	req := base.BizClient.NewDescribeURedisGroupRequest()
	req.ProjectId = &project
	req.Region = &region
	list := []string{}

	for limit, offset := 50, 0; ; offset += limit {
		req.Limit = sdk.Int(limit)
		req.Offset = sdk.Int(offset)
		resp, err := base.BizClient.DescribeURedisGroup(req)
		if err != nil {
			return nil
		}
		for _, ins := range resp.DataSet {
			list = append(list, fmt.Sprintf("%s/%s", ins.GroupId, ins.Name))
		}
		if offset+limit >= resp.TotalCount {
			break
		}
	}
	return list
}

// UMemMemcacheRow 表格行
type UMemMemcacheRow struct {
	ResourceID string
	Name       string
	Address    string
	Size       string
	UsedSize   string
	State      string
	Group      string
	CreateTime string
}

// NewCmdMemcacheList ucloud memcache list
func NewCmdMemcacheList(out io.Writer) *cobra.Command {
	req := base.BizClient.NewDescribeUMemcacheGroupRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List memcache instances",
		Long:  "List memcache instances",
		Run: func(c *cobra.Command, args []string) {
			resp, err := base.BizClient.DescribeUMemcacheGroup(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			list := []UMemMemcacheRow{}
			for _, ins := range resp.DataSet {
				row := UMemMemcacheRow{
					ResourceID: ins.GroupId,
					Name:       ins.Name,
					Group:      ins.Tag,
					Size:       fmt.Sprintf("%dGB", ins.Size),
					UsedSize:   fmt.Sprintf("%dMB", ins.UsedSize),
					State:      ins.State,
					CreateTime: base.FormatDate(ins.CreateTime),
					Address:    fmt.Sprintf("%s:%d", ins.VirtualIP, ins.Port),
				}
				list = append(list, row)
			}
			base.PrintList(list, out)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.GroupId = flags.String("umem-id", "", "Optional. Resource ID of the redis to list")
	bindRegion(req, flags)
	bindZoneEmpty(req, flags)
	bindProjectID(req, flags)
	bindOffset(req, flags)
	bindLimit(req, flags)

	return cmd
}

// NewCmdMemcacheCreate ucloud memcache create
func NewCmdMemcacheCreate(out io.Writer) *cobra.Command {
	req := base.BizClient.NewCreateUMemcacheGroupRequest()
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create memcache instance",
		Long:  "Create memcache instance",
		Run: func(c *cobra.Command, args []string) {
			if *req.Size > 32 || *req.Size < 1 {
				fmt.Fprintln(out, "size-gb should be between 1 and 32")
				return
			}
			resp, err := base.BizClient.CreateUMemcacheGroup(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			fmt.Fprintf(out, "memcache[%s] created\n", resp.GroupId)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.Name = flags.String("name", "", "Required. Name of memcache instance to create")
	req.Size = flags.Int("size-gb", 1, "Optional. Memory size of memcache instance. Unit GB. Accpet values:1,2,4,8,16,32")
	req.VPCId = flags.String("vpc-id", "", "Optional. VPC ID. See 'ucloud vpc list'")
	req.SubnetId = flags.String("subnet-id", "", "Optional. Subnet ID. See 'ucloud subnet list'")
	bindProjectID(req, flags)
	bindRegion(req, flags)
	bindZone(req, flags)
	bindChargeType(req, flags)
	bindQuantity(req, flags)
	bindGroup(req, flags)

	flags.SetFlagValues("size-gb", "1", "2", "4", "8", "16", "32")
	flags.SetFlagValuesFunc("vpc-id", func() []string {
		return getAllVPCIdNames(*req.ProjectId, *req.Region)
	})
	flags.SetFlagValuesFunc("subnet-id", func() []string {
		return getAllSubnetIDNames(*req.VPCId, *req.ProjectId, *req.Region)
	})

	cmd.MarkFlagRequired("name")

	return cmd
}

// NewCmdMemcacheDelete ucloud memcache delete
func NewCmdMemcacheDelete(out io.Writer) *cobra.Command {
	var idNames []string
	req := base.BizClient.NewDeleteUMemcacheGroupRequest()
	cmd := &cobra.Command{
		Use:     "delete",
		Short:   "Delete memcache instances",
		Long:    "Delete memcache instances",
		Example: "ucloud memcache delete --umem-id umemcache-rl5xuxx/testcli1,umemcache-xsdfa/testcli2",
		Run: func(c *cobra.Command, args []string) {
			for _, idname := range idNames {
				id := base.PickResourceID(idname)
				req.GroupId = &id
				_, err := base.BizClient.DeleteUMemcacheGroup(req)
				if err != nil {
					base.HandleError(err)
					continue
				}
				fmt.Fprintf(out, "memcache[%s] deleted\n", idname)
			}
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&idNames, "umem-id", nil, "Required. Resource ID of memcache intances to delete")
	bindProjectID(req, flags)
	bindRegion(req, flags)
	bindZoneEmpty(req, flags)

	cmd.MarkFlagRequired("umem-id")

	flags.SetFlagValuesFunc("umem-id", func() []string {
		return getMemcacheIDList(*req.ProjectId, *req.Region)
	})

	return cmd
}

// NewCmdMemcacheRestart ucloud memcache restart
func NewCmdMemcacheRestart(out io.Writer) *cobra.Command {
	idNames := make([]string, 0)
	req := base.BizClient.NewRestartUMemcacheGroupRequest()
	cmd := &cobra.Command{
		Use:   "restart",
		Short: "Restart memcache instances",
		Long:  "Restart memcache instances",
		Run: func(c *cobra.Command, args []string) {
			reqs := make([]request.Common, len(idNames))
			for idx, idname := range idNames {
				id := base.PickResourceID(idname)
				_req := *req
				_req.GroupId = &id
				reqs[idx] = &_req
			}
			coAction := newConcurrentAction(reqs, 10, restartMemcache)
			coAction.Do()
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&idNames, "umem-id", nil, "Required. Resource ID of memcache to restart")
	bindRegion(req, flags)
	bindZone(req, flags)
	bindProjectID(req, flags)

	flags.SetFlagValuesFunc("umem-id", func() []string {
		return getMemcacheIDList(*req.ProjectId, *req.Region)
	})

	cmd.MarkFlagRequired("umem-id")
	return cmd
}

func restartMemcache(creq request.Common) (bool, []string) {
	req := creq.(*umem.RestartUMemcacheGroupRequest)
	block := ux.NewBlock()
	ux.Doc.Append(block)
	logs := make([]string, 0)
	logs = append(logs, fmt.Sprintf("api:RestartUMemcacheGroup, request:%v", base.ToQueryMap(req)))
	_, err := base.BizClient.RestartUMemcacheGroup(req)
	if err != nil {
		block.Append(base.ParseError(err))
		logs = append(logs, fmt.Sprintf("restart memcache[%s] failed: %s", *req.GroupId, base.ParseError(err)))
		return false, logs
	}
	poller := base.NewSpoller(describeMemcacheByID, base.Cxt.GetWriter())
	text := fmt.Sprintf("memcache[%s] is restarting", *req.GroupId)
	ret := poller.Sspoll(*req.GroupId, text, []string{status.UMEM_RUNNING, status.UMEM_FAIL}, block)
	if ret.Err != nil {
		block.Append(base.ParseError(err))
		logs = append(logs, ret.Err.Error())
	}
	if ret.Timeout {
		logs = append(logs, "poll memcache[%s] timeout", *req.GroupId)
	}
	return ret.Done, logs
}

func describeMemcacheByID(memcacheID string) (interface{}, error) {
	req := base.BizClient.NewDescribeUMemRequest()
	req.Protocol = sdk.String("memcache")
	req.ResourceId = &memcacheID

	resp, err := base.BizClient.DescribeUMem(req)
	if err != nil {
		return nil, err
	}
	if len(resp.DataSet) < 1 {
		return nil, fmt.Errorf(fmt.Sprintf("resource [%s] may not exist", memcacheID))
	}
	return &resp.DataSet[0], nil
}
func describeRedisByID(redisID string) (interface{}, error) {
	req := base.BizClient.NewDescribeUMemRequest()
	req.Protocol = sdk.String("redis")
	req.ResourceId = &redisID

	resp, err := base.BizClient.DescribeUMem(req)
	if err != nil {
		return nil, err
	}
	if len(resp.DataSet) < 1 {
		return nil, fmt.Errorf(fmt.Sprintf("resource [%s] may not exist", redisID))
	}
	return &resp.DataSet[0], nil
}

func getMemcacheIDList(project, region string) []string {
	req := base.BizClient.NewDescribeUMemcacheGroupRequest()
	req.ProjectId = &project
	req.Region = &region
	list := []string{}

	for limit, offset := 50, 0; ; offset += limit {
		req.Limit = sdk.Int(limit)
		req.Offset = sdk.Int(offset)
		resp, err := base.BizClient.DescribeUMemcacheGroup(req)
		if err != nil {
			fmt.Println(err)
			return nil
		}
		for _, ins := range resp.DataSet {
			list = append(list, fmt.Sprintf("%s/%s", ins.GroupId, ins.Name))
		}
		if offset+limit >= resp.TotalCount {
			break
		}
	}
	return list
}
