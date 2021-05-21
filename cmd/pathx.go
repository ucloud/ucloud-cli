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
	"strings"

	"github.com/spf13/cobra"

	ppathx "github.com/ucloud/ucloud-sdk-go/private/services/pathx"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"
	uerr "github.com/ucloud/ucloud-sdk-go/ucloud/error"

	"github.com/ucloud/ucloud-cli/base"
)

//NewCmdPathx ucloud pathx
func NewCmdPathx() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pathx",
		Short: "Manipulate uga and upath instances",
		Long:  "Manipulate uga and upath instances",
	}
	cmd.AddCommand(NewCmdUGA())
	cmd.AddCommand(NewCmdUpath())
	return cmd
}

//NewCmdUpath ucloud pathx upath
func NewCmdUpath() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upath",
		Short: "List pathx upath instances",
		Long:  "List pathx upath instances",
	}
	out := base.Cxt.GetWriter()
	cmd.AddCommand(NewCmdUpathList(out))
	return cmd
}

type upathRow struct {
	ResourceID      string
	UPathName       string
	AcceleratedPath string
	BoundUGA        string
}

//NewCmdUpathList ucloud pathx upath list
func NewCmdUpathList(out io.Writer) *cobra.Command {
	req := base.BizClient.PrivatePathxClient.NewDescribeUPathRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list upath instances",
		Long:  "list upath instances",
		Run: func(c *cobra.Command, args []string) {
			resp, err := base.BizClient.PrivatePathxClient.DescribeUPath(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			list := make([]upathRow, 0)
			for _, ins := range resp.UPathSet {
				row := upathRow{
					ResourceID:      ins.UPathId,
					UPathName:       ins.Name,
					AcceleratedPath: fmt.Sprintf("%s->%s %dM", ins.LineFromName, ins.LineToName, ins.Bandwidth),
				}
				ids := []string{}
				for _, ga := range ins.UGAList {
					ids = append(ids, ga.UGAId)
				}
				row.BoundUGA = strings.Join(ids, ",")
				list = append(list, row)
			}
			base.PrintList(list, out)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	bindProjectID(req, flags)
	req.UPathId = flags.String("upath-id", "", "Optional. Resource ID of upath instance to list")

	return cmd
}

//NewCmdUGA ucloud uga
func NewCmdUGA() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uga",
		Short: "Create,list,update and delete pathx uga instances",
		Long:  `Create,list,update and delete pathx uga instances`,
	}

	out := base.Cxt.GetWriter()
	cmd.AddCommand(NewCmdUGAList(out))
	cmd.AddCommand(NewCmdUGADescribe(out))
	cmd.AddCommand(NewCmdUGACreate(out))
	cmd.AddCommand(NewCmdUGADelete(out))
	cmd.AddCommand(NewCmdUGAAddPort(out))
	cmd.AddCommand(NewCmdUGARemovePort(out))

	return cmd
}

//UGARow 表格行
type UGARow struct {
	ResourceID      string
	UGAName         string
	CName           string
	Origin          string
	AcceleratedPath string
}

var protocols = []string{"tcp", "udp"}

//NewCmdUGAList ucloud uga list
func NewCmdUGAList(out io.Writer) *cobra.Command {
	req := base.BizClient.PrivatePathxClient.NewDescribeUGAInstanceRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list uga instances",
		Long:  "list uga instances",
		Run: func(c *cobra.Command, args []string) {
			*req.UGAId = base.PickResourceID(*req.UGAId)
			resp, err := base.BizClient.PrivatePathxClient.DescribeUGAInstance(req)
			if err != nil {
				base.HandleError(err)
				return
			}

			list := make([]UGARow, 0)
			for _, ins := range resp.UGAList {
				row := UGARow{
					ResourceID: ins.UGAId,
					UGAName:    ins.UGAName,
					CName:      ins.CName,
					Origin:     fmt.Sprintf("%s%s", strings.Join(ins.IPList, ","), ins.Domain),
				}
				row.AcceleratedPath = getUpathStr(ins.UPathSet)
				list = append(list, row)
			}
			base.PrintList(list, out)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.UGAId = flags.String("uga-id", "", "Optional. Resource ID of uga instance")
	bindProjectID(req, flags)

	return cmd
}

func getUpathStr(list []ppathx.UPathSet) string {
	paths := make([]string, 0)
	for _, p := range list {
		paths = append(paths, fmt.Sprintf("%s->%s %dM", p.LineFromName, p.LineToName, p.Bandwidth))
	}
	return strings.Join(paths, "\n")
}

func getOutIPStr(list []ppathx.OutPublicIpInfo) string {
	strs := make([]string, 0)
	for _, p := range list {
		strs = append(strs, fmt.Sprintf("%s %s", p.IP, base.RegionLabel[p.Area]))
	}
	return strings.Join(strs, "\n")
}

func getPortStr(list []ppathx.UGAATask) string {
	strs := make([]string, 0)
	for _, t := range list {
		strs = append(strs, fmt.Sprintf("%s %d", t.Protocol, t.Port))
	}
	return strings.Join(strs, "\n")
}

//NewCmdUGADescribe ucloud uga describe
func NewCmdUGADescribe(out io.Writer) *cobra.Command {
	req := base.BizClient.PrivatePathxClient.NewDescribeUGAInstanceRequest()
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Display detail informations about uga instances",
		Long:  "Display detail informations about uga instances",
		Run: func(c *cobra.Command, args []string) {
			*req.UGAId = base.PickResourceID(*req.UGAId)
			resp, err := base.BizClient.PrivatePathxClient.DescribeUGAInstance(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			if len(resp.UGAList) != 1 {
				base.HandleError(fmt.Errorf("uga[%s] may not exist", *req.UGAId))
				return
			}

			ins := resp.UGAList[0]
			list := []base.DescribeTableRow{
				{"ResourceID", ins.UGAId},
				{"UGAName", ins.UGAName},
				{"Origin", fmt.Sprintf("%s%s", ins.Domain, strings.Join(ins.IPList, ","))},
				{"CName", ins.CName},
				{"AcceleratedPath", getUpathStr(ins.UPathSet)},
				{"OutIP", getOutIPStr(ins.OutPublicIpList)},
				{"Port", getPortStr(ins.TaskSet)},
			}
			base.PrintList(list, out)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.UGAId = flags.String("uga-id", "", "Required. Resource ID of uga instance")
	bindProjectID(req, flags)

	cmd.MarkFlagRequired("uga-id")
	flags.SetFlagValuesFunc("uga-id", func() []string {
		return getUGAIDList(*req.ProjectId)
	})

	return cmd
}

func formatPortList(userPorts []string) ([]string, error) {
	portList := make([]string, 0)
	for _, port := range userPorts {
		if strings.Contains(port, "-") {
			portRange := strings.Split(port, "-")
			if len(portRange) != 2 {
				return nil, fmt.Errorf("port %s is invalid, it's pattern should be like 3000-3100", port)
			}
			min, err := strconv.Atoi(portRange[0])
			if err != nil {
				return nil, fmt.Errorf("parse port failed: %v", err)
			}
			max, err := strconv.Atoi(portRange[1])
			if err != nil {
				return nil, fmt.Errorf("parse port failed: %v", err)
			}

			for i := min; i <= max; i++ {
				portList = append(portList, strconv.Itoa(i))
			}
		} else {
			portList = append(portList, port)
		}
	}
	return portList, nil
}

//NewCmdUGACreate ucloud uga create
func NewCmdUGACreate(out io.Writer) *cobra.Command {
	var protocol string
	var ports, lines []string
	req := base.BizClient.PrivatePathxClient.NewCreateUGAInstanceRequest()
	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Create uga instance",
		Long:    "Create uga instance",
		Example: "ucloud pathx uga create --name testcli1 --protocol tcp --origin-location 中国 --origin-domain lixiaojun.xyz --upath-id upath-auvfexxx/test_0 --port 80-90,100,110-115",
		Run: func(c *cobra.Command, args []string) {
			if *req.IPList == "" && *req.Domain == "" {
				fmt.Fprintln(out, "origin-ip and origin-domain can not be both empty")
				return
			}

			portList, err := formatPortList(ports)
			if err != nil {
				base.HandleError(err)
				return
			}

			switch strings.ToLower(protocol) {
			case "tcp":
				req.TCP = portList
			case "udp":
				req.UDP = portList
			case "http":
				req.HTTP = portList
			case "https":
				req.HTTPS = portList
			default:
				fmt.Fprintf(out, "protocol should be one of %s, received:%s\n", strings.Join(protocols, ","), protocol)
			}

			resp, err := base.BizClient.PrivatePathxClient.CreateUGAInstance(req)
			if err != nil {
				if uErr, ok := err.(uerr.Error); ok && uErr.Code() == 33756 {
					fmt.Fprintf(out, "The number of ports added exceeds the limit(50). We recommend that you could reduce the number of ports, then create an uga instance, \nand then add the remaining ports by executing 'ucloud pathx uga add-port --protocol %s --uga-id <uga-id> --port <PortList>'\n", protocol)
				}
				return
			}

			fmt.Fprintf(out, "uga[%s] created\n", resp.UGAId)

			for _, path := range lines {
				p := base.PickResourceID(path)
				bindReq := base.BizClient.PrivatePathxClient.NewUGABindUPathRequest()
				bindReq.ProjectId = req.ProjectId
				bindReq.UGAId = sdk.String(resp.UGAId)
				bindReq.UPathId = &p
				_, err := base.BizClient.PrivatePathxClient.UGABindUPath(bindReq)
				if err != nil {
					fmt.Fprintf(out, "bind uga[%s] and upath[%s] failed: %v\n", resp.UGAId, p, err)
				} else {
					fmt.Fprintf(out, "bound uga[%s] and upath[%s]\n", resp.UGAId, p)
				}
			}
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	bindProjectID(req, flags)
	req.Name = flags.String("name", "", "Required. Name of uga instance to create")
	req.IPList = flags.String("origin-ip", "", "Required if origin-domain is empty. IP address of origin. multiple IP address separated by ','")
	req.Domain = flags.String("origin-domain", "", "Required if origin-ip is empty.")
	req.Location = flags.String("origin-location", "", "Required. Location of origin ip or domain. accpet valeus:'中国','洛杉矶','法兰克福','中国香港','雅加达','孟买','东京','莫斯科','新加坡','曼谷','中国台北','华盛顿','首尔'")
	flags.StringVar(&protocol, "protocol", "", fmt.Sprintf("Required. accept values: %s", strings.Join(protocols, ",")))
	flags.StringSliceVar(&ports, "port", nil, "Required. Single port or port range, separated by ',', for example 80,3000-3010")
	flags.StringSliceVar(&lines, "upath-id", nil, "Required. Accelerated path to bind with the uga instance to create. multiple upath-id separated by ','; see 'ucloud pathx upath list")

	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("origin-location")
	cmd.MarkFlagRequired("protocol")
	cmd.MarkFlagRequired("port")
	cmd.MarkFlagRequired("upath-id")

	flags.SetFlagValues("origin-location", "中国", "洛杉矶", "法兰克福", "中国香港", "雅加达", "孟买", "东京", "莫斯科", "新加坡", "曼谷", "中国台北", "华盛顿", "首尔")
	flags.SetFlagValues("protocol", protocols...)
	flags.SetFlagValuesFunc("upath-id", func() []string {
		return getUpathIDList(*req.ProjectId)
	})

	return cmd
}

//NewCmdUGADelete ucloud uga delete
func NewCmdUGADelete(out io.Writer) *cobra.Command {
	idNames := []string{}
	req := base.BizClient.PrivatePathxClient.NewDeleteUGAInstanceRequest()
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete uga instances",
		Long:  "Delete uga instances",
		Run: func(c *cobra.Command, args []string) {
			for _, idname := range idNames {
				id := base.PickResourceID(idname)
				req.UGAId = &id
				_, err := base.BizClient.PrivatePathxClient.DeleteUGAInstance(req)
				if err != nil {
					base.HandleError(err)
				} else {
					fmt.Fprintf(out, "uga[%s] deleted\n", id)
				}
			}
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	bindProjectID(req, flags)
	flags.StringSliceVar(&idNames, "uga-id", nil, "Required. Resource ID of uga instances to delete. Multiple resource ids separated by comma")

	cmd.MarkFlagRequired("uga-id")
	flags.SetFlagValuesFunc("uga-id", func() []string {
		return getUGAIDList(*req.ProjectId)
	})

	return cmd
}

//NewCmdUGAAddPort ucloud pathx uga add-port
func NewCmdUGAAddPort(out io.Writer) *cobra.Command {
	var ports []string
	var protocol string
	req := base.BizClient.NewAddUGATaskRequest()
	cmd := &cobra.Command{
		Use:   "add-port",
		Short: "Add port for uga instance",
		Long:  "Add port for uga instance",
		Run: func(c *cobra.Command, args []string) {
			portList, err := formatPortList(ports)
			if err != nil {
				base.HandleError(err)
				return
			}

			switch strings.ToLower(protocol) {
			case "tcp":
				req.TCP = portList
			case "udp":
				req.UDP = portList
			case "http":
				req.HTTP = portList
			case "https":
				req.HTTPS = portList
			default:
				fmt.Fprintf(out, "protocol should be one of %s, received:%s\n", strings.Join(protocols, ","), protocol)
			}

			*req.UGAId = base.PickResourceID(*req.UGAId)
			_, err = base.BizClient.AddUGATask(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			fmt.Fprintf(out, "port %v added\n", ports)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	bindProjectID(req, flags)
	req.UGAId = flags.String("uga-id", "", "Required. Resource ID of uga instance to add port")
	flags.StringVar(&protocol, "protocol", "", fmt.Sprintf("Required. accept values: %s", strings.Join(protocols, ",")))
	flags.StringSliceVar(&ports, "port", nil, "Required. Single port or port range, separated by ',', for example 80,3000-3010")

	cmd.MarkFlagRequired("protocol")
	cmd.MarkFlagRequired("uga-id")
	cmd.MarkFlagRequired("port")

	flags.SetFlagValues("protocol", protocols...)
	flags.SetFlagValuesFunc("uga-id", func() []string {
		return getUGAIDList(*req.ProjectId)
	})

	return cmd
}

//NewCmdUGARemovePort ucloud pathx uga delete-port
func NewCmdUGARemovePort(out io.Writer) *cobra.Command {
	var ports []string
	var protocol string
	req := base.BizClient.NewDeleteUGATaskRequest()
	cmd := &cobra.Command{
		Use:   "delete-port",
		Short: "Delete port for uga instance",
		Long:  "Delete port for uga instance",
		Run: func(c *cobra.Command, args []string) {
			portList, err := formatPortList(ports)
			if err != nil {
				base.HandleError(err)
				return
			}

			switch strings.ToLower(protocol) {
			case "tcp":
				req.TCP = portList
			case "udp":
				req.UDP = portList
			case "http":
				req.HTTP = portList
			case "https":
				req.HTTPS = portList
			default:
				fmt.Fprintf(out, "protocol should be one of %s, received:%s\n", strings.Join(protocols, ","), protocol)
			}

			*req.UGAId = base.PickResourceID(*req.UGAId)
			_, err = base.BizClient.DeleteUGATask(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			fmt.Fprintf(out, "port %v deleted\n", ports)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	bindProjectID(req, flags)
	req.UGAId = flags.String("uga-id", "", "Required. Resource ID of uga instance to delete port")
	flags.StringVar(&protocol, "protocol", "", fmt.Sprintf("Required. accept values: %s", strings.Join(protocols, ",")))
	flags.StringSliceVar(&ports, "port", nil, "Required. Single port or port range, separated by ',', for example 80,3000-3010")

	cmd.MarkFlagRequired("protocol")
	cmd.MarkFlagRequired("uga-id")
	cmd.MarkFlagRequired("port")

	flags.SetFlagValues("protocol", protocols...)
	flags.SetFlagValuesFunc("uga-id", func() []string {
		return getUGAIDList(*req.ProjectId)
	})

	return cmd
}

func getUGAList(project string) ([]ppathx.UGAAInfo, error) {
	req := base.BizClient.PrivatePathxClient.NewDescribeUGAInstanceRequest()
	req.ProjectId = &project
	resp, err := base.BizClient.PrivatePathxClient.DescribeUGAInstance(req)
	if err != nil {
		return nil, err
	}
	return resp.UGAList, nil
}

func getUGAIDList(project string) []string {
	list, err := getUGAList(project)
	if err != nil {
		base.LogError(fmt.Sprintf("getUDGAIDList filed:%v", err))
		return nil
	}
	strs := make([]string, 0)
	for _, ins := range list {
		strs = append(strs, fmt.Sprintf("%s/%s", ins.UGAId, ins.UGAName))
	}
	return strs
}

func getUpathList(project string) ([]ppathx.UPathInfo, error) {
	req := base.BizClient.PrivatePathxClient.NewDescribeUPathRequest()
	req.ProjectId = &project
	resp, err := base.BizClient.PrivatePathxClient.DescribeUPath(req)
	if err != nil {
		return nil, err
	}
	return resp.UPathSet, nil
}

func getUpathIDList(project string) []string {
	list, err := getUpathList(project)
	if err != nil {
		base.LogError(fmt.Sprintf("getUpathIDList failed:%v", err))
		return nil
	}
	strs := make([]string, 0)
	for _, ins := range list {
		strs = append(strs, fmt.Sprintf("%s/%s", ins.UPathId, ins.Name))
	}
	return strs
}
