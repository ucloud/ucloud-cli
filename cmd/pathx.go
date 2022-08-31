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

	ppathx "github.com/ucloud/ucloud-sdk-go/private/services/pathx"
	"github.com/ucloud/ucloud-sdk-go/services/pathx"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"
	uerr "github.com/ucloud/ucloud-sdk-go/ucloud/error"

	"github.com/spf13/cobra"
	"github.com/ucloud/ucloud-cli/base"
	"github.com/ucloud/ucloud-cli/ux"
)

//NewCmdPathx ucloud pathx
func NewCmdPathx() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pathx",
		Short: "Manipulate uga and upath instances",
		Long:  "Manipulate uga and upath instances",
	}
	out := base.Cxt.GetWriter()
	cmd.AddCommand(NewCmdUGA())
	cmd.AddCommand(NewCmdUpath())
	cmd.AddCommand(NewCmdUGA3Create(out))
	cmd.AddCommand(NewCmdUGA3Delete(out))
	cmd.AddCommand(NewCmdUGA3Modify(out))
	cmd.AddCommand(NewCmdUGA3List(out))
	cmd.AddCommand(NewCmdPathxPrice(out))
	cmd.AddCommand(NewCmdPathxArea(out))

	return cmd
}

// create pathx instance
func NewCmdUGA3Create(out io.Writer) *cobra.Command {
	createPathxReq := base.BizClient.NewCreateUGA3InstanceRequest()
	createPathxPortReq := base.BizClient.NewCreateUGA3PortRequest()
	spinner := ux.NewDotSpinner(out)
	var ports, originPorts []string
	protocol := "tcp"
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create the pathx resource and port",
		Long:  "Create global unified access acceleration configuration item",
		Example: "ucloud pathx create --bandwidth 10 --area-code DXB" +
			"--charge-type Month --quantity 4 --accel Global --origin-ip 110.111.111.111" +
			"--protocol TCP --port 30654 --origin-port 30564",
		Run: func(cmd *cobra.Command, args []string) {
			spinner.Start("The pathx resource creating")
			if *createPathxReq.OriginIPList == "" && *createPathxReq.OriginDomain == "" {
				spinner.Fail(fmt.Errorf("The origin-ip and origin-domain cannot be empty at the same time"))
				return
			}
			portIntList := make([]int, 0)
			originPortIntList := make([]int, 0)
			if len(ports) > 0 || len(originPorts) > 0 {
				if len(ports) == 0 {
					spinner.Fail(fmt.Errorf("The port cannot be empty."))
					return
				} else if len(originPorts) == 0 {
					spinner.Fail(fmt.Errorf("The origin-port cannot be empty."))
					return
				}
				if strings.EqualFold(protocol, "UDP") {
					spinner.Fail(fmt.Errorf("The udp protocol is temporarily not supported for create"))
					return
				} else if !strings.EqualFold(protocol, "TCP") &&
					!strings.EqualFold(protocol, "UDP") {
					spinner.Fail(fmt.Errorf("The value of protocol input error,please input 'TCP' or 'UDP',and the value entered is not case sensitive"))
					return
				}
				tcpPortList, err := formatPortList(ports)
				if err != nil {
					spinner.Fail(err)
					return
				}
				// tcpPorts convert to []int
				for _, tcpPort := range tcpPortList {
					port, _ := strconv.Atoi(tcpPort)
					portIntList = append(portIntList, port)
				}
				rsTcpPortList, err := formatPortList(originPorts)
				if err != nil {
					spinner.Fail(err)
					return
				}
				// rsTcpPorts convert to []int
				for _, rsTcpPort := range rsTcpPortList {
					rsPort, _ := strconv.Atoi(rsTcpPort)
					originPortIntList = append(originPortIntList, rsPort)
				}
				if len(portIntList) != len(originPortIntList) {
					spinner.Fail(fmt.Errorf("The number of port must be consistent with the number of origin-port."))
					return
				} else if len(portIntList) >= 10 {
					spinner.Fail(fmt.Errorf("The number of port cannot greater than or equals to 10"))
					return
				}
			}
			if strings.EqualFold(*createPathxReq.ChargeType, "Month") {
				*createPathxReq.Quantity = 0
			} else if *createPathxReq.Quantity <= 0 {
				spinner.Fail(fmt.Errorf("If the value of charge-type is 'Year' or 'Hour',the value of quantity must be greater than 0"))
				return
			}

			switch strings.ToLower(*createPathxReq.ChargeType) {
			case "hour":
				*createPathxReq.ChargeType = "Dynamic"
			case "month":
				*createPathxReq.ChargeType = "Month"
			case "year":
				*createPathxReq.ChargeType = "Year"
			}
			// post create pathx resource
			createUGA3InstanceResp, err := base.BizClient.CreateUGA3Instance(createPathxReq)
			if err != nil {
				spinner.Fail(err)
				return
			}
			if createUGA3InstanceResp == nil || createUGA3InstanceResp.InstanceId == "" ||
				&createUGA3InstanceResp.InstanceId == nil {
				spinner.Fail(fmt.Errorf("An unknown error occurred and could not be created successfully."))
				return
			}
			spinner.Stop()

			// CreatePathxPort
			if len(portIntList) > 0 && len(originPortIntList) > 0 {
				createPathxPortReq.InstanceId = &createUGA3InstanceResp.InstanceId
				createPathxPortReq.SetRegionRef(createPathxReq.GetRegionRef())
				createPathxPortReq.SetProjectIdRef(createPathxReq.GetProjectIdRef())
				createPathxPortReq.SetZoneRef(createPathxReq.GetZoneRef())
				spinner.Start("The pathx port creating")
				// Temporary support tcp protocol
				if strings.EqualFold(protocol, "TCP") {
					createPathxPortReq.TCP = portIntList
					createPathxPortReq.TCPRS = originPortIntList
				}
				_, err := base.BizClient.CreateUGA3Port(createPathxPortReq)
				if err != nil {
					spinner.Fail(err)
					return
				}
				spinner.Stop()
			}

			fmt.Fprintf(out, "The resource is created, and the resource ID is: %s\n", createUGA3InstanceResp.InstanceId)
		},
	}
	flags := createCmd.Flags()
	flags.SortFlags = false

	bindProjectID(createPathxReq, flags)
	bindRegion(createPathxReq, flags)
	bindZone(createPathxReq, flags)

	createPathxReq.Bandwidth = flags.Int("bandwidth", 0,
		"Required. Shared bandwidth of the resource")
	flags.String("area-code", "",
		"Optional. When it is empty,the nearest zone will be selected based on the origin-domain and origin-ip. "+
			"Acceptable values:'BKK'(曼谷),'DXB'(迪拜),'FRA'(法兰克福),'SGN'(胡志明市),'HKG'(香港),'CGK'(雅加达),'LOS'(拉各斯),'LHR'(伦敦),'LAX'(洛杉矶),"+
			"'MNL'(马尼拉),'DME'(莫斯科),'BOM'(孟买),'MSP'(圣保罗),'ICN'(首尔),'PVG'(上海),'SIN'(新加坡),'NRT'(东京),'IAD'(华盛顿),'TPE'(台北)")

	createPathxReq.ChargeType = flags.String("charge-type", "",
		"Optional. Payment method,its value is not case sensitive,acceptable values:'Year',pay yearly;'Month',pay monthly;'Hour', pay hourly")
	createPathxReq.Quantity = flags.Int("quantity", 1,
		"Optional. The duration of the pathx resource, the value cannot be less than or equal to 0. N years/months")
	createPathxReq.AccelerationArea = flags.String("accel", "",
		"Optional. The default value is 'Global'(全球). "+
			"Other acceptable values:'AP'(亚太);'EU'(欧洲);'ME'(中东);'OA'(大洋洲);'AF'(非洲);'NA'(北美洲);'SA'(南美洲)")
	createPathxReq.OriginIPList = flags.String("origin-ip", "",
		"Optional. But when the origin-domain is empty,it cannot be empty. If multiple values exist,please split by ','. For example '0.0.0.0,110.110.100.100'")
	createPathxReq.OriginDomain = flags.String("origin-domain", "",
		"Optional. But when the origin-ip is empty,it cannot be empty")

	flags.StringSliceVar(&ports, "port", nil,
		"Optional. Disable 65123 port,the port can be multiple,please split by ',' for example 80,3000-3010. "+
			"The number of port must be consistent with the number of origin-port,and the number cannot greater than or equals to 10")
	flags.StringSliceVar(&originPorts, "origin-port", nil,
		"Optional. The origin-port can be multiple,please split by ',' for example 80,3000-3010."+
			"The number of origin-port must be consistent with the number of port")
	flags.StringVar(&protocol, "protocol", "TCP", "Its values can be TCP and UDP, but currently only supports TCP")

	createCmd.MarkFlagRequired("bandwidth")
	flags.SetFlagValues("area-code",
		"BKK", "DXB", "FRA", "SGN", "HKG", "CGK", "LOS", "LHR", "LAX", "MNL", "DME", "BOM", "MSP", "ICN", "PVG", "SIN", "NRT", "IAD", "TPE")
	flags.SetFlagValues("charge-type", "Month", "Year", "Hour")
	flags.SetFlagValues("accel", "Global", "AP", "EU", "ME", "OA", "AF", "NA", "SA")
	flags.SetFlagValues("protocol", "TCP", "UDP")
	return createCmd
}

// delete pathx instance
func NewCmdUGA3Delete(out io.Writer) *cobra.Command {
	deleteUga3Req := base.BizClient.NewDeleteUGA3InstanceRequest()
	deleteUga3PortReq := base.BizClient.NewDeleteUGA3PortRequest()
	spinner := ux.NewDotSpinner(out)
	var yes *bool
	var instanceId string
	removeCmd := &cobra.Command{
		Use:     "delete",
		Short:   "Delete the pathx resource and port",
		Long:    "Delete the pathx resource and port",
		Example: "ucloud pathx delete --id uga3-xxx",
		Run: func(cmd *cobra.Command, args []string) {
			if !*yes {
				sure, err := ux.Prompt("Are you sure you want to delete this resource ?")
				if err != nil {
					base.Cxt.Println(err)
					return
				}
				if !sure {
					return
				}
			}
			spinner.Start(fmt.Sprintf("Starting delete the pathx[%s] resource port", instanceId))
			deleteUga3PortReq.InstanceId = &instanceId
			_, deletePortErr := base.BizClient.DeleteUGA3Port(deleteUga3PortReq)
			if deletePortErr != nil {
				spinner.Fail(deletePortErr)
				return
			}
			spinner.Stop()

			spinner.Start(fmt.Sprintf("Starting delete the pathx[%s] resource", instanceId))
			deleteUga3Req.InstanceId = &instanceId
			deleteUga3Req.SetProjectIdRef(deleteUga3PortReq.GetProjectIdRef())
			deleteUga3Req.SetRegionRef(deleteUga3PortReq.GetRegionRef())
			deleteUga3Req.SetZoneRef(deleteUga3PortReq.GetZoneRef())
			_, err := base.BizClient.DeleteUGA3Instance(deleteUga3Req)
			if err != nil {
				spinner.Fail(err)
				return
			}
			spinner.Stop()
		},
	}
	flags := removeCmd.Flags()
	flags.SortFlags = false
	flags.StringVar(&instanceId, "id", "",
		"Required. It is the resource ID of pathx, and the deletion will be performed according to this")

	bindProjectID(deleteUga3PortReq, flags)
	bindRegion(deleteUga3PortReq, flags)
	bindZone(deleteUga3PortReq, flags)

	removeCmd.MarkFlagRequired("id")
	yes = removeCmd.Flags().BoolP("yes", "y", false, "Optional. Do not prompt for confirmation.")
	flags.SetFlagValuesFunc("id", func() []string {
		return getPathxList(*deleteUga3PortReq.ProjectId, *deleteUga3PortReq.Region, *deleteUga3PortReq.Zone)
	})
	return removeCmd
}

func getPathxList(project, region, zone string) []string {
	getInstanceReq := base.BizClient.NewDescribeUGA3InstanceRequest()
	getInstanceReq.ProjectId = sdk.String(project)
	getInstanceReq.Region = sdk.String(region)
	getInstanceReq.Zone = sdk.String(zone)
	getInstanceReq.Limit = sdk.Int(50)
	resp, err := base.BizClient.DescribeUGA3Instance(getInstanceReq)
	if err != nil {
		base.HandleError(err)
		return nil
	}
	list := make([]string, 0)
	for _, item := range resp.ForwardInstanceInfos {
		list = append(list, item.InstanceId)
	}
	return list
}

// modify UGA3 instance
func NewCmdUGA3Modify(out io.Writer) *cobra.Command {
	modifyBandwidthReq := base.BizClient.NewModifyUGA3BandwidthRequest()
	modifyOriginInfoReq := base.BizClient.NewModifyUGA3OriginInfoRequest()
	modifyInstanceReq := base.BizClient.NewModifyUGA3InstanceRequest()
	modifyPortReq := base.BizClient.NewModifyUGA3PortRequest()

	spinner := ux.NewDotSpinner(out)
	var tcpPorts, rsTcpPorts []string
	var instanceId string
	protocol := "TCP"
	modifyCmd := &cobra.Command{
		Use:     "modify",
		Short:   "Modify the pathx associated information. Example bandwidth or origin information or resource information",
		Long:    "Support modify bandwidth,origin information,resource information,port",
		Example: "ucloud pathx modify --id uga3-xxx --bandwidth 1 --origin-ip 127.0.0.1 --name Pathx测试 --remark 加速资源 --protocol TCP --port 30010 --origin-port 39999",
		Run: func(cmd *cobra.Command, args []string) {
			modifyBandwidthReq.InstanceId = &instanceId
			modifyInstanceReq.InstanceId = &instanceId
			modifyOriginInfoReq.InstanceId = &instanceId
			modifyPortReq.InstanceId = &instanceId
			if *modifyBandwidthReq.Bandwidth != 0 {
				spinner.Start(fmt.Sprintf("Starting modify the pathx[%s] bandwidth", instanceId))
				if *modifyBandwidthReq.Bandwidth < 1 || *modifyBandwidthReq.Bandwidth > 100 {
					spinner.Fail(fmt.Errorf("The value of bandwidth size cannot be less than 1 and cannot be greater than 100"))
					return
				}
				modifyBandwidthReq.SetProjectIdRef(modifyInstanceReq.GetProjectIdRef())
				modifyBandwidthReq.SetRegionRef(modifyInstanceReq.GetRegionRef())
				modifyBandwidthReq.SetZoneRef(modifyInstanceReq.GetZoneRef())
				_, err := base.BizClient.ModifyUGA3Bandwidth(modifyBandwidthReq)
				if err != nil {
					spinner.Fail(err)
					return
				}
				spinner.Stop()
			}
			if *modifyOriginInfoReq.OriginIPList != "" || *modifyOriginInfoReq.OriginDomain != "" {
				spinner.Start(fmt.Sprintf("Starting modify the pathx[%s] origin information", instanceId))
				modifyOriginInfoReq.SetProjectIdRef(modifyInstanceReq.GetProjectIdRef())
				modifyOriginInfoReq.SetRegionRef(modifyInstanceReq.GetRegionRef())
				modifyOriginInfoReq.SetZoneRef(modifyInstanceReq.GetZoneRef())
				_, err := base.BizClient.ModifyUGA3OriginInfo(modifyOriginInfoReq)
				if err != nil {
					spinner.Fail(err)
					return
				}
				spinner.Stop()
			}
			if *modifyInstanceReq.Name != "" || *modifyInstanceReq.Remark != "" {
				spinner.Start(fmt.Sprintf("Starting modify the pathx[%s] resource information", instanceId))
				_, err := base.BizClient.ModifyUGA3Instance(modifyInstanceReq)
				if err != nil {
					spinner.Fail(err)
					return
				}
				spinner.Stop()
			}

			// modify port
			tcpPortIntList := make([]int, 0)
			rsTcpPortIntList := make([]int, 0)
			if len(tcpPorts) > 0 || len(rsTcpPorts) > 0 {
				spinner.Start(fmt.Sprintf("Starting modify the pathx[%s] port", instanceId))
				if len(tcpPorts) == 0 {
					spinner.Fail(fmt.Errorf("The port cannot be empty."))
					return
				} else if len(rsTcpPorts) == 0 {
					spinner.Fail(fmt.Errorf("The origin-port cannot be empty."))
					return
				}
				if strings.EqualFold(protocol, "UDP") {
					spinner.Fail(fmt.Errorf("The udp protocol is temporarily not supported for create"))
					return
				} else if !strings.EqualFold(protocol, "TCP") &&
					!strings.EqualFold(protocol, "UDP") {
					spinner.Fail(fmt.Errorf("The value of protocol input error,please input 'TCP' or 'UDP',and the value entered is not case sensitive"))
					return
				}
				tcpPortList, err := formatPortList(tcpPorts)
				if err != nil {
					spinner.Fail(err)
					return
				}
				// tcpPorts convert to []int
				for _, tcpPort := range tcpPortList {
					port, _ := strconv.Atoi(tcpPort)
					tcpPortIntList = append(tcpPortIntList, port)
				}
				rsTcpPortList, err := formatPortList(rsTcpPorts)
				if err != nil {
					spinner.Fail(err)
					return
				}
				// rsTcpPorts convert to []int
				for _, rsTcpPort := range rsTcpPortList {
					rsPort, _ := strconv.Atoi(rsTcpPort)
					rsTcpPortIntList = append(rsTcpPortIntList, rsPort)
				}
				if len(tcpPortIntList) != len(rsTcpPortIntList) {
					spinner.Fail(fmt.Errorf("The number of port must be consistent with the number of origin-port."))
					return
				} else if len(tcpPortIntList) >= 10 {
					spinner.Fail(fmt.Errorf("The number of port cannot greater than or equals to 10"))
					return
				}
			}
			// ModifyUGA3Port
			if len(tcpPortIntList) > 0 && len(rsTcpPortIntList) > 0 {
				if strings.EqualFold(protocol, "TCP") {
					modifyPortReq.TCP = tcpPortIntList
					modifyPortReq.TCPRS = rsTcpPortIntList
				}
				modifyPortReq.SetProjectIdRef(modifyInstanceReq.GetProjectIdRef())
				modifyPortReq.SetRegionRef(modifyInstanceReq.GetRegionRef())
				modifyPortReq.SetZoneRef(modifyInstanceReq.GetZoneRef())
				_, err := base.BizClient.ModifyUGA3Port(modifyPortReq)
				if err != nil {
					base.HandleError(err)
					return
				}
				spinner.Stop()
			}
		},
	}
	flags := modifyCmd.Flags()
	flags.SortFlags = false

	bindProjectID(modifyInstanceReq, flags)
	bindRegion(modifyInstanceReq, flags)
	bindZone(modifyInstanceReq, flags)

	flags.StringVar(&instanceId, "id", "",
		"Required. It is the resource ID of the pathx")
	modifyBandwidthReq.Bandwidth = flags.Int("bandwidth", 0,
		"Optional. The bandwidth size. Its value range [1-100],no update if no value is specified")
	modifyOriginInfoReq.OriginIPList = flags.String("origin-ip", "",
		"Optional. Acceleration source IP. If multiple values exist,please split by ','")
	modifyOriginInfoReq.OriginDomain = flags.String("origin-domain", "",
		"Optional. Acceleration source domain name. Only 1 domain is supported")
	modifyInstanceReq.Name = flags.String("name", "",
		"Optional. Accelerate configuration resource name. If its value is not filled in or an empty string is not updated")
	modifyInstanceReq.Remark = flags.String("remark", "",
		"Optional. It will be modified if its value is not empty")

	flags.StringSliceVar(&tcpPorts, "port", nil,
		"Optional. Disable 65123 port,the port can be multiple,please split by ',' for example 80,3000-3010. "+
			"The number of port must be consistent with the number of origin-port,and the number cannot greater than or equals to 10")
	flags.StringSliceVar(&rsTcpPorts, "origin-port", nil,
		"Optional. The origin-port can be multiple,please split by ',' for example 80,3000-3010."+
			"The number of origin-port must be consistent with the number of port")
	flags.StringVar(&protocol, "protocol", "TCP", "Its values can be TCP and UDP, but currently only supports TCP")

	modifyCmd.MarkFlagRequired("id")
	flags.SetFlagValuesFunc("id", func() []string {
		return getPathxList(*modifyInstanceReq.ProjectId, *modifyInstanceReq.Region, *modifyInstanceReq.Zone)
	})
	return modifyCmd
}

// ucloud pathx list
func NewCmdUGA3List(out io.Writer) *cobra.Command {
	getPathxListReq := base.BizClient.NewDescribeUGA3InstanceRequest()
	var instanceId string
	var detail bool
	listCmd := &cobra.Command{
		Use:     "list",
		Short:   "List all the pathx resource of project",
		Long:    "List all the pathx resource of project",
		Example: "'ucloud pathx list or ucloud pathx list --id uga-xxx or ucloud pathx list --id uga-xxx --detail",
		Run: func(cmd *cobra.Command, args []string) {
			if len(instanceId) > 0 {
				getPathxListReq.InstanceId = &instanceId
			}
			resp, err := base.BizClient.DescribeUGA3Instance(getPathxListReq)
			if err != nil {
				base.HandleError(err)
				return
			}
			forwardInfos := resp.ForwardInstanceInfos
			// may be no UGA3 instance under the current project
			if len(forwardInfos) == 0 {
				base.HandleError(fmt.Errorf("No pathx resource found under the current project."))
				return
			}
			// print pathx detail information
			if detail && len(instanceId) > 0 {
				instanceInfo := forwardInfos[0]
				printPathxDetail(instanceInfo, out)
				return
			}
			list := make([]Uga3DescribeRow, 0)
			for _, item := range forwardInfos {
				row := Uga3DescribeRow{}
				row.ResourceID = item.InstanceId
				row.CName = item.CName
				row.Name = item.Name
				row.AccelerationArea = item.AccelerationArea
				row.Bandwidth = item.Bandwidth
				row.OriginAreaCode = item.OriginAreaCode
				row.IPList = strings.Join(item.IPList, ",")
				row.Domain = item.Domain
				row.CreateTime = base.FormatDate(item.CreateTime)

				var egressIps []string
				for _, egressIp := range item.EgressIpList {
					egressIps = append(egressIps, fmt.Sprintf("%s:%s", egressIp.Area, egressIp.IP))
				}
				row.EgressIpList = strings.Join(egressIps, "|")

				list = append(list, row)
			}
			base.PrintTable(list, []string{
				"ResourceID", "CName", "Name", "AccelerationArea", "OriginAreaCode",
				"Bandwidth", "EgressIpList", "IPList", "Domain", "CreateTime"})
		},
	}
	flags := listCmd.Flags()
	flags.SortFlags = false

	bindProjectID(getPathxListReq, flags)
	bindRegion(getPathxListReq, flags)
	bindZone(getPathxListReq, flags)

	flags.StringVar(&instanceId, "id", "", "Required. It is the resource ID of pathx resource")
	flags.BoolVar(&detail, "detail", false, "Optional. If it is specified,the details will be printed")
	flags.SetFlagValuesFunc("id", func() []string {
		return getPathxList(*getPathxListReq.ProjectId, *getPathxListReq.Region, *getPathxListReq.Zone)
	})
	return listCmd
}

func printPathxDetail(instanceInfo pathx.ForwardInfo, out io.Writer) {
	attrs := []base.DescribeTableRow{
		{Attribute: "ResourceID", Content: instanceInfo.InstanceId},
		{Attribute: "CName", Content: instanceInfo.CName},
		{Attribute: "Name", Content: instanceInfo.Name},
		{Attribute: "AccelerationArea", Content: instanceInfo.AccelerationArea},
		{Attribute: "AccelerationAreaName", Content: instanceInfo.AccelerationAreaName},
		{Attribute: "OriginAreaCode", Content: instanceInfo.OriginAreaCode},
		{Attribute: "OriginArea", Content: instanceInfo.OriginArea},
		{Attribute: "Bandwidth", Content: strconv.Itoa(instanceInfo.Bandwidth)},
		{Attribute: "ChargeType", Content: instanceInfo.ChargeType},
		{Attribute: "IPList", Content: strings.Join(instanceInfo.IPList, ",")},
		{Attribute: "Domain", Content: instanceInfo.Domain},
		{Attribute: "Remark", Content: instanceInfo.Remark},
		{Attribute: "CreateTime", Content: base.FormatDateTime(instanceInfo.CreateTime)},
		{Attribute: "ExpireTime", Content: base.FormatDateTime(instanceInfo.ExpireTime)},
	}
	for _, attr := range attrs {
		fmt.Fprintf(out, "%-22s: %s", attr.Attribute, attr.Content)
		fmt.Println()
	}
	// 加速节点列表
	if len(instanceInfo.AccelerationAreaInfos) > 0 {
		fmt.Println()
		fmt.Fprintln(out, "Acceleration area list:")
		for _, area := range instanceInfo.AccelerationAreaInfos {
			fmt.Fprintf(out, "%s:%5s\n", "Area", area.AccelerationArea)
			areaList := make([]PathxOptionalAreaRow, 0)
			for _, node := range area.AccelerationNodes {
				row := PathxOptionalAreaRow{
					AreaCode:    node.AreaCode,
					Area:        node.Area,
					FlagUnicode: node.FlagUnicode,
					FlagEmoji:   node.FlagEmoji,
				}
				areaList = append(areaList, row)
			}
			base.PrintTable(areaList, []string{"AreaCode", "Area", "FlagUnicode", "FlagEmoji"})
		}
	}
	// 回源出口IP地址
	if len(instanceInfo.EgressIpList) > 0 {
		fmt.Println()
		fmt.Fprintln(out, "Egress ip list:")
		egressIpList := make([]EgressIpInfoRow, 0)
		for _, egressIp := range instanceInfo.EgressIpList {
			row := EgressIpInfoRow{
				IP:   egressIp.IP,
				Area: egressIp.Area,
			}
			egressIpList = append(egressIpList, row)
		}
		base.PrintTable(egressIpList, []string{"Area", "IP"})
	}
	if len(instanceInfo.PortSets) > 0 {
		fmt.Println()
		fmt.Fprintln(out, "Port list:")
		portList := make([]Uga3PortRow, 0)
		for _, portItem := range instanceInfo.PortSets {
			row := Uga3PortRow{
				Protocol: portItem.Protocol,
				Port:     portItem.Port,
				RSPort:   portItem.RSPort,
			}
			portList = append(portList, row)
		}
		base.PrintTable(portList, []string{"Protocol", "Port", "RSPort"})
	}
}

// ucloud pathx-price
func NewCmdPathxPrice(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "price",
		Short: "List all the acceleration area price",
		Long:  "List all the acceleration area price",
	}
	cmd.AddCommand(NewPathxPriceList(out))
	// temporary not supports
	//cmd.AddCommand(NewPathxPriceUpgradeInfo())
	return cmd
}

// ucloud pathx price list
func NewPathxPriceList(out io.Writer) *cobra.Command {
	priceReq := base.BizClient.NewGetUGA3PriceRequest()
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List all the pathx acceleration area price",
		Long:    "List all the pathx acceleration area price",
		Example: "ucloud pathx price list --bandwidth 10 --area-code BKK --charge-type Month",
		Run: func(cmd *cobra.Command, args []string) {
			if strings.EqualFold(*priceReq.ChargeType, "Month") {
				*priceReq.Quantity = 0
			} else if *priceReq.Quantity <= 0 {
				base.HandleError(fmt.Errorf("If the value of charge-type is 'Year' or 'Hour',its value must be greater than 0"))
				return
			}

			switch strings.ToLower(*priceReq.ChargeType) {
			case "hour":
				*priceReq.ChargeType = "Dynamic"
			case "month":
				*priceReq.ChargeType = "Month"
			case "year":
				*priceReq.ChargeType = "Year"
			}

			response, err := base.BizClient.GetUGA3Price(priceReq)
			if err != nil {
				base.HandleError(err)
				return
			}
			list := make([]UGA3PriceRow, 0)
			priceList := response.UGA3Price
			if len(priceList) == 0 {
				base.HandleError(fmt.Errorf("Not found acceleration area price information."))
				return
			}
			//fmt.Fprintf(out,"Aceeleration area price information (unit:￥) :")
			for _, info := range priceList {
				row := UGA3PriceRow{
					AccelerationBandwidthPrice: fmt.Sprintf("%s%s", "￥", strconv.FormatFloat(info.AccelerationBandwidthPrice, 'g', 12, 64)),
					//AccelerationAreaName: info.AccelerationAreaName,
					AccelerationForwarderPrice: fmt.Sprintf("%s%s", "￥", strconv.FormatFloat(info.AccelerationForwarderPrice, 'g', 12, 64)),
					AccelerationArea:           info.AccelerationArea,
				}
				list = append(list, row)
			}
			base.PrintTable(list, []string{"AccelerationArea", "AccelerationBandwidthPrice", "AccelerationForwarderPrice"})
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	bindProjectID(priceReq, flags)
	bindRegion(priceReq, flags)
	bindZone(priceReq, flags)

	priceReq.Bandwidth = flags.Int("bandwidth", 1,
		"Required. The bandwidth of acceleration area to get price")
	priceReq.AreaCode = flags.String("area-code", "",
		"Required. The area-code of acceleration area to get price")
	priceReq.Quantity = flags.Int("quantity", 1,
		"Optional. When the value of the charge-type is 'Month',its default value is 0,"+
			"if the value of charge-type is 'Year' or 'Hour',its value must be greater than 0")
	priceReq.ChargeType = flags.String("charge-type", "",
		"Optional. Its value is not case sensitive,acceptable values:'Year',pay yearly;'Month',pay monthly;'Hour',pay hourly")
	priceReq.AccelerationArea = flags.String("accel", "",
		"Optional. The acceleration-area to get price")

	_ = cmd.MarkFlagRequired("bandwidth")
	_ = cmd.MarkFlagRequired("area-code")
	_ = flags.SetFlagValues("area-code", "BKK", "DXB", "FRA", "SGN", "HKG", "CGK", "LOS", "LHR", "LAX", "MNL", "DME", "BOM", "MSP", "ICN", "PVG", "SIN", "NRT", "IAD", "TPE")
	_ = flags.SetFlagValues("charge-type", "Year", "Month", "Hour")
	_ = flags.SetFlagValues("accel", "Global", "AP", "EU", "ME", "OA", "AF", "NA", "SA")
	return cmd
}

// ucloud pathx-area
func NewCmdPathxArea(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "area",
		Short: "List origin area or acceleration area information",
		Long:  "List origin area or acceleration area information",
	}
	cmd.AddCommand(NewCmdPathxAreaList(out))
	return cmd
}

// ucloud pathx area list
func NewCmdPathxAreaList(out io.Writer) *cobra.Command {
	areaGetReq := base.BizClient.NewDescribeUGA3AreaRequest()
	optimizationReq := base.BizClient.NewDescribeUGA3OptimizationRequest()
	var timeRange, accelerationArea, originDomain, originIp string
	var noAccel bool
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List origin area or acceleration area information",
		Long:    "Provide optional flags to get the optional list of global access source stations",
		Example: "ucloud pathx area list --origin-ip 0.0.0.0 --origin-domain test.com",
		Run: func(cmd *cobra.Command, args []string) {
			if len(originDomain) == 0 && len(originIp) == 0 {
				response, err := base.BizClient.DescribeUGA3Area(areaGetReq)
				if err != nil {
					base.HandleError(err)
					return
				}
				forwardAreas := response.AreaSet
				if len(forwardAreas) == 0 {
					base.HandleError(fmt.Errorf("Not found the origin area list"))
					return
				}
				areasGroup := make(map[string][]PathxOptionalAreaRow)
				for _, item := range forwardAreas {
					if areasGroup[item.ContinentCode] == nil {
						areasGroup[item.ContinentCode] = make([]PathxOptionalAreaRow, 0)
					}
					areasGroup[item.ContinentCode] = append(areasGroup[item.ContinentCode], PathxOptionalAreaRow{
						AreaCode:    item.AreaCode,
						Area:        item.Area,
						CountryCode: item.CountryCode,
						FlagUnicode: item.FlagUnicode,
						FlagEmoji:   item.FlagEmoji,
					})
				}
				fmt.Fprintln(out, "Origin areas :")
				for area := range areasGroup {
					fmt.Fprintf(out, "ContinentCode:  %s\n", area)
					rows := areasGroup[area]
					base.PrintTable(rows, []string{"AreaCode", "Area", "CountryCode", "FlagUnicode", "FlagEmoji"})
					fmt.Println()
				}
				return
			}
			areaGetReq.Domain = &originDomain
			areaGetReq.IPList = &originIp
			response, err := base.BizClient.DescribeUGA3Area(areaGetReq)
			if err != nil {
				base.HandleError(err)
				return
			}
			forwardAreas := response.AreaSet
			if len(forwardAreas) == 0 {
				base.HandleError(fmt.Errorf("Not found the origin area list"))
				return
			}
			// recommend one area for user
			forwardArea := forwardAreas[0]

			fmt.Fprintf(out, "Recommend origin area:(%s)\n", forwardArea.ContinentCode)
			areas := make([]PathxOptionalAreaRow, 0)
			areas = append(areas, PathxOptionalAreaRow{
				AreaCode:    forwardArea.AreaCode,
				Area:        forwardArea.Area,
				CountryCode: forwardArea.CountryCode,
				FlagUnicode: forwardArea.FlagUnicode,
				FlagEmoji:   forwardArea.FlagEmoji,
			})
			base.PrintTable(areas, []string{"AreaCode", "Area", "CountryCode", "FlagUnicode", "FlagEmoji"})
			fmt.Println()

			// display acceleration areas
			if !noAccel {
				areaCode := forwardAreas[0].AreaCode
				optimizationReq.AreaCode = &areaCode
				optimizationReq.AccelerationArea = &accelerationArea
				optimizationReq.TimeRange = &timeRange
				optimizationReq.SetProjectIdRef(areaGetReq.GetProjectIdRef())
				optimizationReq.SetRegionRef(areaGetReq.GetRegionRef())
				optimizationReq.SetZoneRef(areaGetReq.GetZoneRef())
				optimizationResponse, err := base.BizClient.DescribeUGA3Optimization(optimizationReq)
				if err != nil {
					base.HandleError(err)
					return
				}
				accelerationInfos := optimizationResponse.AccelerationInfos
				if len(accelerationInfos) == 0 {
					base.HandleError(fmt.Errorf("Not found the acceleration area information."))
					return
				}
				fmt.Fprintf(out, "Acceleration areas :\n")
				for _, item := range accelerationInfos {
					// User did not provide acceleration-area flag
					if len(accelerationArea) == 0 {
						fmt.Fprintf(out, "%s(%s):\n", item.AccelerationName, item.AccelerationArea)
					}
					list := make([]PathxOptimizationRow, 0)
					nodeDelays := item.NodeInfo
					for _, node := range nodeDelays {
						row := PathxOptimizationRow{}
						row.Area = node.Area
						row.AreaCode = node.AreaCode
						row.CountryCode = node.CountryCode
						row.FlagUnicode = node.FlagUnicode
						row.FlagEmoji = node.FlagEmoji
						row.Latency = fmt.Sprintf("%s%s", strconv.FormatFloat(node.Latency, 'g', 12, 64), "ms")
						row.LatencyWAN = fmt.Sprintf("%s%s", strconv.FormatFloat(node.LatencyInternet, 'g', 12, 64), "ms")
						row.LatencyPathX = fmt.Sprintf("%s%s", strconv.FormatFloat(node.LatencyOptimization, 'g', 12, 64), "%")
						row.Loss = fmt.Sprintf("%s%s", strconv.FormatFloat(node.Loss, 'g', 12, 64), "%")
						row.LossWAN = fmt.Sprintf("%s%s", strconv.FormatFloat(node.LossInternet, 'g', 12, 64), "%")
						row.LossPathx = fmt.Sprintf("%s%s", strconv.FormatFloat(node.LossOptimization, 'g', 12, 64), "%")
						list = append(list, row)
					}
					base.PrintTable(list, []string{"AreaCode", "Area", "CountryCode", "FlagUnicode", "FlagEmoji",
						"Latency", "LatencyWAN", "LatencyPathX", "Loss", "LossWAN", "LossPathx"})
				}
				return
			}

		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	bindProjectID(areaGetReq, flags)
	bindRegion(areaGetReq, flags)
	bindZone(areaGetReq, flags)

	flags.StringVar(&timeRange, "time-range", "",
		"Optional. The default value is 1 day. Acceptable values:'Hour','Day','Week',and its value is not case sensitive")
	flags.StringVar(&accelerationArea, "accel", "",
		"Optional. The acceleration area,acceptable values:'Global','AP','EU','ME','OA','AF','NA','SA'")
	flags.StringVar(&originDomain, "origin-domain", "",
		"Optional. If you fill in the IP or domain name, a region will be recommended as the first in the return list")
	flags.StringVar(&originIp, "origin-ip", "",
		"Optional. If you fill in the IP or domain name, a region will be recommended as the first IP collection of the source station in the return list, "+
			"split by ',' example:110.10.10.1,111.100.0.10 ")
	flags.BoolVar(&noAccel, "no-accel", false,
		"Optional. If it is specified,the print result will not be displayed acceleration areas")

	return cmd
}

type UGA3PriceRow struct {
	// 加速大区代码
	AccelerationArea string
	// 加速大区名称
	AccelerationAreaName string
	// 转发配置价格
	AccelerationForwarderPrice string
	// 加速配置带宽价格
	AccelerationBandwidthPrice string
}

// describe UGA3 instance information row
type Uga3DescribeRow struct {
	// 加速配置ID
	ResourceID string
	// 加速域名
	CName string
	// 加速实例名称
	Name string
	// 加速区域
	AccelerationArea string
	// 加速区域名称
	AccelerationAreaName string
	// 回源出口IP地址
	EgressIpList string
	// 购买的带宽值
	Bandwidth int
	// 备注
	Remark string
	// 源站中文名
	OriginArea string
	// 源站AreaCode
	OriginAreaCode string
	// 资源创建时间
	CreateTime string
	// 资源过期时间
	ExpireTime string
	// 计费方式
	ChargeType string
	// 源站IP列表，多个值由半角英文逗号相隔
	IPList string
	// 源站域名
	Domain string
}

// pathx port print row
type Uga3PortRow struct {
	// 转发协议，枚举值["TCP"，"UDP"，"HTTPHTTP"，"HTTPSHTTP"，"HTTPSHTTPS"，"WSWS"，"WSSWS"，"WSSWSS"]。TCP和UDP代表四层转发，其余为七层转发。
	Protocol string
	// 源站服务器监听的端口号
	RSPort int
	// 加速端口
	Port int
}

// pathx price upgrade-info print row
type PathxUpdatePriceRow struct {
	// 实例ID
	InstanceId string
	// 带宽
	Bandwidth int
	// 更新价格
	UpdatePrice float64
}

// pathx optimization print row
type PathxOptimizationRow struct {
	// 加速大区名称
	AccelerationName string
	// 加速大区代码
	AccelerationArea string
	// 加速区域
	Area string
	// 加速区域Code
	AreaCode string
	// 国家代码
	CountryCode string
	// 国旗Code
	FlagUnicode string
	// 国旗Emoji
	FlagEmoji string
	// 加速延迟
	Latency string
	// 公网延迟
	LatencyWAN string
	// 加速提升比例
	LatencyPathX string
	// 加速后丢包率
	Loss string
	// 原始丢包率
	LossWAN string
	// 丢包下降比例
	LossPathx string
}

// row for print of pathx area
type PathxOptionalAreaRow struct {
	AreaCode      string
	Area          string
	CountryCode   string
	FlagUnicode   string
	FlagEmoji     string
	ContinentCode string
}

// row for print of egressIpList
type EgressIpInfoRow struct {
	// 线路出口EIP
	IP string
	// 线路出口机房代号
	Area string
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
