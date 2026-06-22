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
	"unicode/utf8"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/umem"
	"github.com/ucloud/ucloud-sdk-go/services/vpc"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/base"
	"github.com/ucloud/ucloud-cli/model/status"
	"github.com/ucloud/ucloud-cli/pkg/command"
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
	bindRegion(req, cmd)
	bindZoneEmpty(req, cmd)
	bindProjectID(req, cmd)
	bindOffset(req, flags)
	bindLimit(req, flags)
	req.Protocol = sdk.String("redis")

	command.SetCompletion(cmd, "umem-id", func() []string {
		return getRedisIDList(*req.ProjectId, *req.Region)
	})

	return cmd
}

// redisCreateParams holds the shared flag values for redis create commands.
type redisCreateParams struct {
	name       string
	password   string
	size       int
	region     string
	zone       string
	projectID  string
	chargeType string
	quantity   int
	group      string
	vpcID      string
	subnetID   string
	version    string
	blockCnt   int
	proxySize  int
}

// NewCmdRedisCreate ucloud redis create
func NewCmdRedisCreate(out io.Writer) *cobra.Command {
	var redisType string
	var p redisCreateParams
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create redis instance",
		Long:  "Create redis instance",
		Run: func(c *cobra.Command, args []string) {
			// Validate name，support Chinese name
			if l := utf8.RuneCountInString(p.name); l < 6 || l > 63 {
				fmt.Fprintln(out, "length of name should be between 6 and 63")
				return
			}
			// Validate password
			if p.password != "" {
				if l := len(p.password); l < 6 || l > 36 {
					fmt.Fprintln(out, "length of password should be between 6 and 36")
					return
				}
			}
			if err := fillDefaultVPCAndSubnet(&p.vpcID, &p.subnetID, p.projectID, p.region, p.zone); err != nil {
				fmt.Fprintln(out, err)
				return
			}
			switch redisType {
			case "master-replica":
				createMasterReplicaRedis(out, &p)
			case "distributed":
				createDistributedRedis(out, &p)
			default:
				fmt.Printf("unknow redis type[%s], it's should be 'master-replica' or 'distributed'\n", redisType)
			}
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringVar(&p.name, "name", "", "Required. Name of the redis to create. Range of the name length is [6,63]")
	flags.StringVar(&redisType, "type", "", "Required. Type of the redis. Accept values:'master-replica','distributed'")
	flags.IntVar(&p.size, "size-gb", 2, "Optional. Memory size. Default value 2GB. Unit GB")
	flags.StringVar(&p.version, "version", "6.0", "Optional. Version of redis. Accept values: '4.0', '5.0', '6.0', '7.0'")
	flags.StringVar(&p.vpcID, "vpc-id", "", "Optional. VPC ID. This field is required under VPC2.0. See 'ucloud vpc list'")
	flags.StringVar(&p.subnetID, "subnet-id", "", "Optional. Subnet ID. This field is required under VPC2.0. See 'ucloud subnet list'")
	flags.StringVar(&p.password, "password", "", "Optional. Password of redis to create. Range of the password length is [6,36] and the password can only contain letters and numbers")

	//distributed optional params
	flags.IntVar(&p.blockCnt, "block-cnt", 2, "Optional. Block count. Default value 2(for distributed redis type).")
	flags.IntVar(&p.proxySize, "proxy-size", 2, "Optional. Proxy size. Default value 2(for distributed redis type) Unit Core")

	bindRegionS(&p.region, cmd)
	bindZoneS(&p.zone, &p.region, cmd)
	bindProjectIDS(&p.projectID, cmd)
	flags.StringVar(&p.chargeType, "charge-type", "Month", "Optional. Enumeration value.'Year',pay yearly;'Month',pay monthly; 'Dynamic', pay hourly; 'Trial', free trial(need permission)")
	flags.IntVar(&p.quantity, "quantity", 1, "Optional. The duration of the instance. N years/months.")
	flags.StringVar(&p.group, "group", "", "Optional. Business group")

	command.SetFlagValues(cmd, "version", "4.0", "5.0", "6.0", "7.0")
	command.SetFlagValues(cmd, "type", "master-replica", "distributed")
	command.SetFlagValues(cmd, "charge-type", "Month", "Dynamic", "Year")
	command.SetCompletion(cmd, "vpc-id", func() []string {
		return getAllVPCIdNames(p.projectID, p.region)
	})
	command.SetCompletion(cmd, "subnet-id", func() []string {
		return getAllSubnetIDNames(p.vpcID, p.projectID, p.region)
	})

	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("type")

	return cmd
}

func createMasterReplicaRedis(out io.Writer, p *redisCreateParams) {
	req := base.BizClient.NewCreateURedisGroupRequest()
	req.Region = &p.region
	req.Zone = &p.zone
	req.ProjectId = &p.projectID
	req.Name = &p.name
	req.HighAvailability = sdk.String("enable")
	req.Size = &p.size
	req.Version = &p.version
	req.VPCId = &p.vpcID
	req.SubnetId = &p.subnetID
	req.ChargeType = &p.chargeType
	req.Quantity = &p.quantity
	req.Tag = &p.group
	if p.password != "" {
		req.Password = &p.password
	}

	resp, err := base.BizClient.CreateURedisGroup(req)
	if err != nil {
		base.HandleError(err)
		return
	}
	fmt.Fprintf(out, "redis[%s] created\n", resp.GroupId)
}

func createDistributedRedis(out io.Writer, p *redisCreateParams) {
	req := base.BizClient.NewCreateUMemSpaceRequest()
	req.Region = &p.region
	req.Zone = &p.zone
	req.ProjectId = &p.projectID
	req.Name = &p.name
	req.Protocol = sdk.String("redis")

	// Validate block-cnt
	if p.blockCnt <= 0 {
		fmt.Fprintln(out, "block-cnt should be greater than 0")
		return
	}

	// Validate size is divisible by block-cnt
	if p.size%p.blockCnt != 0 {
		fmt.Fprintf(out, "size-gb(%d) should be divisible by block-cnt(%d)\n", p.size, p.blockCnt)
		return
	}

	// Validate proxy-size is a multiple of 2
	if p.proxySize%2 != 0 {
		fmt.Fprintf(out, "proxy-size(%d) should be a multiple of 2\n", p.proxySize)
		return
	}

	req.BlockCnt = &p.blockCnt
	req.ProxySize = &p.proxySize
	req.Size = &p.size
	req.Version = &p.version
	req.VPCId = &p.vpcID
	req.SubnetId = &p.subnetID
	req.ChargeType = &p.chargeType
	req.Quantity = &p.quantity
	req.Tag = &p.group
	if p.password != "" {
		req.Password = &p.password
	}

	resp, err := base.BizClient.CreateUMemSpace(req)
	if err != nil {
		base.HandleError(err)
		return
	}
	fmt.Fprintf(out, "redis[%s] created\n", resp.SpaceId)
}

// redisDeleteParams holds the shared flag values for redis delete commands.
type redisDeleteParams struct {
	region    string
	zone      string
	projectID string
}

// NewCmdRedisDelete ucloud redis delete
func NewCmdRedisDelete(out io.Writer) *cobra.Command {
	var idNames []string
	var p redisDeleteParams
	cmd := &cobra.Command{
		Use:     "delete",
		Short:   "Delete redis instances",
		Long:    "Delete redis instances",
		Example: "ucloud redis delete --umem-id uredis-rl5xuxx/testcli1,uredis-xsdfa/testcli2",
		Run: func(c *cobra.Command, args []string) {
			for _, idname := range idNames {
				id := base.PickResourceID(idname)
				if strings.HasPrefix(id, "uredis") || strings.HasPrefix(id, "uhredis") || strings.HasPrefix(id, "uregionredis") {
					deleteMasterReplicaRedis(out, &p, id)
				} else if strings.HasPrefix(id, "udredis") {
					deleteDistributedRedis(out, &p, id)
				} else {
					fmt.Fprintf(out, "redis[%s] unknown id prefix, skip\n", idname)
				}
			}
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&idNames, "umem-id", nil, "Required. Resource ID of redis instances to delete")
	bindRegionS(&p.region, cmd)
	bindZoneS(&p.zone, &p.region, cmd)
	bindProjectIDS(&p.projectID, cmd)

	cmd.MarkFlagRequired("umem-id")

	command.SetCompletion(cmd, "umem-id", func() []string {
		return getRedisIDList(p.projectID, p.region)
	})

	return cmd
}

func deleteMasterReplicaRedis(out io.Writer, p *redisDeleteParams, id string) {
	req := base.BizClient.NewDeleteURedisGroupRequest()
	req.Region = &p.region
	req.Zone = &p.zone
	req.ProjectId = &p.projectID
	req.GroupId = &id
	_, err := base.BizClient.DeleteURedisGroup(req)
	if err != nil {
		base.HandleError(err)
		return
	}
	fmt.Fprintf(out, "redis[%s] deleted\n", id)
}

func deleteDistributedRedis(out io.Writer, p *redisDeleteParams, id string) {
	req := base.BizClient.NewDeleteUMemSpaceRequest()
	req.Region = &p.region
	req.Zone = &p.zone
	req.ProjectId = &p.projectID
	req.SpaceId = &id
	_, err := base.BizClient.DeleteUMemSpace(req)
	if err != nil {
		base.HandleError(err)
		return
	}
	fmt.Fprintf(out, "redis[%s] deleted\n", id)
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
	bindProjectID(req, cmd)
	bindRegion(req, cmd)
	bindZone(req, cmd)

	cmd.MarkFlagRequired("umem-id")
	command.SetCompletion(cmd, "umem-id", func() []string {
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
	ret := poller.Sspoll(*req.GroupId, text, []string{status.UMEM_RUNNING, status.UMEM_FAIL}, block, nil)
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
	bindRegion(req, cmd)
	bindZoneEmpty(req, cmd)
	bindProjectID(req, cmd)
	bindOffset(req, flags)
	bindLimit(req, flags)

	return cmd
}

// NewCmdMemcacheCreate ucloud memcache create
func NewCmdMemcacheCreate(out io.Writer) *cobra.Command {
	req := base.BizClient.NewCreateUMemcacheGroupRequest()
	var region, zone, projectID string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create memcache instance",
		Long:  "Create memcache instance",
		Run: func(c *cobra.Command, args []string) {
			if *req.Size > 32 || *req.Size < 1 {
				fmt.Fprintln(out, "size-gb should be between 1 and 32")
				return
			}
			if err := fillDefaultVPCAndSubnet(req.VPCId, req.SubnetId, *req.ProjectId, *req.Region, *req.Zone); err != nil {
				fmt.Fprintln(out, err)
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
	bindRegionS(&region, cmd)
	bindZoneS(&zone, &region, cmd)
	bindProjectIDS(&projectID, cmd)
	req.ChargeType = flags.String("charge-type", "Month", "Optional. Enumeration value.'Year',pay yearly;'Month',pay monthly; 'Dynamic', pay hourly; 'Trial', free trial(need permission)")
	req.Quantity = flags.Int("quantity", 1, "Optional. The duration of the instance. N years/months.")
	req.Tag = flags.String("group", "", "Optional. Business group")

	// Set region/zone/projectID to request after flag parsing
	req.Region = &region
	req.Zone = &zone
	req.ProjectId = &projectID

	command.SetFlagValues(cmd, "size-gb", "1", "2", "4", "8", "16", "32")
	command.SetFlagValues(cmd, "charge-type", "Month", "Dynamic", "Year")
	command.SetCompletion(cmd, "vpc-id", func() []string {
		return getAllVPCIdNames(projectID, region)
	})
	command.SetCompletion(cmd, "subnet-id", func() []string {
		return getAllSubnetIDNames(*req.VPCId, projectID, region)
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
	bindProjectID(req, cmd)
	bindRegion(req, cmd)
	bindZoneEmpty(req, cmd)

	cmd.MarkFlagRequired("umem-id")

	command.SetCompletion(cmd, "umem-id", func() []string {
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
	bindRegion(req, cmd)
	bindZone(req, cmd)
	bindProjectID(req, cmd)

	command.SetCompletion(cmd, "umem-id", func() []string {
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
	ret := poller.Sspoll(*req.GroupId, text, []string{status.UMEM_RUNNING, status.UMEM_FAIL}, block, nil)
	if ret.Err != nil {
		block.Append(base.ParseError(err))
		logs = append(logs, ret.Err.Error())
	}
	if ret.Timeout {
		logs = append(logs, "poll memcache[%s] timeout", *req.GroupId)
	}
	return ret.Done, logs
}

func describeMemcacheByID(memcacheID string, commonBase *request.CommonBase) (interface{}, error) {
	req := base.BizClient.NewDescribeUMemRequest()
	if commonBase != nil {
		req.CommonBase = *commonBase
	}
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
func describeRedisByID(redisID string, commonBase *request.CommonBase) (interface{}, error) {
	req := base.BizClient.NewDescribeUMemRequest()
	if commonBase != nil {
		req.CommonBase = *commonBase
	}
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

func fillDefaultVPCAndSubnet(vpcID, subnetID *string, projectID, region, zone string) error {
	if *vpcID != "" && *subnetID != "" {
		return nil
	}
	vpcs, err := getAllVPCIns(projectID, region)
	if err != nil {
		return fmt.Errorf("failed to get vpc list: %s", err)
	}
	if len(vpcs) == 0 {
		return fmt.Errorf("no vpc found in region[%s], please specify --vpc-id and --subnet-id", region)
	}

	// Find the default VPC
	var defaultVPC *vpc.VPCInfo
	for i := range vpcs {
		if vpcs[i].VPCType == "DefaultVPC" {
			defaultVPC = &vpcs[i]
			break
		}
	}
	// Fallback to the first VPC if no DefaultVPC found
	if defaultVPC == nil {
		defaultVPC = &vpcs[0]
	}

	if *vpcID == "" {
		*vpcID = defaultVPC.VPCId
	}

	if *subnetID == "" {
		subnets, err := getAllSubnets(*vpcID, projectID, region)
		if err != nil {
			return fmt.Errorf("failed to get subnet list: %s", err)
		}
		if len(subnets) == 0 {
			return fmt.Errorf("no subnet found in vpc[%s], please specify --subnet-id", *vpcID)
		}
		// Filter subnets by zone if specified
		if zone != "" {
			for _, sn := range subnets {
				if sn.Zone == zone {
					*subnetID = sn.SubnetId
					return nil
				}
			}
		}
		// Fallback to the first subnet
		*subnetID = subnets[0].SubnetId
	}

	return nil
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
