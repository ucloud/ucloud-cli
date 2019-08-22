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
	"io/ioutil"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/ulb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/base"
	"github.com/ucloud/ucloud-cli/model/status"
)

//NewCmdULB  ucloud ulb
func NewCmdULB() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ulb",
		Short: "List and manipulate ULB instances",
		Long:  "List and manipulate ULB instances",
	}
	out := base.Cxt.GetWriter()

	cmd.AddCommand(NewCmdULBList(out))
	cmd.AddCommand(NewCmdULBCreate(out))
	cmd.AddCommand(NewCmdULBUpdate(out))
	cmd.AddCommand(NewCmdULBDelete(out))
	cmd.AddCommand(NewCmdULBVserver())
	cmd.AddCommand(NewCmdULBSSL())

	return cmd
}

//ULBRow 表格行
type ULBRow struct {
	Name         string
	ResourceID   string
	Group        string
	Network      string
	VserverCount int
	VPC          string
	CreationTime string
}

//NewCmdULBList ucloud ulb list
func NewCmdULBList(out io.Writer) *cobra.Command {
	req := base.BizClient.NewDescribeULBRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List ULB instances",
		Long:  "List ULB instances",
		Run: func(c *cobra.Command, args []string) {
			req.ProjectId = sdk.String(base.PickResourceID(*req.ProjectId))
			req.VPCId = sdk.String(base.PickResourceID(*req.VPCId))
			resp, err := base.BizClient.DescribeULB(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			list := []ULBRow{}
			for _, ulb := range resp.DataSet {
				row := ULBRow{}
				row.ResourceID = ulb.ULBId
				row.Name = ulb.Name
				row.Group = ulb.BusinessId
				row.VserverCount = len(ulb.VServerSet)
				row.VPC = ulb.VPCId
				row.CreationTime = base.FormatDate(ulb.CreateTime)
				if ulb.ULBType == "OuterMode" {
					ips := []string{}
					for _, ip := range ulb.IPSet {
						ips = append(ips, fmt.Sprintf("%s(%s)", ip.EIP, ip.EIPId))
					}
					row.Network = strings.Join(ips, ",")
				} else {
					row.Network = ulb.PrivateIP
				}
				list = append(list, row)
			}

			base.PrintList(list, out)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	bindRegion(req, flags)
	bindProjectID(req, flags)

	req.ULBId = flags.String("ulb-id", "", "Optional. Resource ID of ULB instance to list")
	req.VPCId = flags.String("vpc-id", "", "Optional. Resource ID of VPC which the ULB instances to list belong to")
	req.SubnetId = flags.String("subnet-id", "", "Optional. Resource ID of subnet which the ULB instances to list belong to")
	req.BusinessId = flags.String("group", "", "Optional. Business group of ULB instances to list")
	req.Offset = flags.Int("offset", 0, "Optional. Offset")
	req.Limit = flags.Int("limit", 50, "Optional. Limit")

	flags.SetFlagValuesFunc("vpc-id", func() []string {
		return getAllVPCIdNames(*req.ProjectId, *req.Region)
	})

	return cmd
}

//NewCmdULBCreate ucloud ulb create
func NewCmdULBCreate(out io.Writer) *cobra.Command {
	var bindEipID *string
	mode := "outer"
	req := base.BizClient.NewCreateULBRequest()
	eipReq := base.BizClient.NewAllocateEIPRequest()
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create ULB instance",
		Long:  "Create ULB instance",
		Run: func(c *cobra.Command, args []string) {
			req.ProjectId = sdk.String(base.PickResourceID(*req.ProjectId))
			if mode == "outer" {
				if *bindEipID == "" && *eipReq.Bandwidth == 0 {
					fmt.Fprintln(out, "Outer mode ULB need a eip to bind, please assign eip by flag 'bind-eip' or create eip by 'create-eip-bandwidth-mb'")
					return
				}
				if *eipReq.OperatorName == "" {
					*eipReq.OperatorName = getEIPLine(*req.Region)
				}
				req.OuterMode = sdk.String("Yes")
			} else if mode == "inner" {
				req.InnerMode = sdk.String("Yes")
			} else {
				fmt.Fprintln(out, "Error, flag mode should be 'outer' or 'inner'")
				return
			}
			req.VPCId = sdk.String(base.PickResourceID(*req.VPCId))
			req.SubnetId = sdk.String(base.PickResourceID(*req.SubnetId))
			resp, err := base.BizClient.CreateULB(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			fmt.Fprintf(out, "ulb[%s] created\n", resp.ULBId)
			if mode == "inner" {
				return
			}
			bindEipID = sdk.String(base.PickResourceID(*bindEipID))
			if *bindEipID != "" {
				bindEIP(sdk.String(resp.ULBId), sdk.String("ulb"), bindEipID, req.ProjectId, req.Region)
				return
			}
			if *eipReq.OperatorName != "" && *eipReq.Bandwidth != 0 {
				eipReq.ChargeType = req.ChargeType
				eipReq.Tag = req.Tag
				eipReq.Region = req.Region
				eipReq.ProjectId = req.ProjectId
				eipResp, err := base.BizClient.AllocateEIP(eipReq)

				if err != nil {
					base.HandleError(err)
					return
				}

				for _, eip := range eipResp.EIPSet {
					base.Cxt.Printf("allocate EIP[%s] ", eip.EIPId)
					for _, ip := range eip.EIPAddr {
						base.Cxt.Printf("IP:%s  Line:%s \n", ip.IP, ip.OperatorName)
					}
					bindEIP(sdk.String(resp.ULBId), sdk.String("ulb"), sdk.String(eip.EIPId), req.ProjectId, req.Region)
				}
			}
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.ULBName = flags.String("name", "", "Required. Name of ULB instance to create")
	flags.StringVar(&mode, "mode", "outer", "Required. Network mode of ULB instance, outer or inner.")
	bindRegion(req, flags)
	bindProjectID(req, flags)
	req.VPCId = flags.String("vpc-id", "", "Optional. Resource ID of VPC which the ULB to create belong to. See 'ucloud vpc list'")
	req.SubnetId = flags.String("subnet-id", "", "Optional. Resource ID of subnet. This flag will be discarded when you are creating an outter mode ULB. See 'ucloud subnet list'")
	req.ChargeType = flags.String("charge-type", "Month", "Optional.'Year',pay yearly;'Month',pay monthly;'Dynamic', pay hourly")
	req.Tag = flags.String("group", "Default", "Optional. Business group")
	req.Remark = flags.String("remark", "", "Optional. Remark of instance to create.")
	bindEipID = flags.String("bind-eip", "", "Optional. Resource ID or IP Address of eip that will be bound to the new created outer mode ulb")
	eipReq.Bandwidth = cmd.Flags().Int("create-eip-bandwidth-mb", 0, "Optional. Required if you want to create new EIP. Bandwidth(Unit:Mbps).The range of value related to network charge mode. By traffic [1, 300]; by bandwidth [1,800] (Unit: Mbps); it could be 0 if the eip belong to the shared bandwidth")
	eipReq.OperatorName = flags.String("create-eip-line", "", "Optional. Line of created eip to bind with the new created outer mode ulb")
	eipReq.PayMode = cmd.Flags().String("create-eip-traffic-mode", "Bandwidth", "Optional. 'Traffic','Bandwidth' or 'ShareBandwidth'")
	eipReq.Name = flags.String("create-eip-name", "", "Optional. Name of created eip to bind with the new created outer mode ulb")
	eipReq.Remark = cmd.Flags().String("create-eip-remark", "", "Optional. Remark of your EIP.")

	flags.SetFlagValues("mode", "outer", "inner")
	flags.SetFlagValues("charge-type", "Month", "Year", "Dynamic")
	flags.SetFlagValues("create-eip-line", "BGP", "International")
	flags.SetFlagValues("create-eip-traffic-mode", "Bandwidth", "Traffic", "ShareBandwidth")
	flags.SetFlagValuesFunc("bind-eip", func() []string {
		return getAllEip(*req.ProjectId, *req.Region, []string{status.EIP_FREE}, nil)
	})
	flags.SetFlagValuesFunc("vpc-id", func() []string {
		return getAllVPCIdNames(*req.ProjectId, *req.Region)
	})
	flags.SetFlagValuesFunc("subnet-id", func() []string {
		return getAllSubnetIDNames(*req.VPCId, *req.ProjectId, *req.Region)
	})

	cmd.MarkFlagRequired("mode")
	cmd.MarkFlagRequired("name")

	return cmd
}

//NewCmdULBDelete ucloud ulb delete
func NewCmdULBDelete(out io.Writer) *cobra.Command {
	idNames := []string{}
	req := base.BizClient.NewDeleteULBRequest()
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete ULB instances by resource ID",
		Long:  "Delete ULB instances by resource ID",
		Run: func(c *cobra.Command, args []string) {
			req.ProjectId = sdk.String(base.PickResourceID(*req.ProjectId))
			for _, idname := range idNames {
				req.ULBId = sdk.String(base.PickResourceID(idname))
				_, err := base.BizClient.DeleteULB(req)
				if err != nil {
					base.HandleError(err)
					return
				}
				fmt.Fprintf(out, "ulb[%s] deleted\n", idname)
			}
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&idNames, "ulb-id", nil, "Required. Resource ID of the ULB instances to delete")
	bindRegion(req, flags)
	bindProjectID(req, flags)

	flags.SetFlagValuesFunc("ulb-id", func() []string {
		return getAllULBIDNames(*req.ProjectId, *req.Region)
	})

	cmd.MarkFlagRequired("ulb-id")

	return cmd
}

//NewCmdULBUpdate ucloud ulb update
func NewCmdULBUpdate(out io.Writer) *cobra.Command {
	var name, group, remark string
	idNames := []string{}
	req := base.BizClient.NewUpdateULBAttributeRequest()
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update ULB instance",
		Long:  "Update ULB instance",
		Run: func(c *cobra.Command, args []string) {
			req.ProjectId = sdk.String(base.PickResourceID(*req.ProjectId))
			for _, idname := range idNames {
				req.ULBId = sdk.String(base.PickResourceID(idname))
				if name == "" && group == "" && remark == "" {
					fmt.Fprintln(out, "Error, name, remark and group can't be all empty")
					return
				}
				if name != "" {
					req.Name = &name
				}
				if group != "" {
					req.Tag = &group
				}
				if remark != "" {
					req.Remark = &remark
				}
				_, err := base.BizClient.UpdateULBAttribute(req)
				if err != nil {
					base.HandleError(err)
					continue
				}
				fmt.Fprintf(out, "ulb[%s] updated\n", *req.ULBId)
			}
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	bindRegion(req, flags)
	bindProjectID(req, flags)
	flags.StringSliceVar(&idNames, "ulb-id", nil, "Required. Resource ID of ULB instances to update")
	flags.StringVar(&name, "name", "", "Optional, Name of ULB instance")
	flags.StringVar(&remark, "remark", "", "Optional, Remark of ULB instance")
	flags.StringVar(&group, "group", "", "Optional, Business group of ULB instance")
	// bindGroup(&group, flags)

	flags.SetFlagValuesFunc("ulb-id", func() []string {
		return getAllULBIDNames(*req.ProjectId, *req.Region)
	})

	cmd.MarkFlagRequired("ulb-id")

	return cmd
}

func getAllULB(project, region string) ([]ulb.ULBSet, error) {
	list := []ulb.ULBSet{}
	req := base.BizClient.NewDescribeULBRequest()
	req.ProjectId = &project
	req.Region = &region

	for offset, limit := 0, 50; ; offset += limit {
		req.Offset = sdk.Int(offset)
		req.Limit = sdk.Int(limit)
		resp, err := base.BizClient.DescribeULB(req)

		if err != nil {
			return nil, err
		}
		list = append(list, resp.DataSet...)

		if resp.TotalCount < offset+limit {
			break
		}
	}
	return list, nil
}

func getAllULBIDNames(project, region string) []string {
	list := []string{}
	ulbList, err := getAllULB(project, region)
	if err != nil {
		return nil
	}
	for _, ulb := range ulbList {
		list = append(list, fmt.Sprintf("%s/%s", ulb.ULBId, ulb.Name))
	}
	return list
}

//NewCmdULBVserver ucloud ulb-vserver
func NewCmdULBVserver() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vserver",
		Short: "List and manipulate ULB Vserver instances",
		Long:  "List and manipulate ULB Vserver instances",
	}
	out := base.Cxt.GetWriter()

	cmd.AddCommand(NewCmdULBVServerList(out))
	cmd.AddCommand(NewCmdULBVServerCreate(out))
	cmd.AddCommand(NewCmdULBVServerUpdate(out))
	cmd.AddCommand(NewCmdULBVServerDelete(out))
	cmd.AddCommand(NewCmdULBVServerNode())
	cmd.AddCommand(NewCmdULBVServerPolicy())

	return cmd
}

//ULBVServerRow 表格行
type ULBVServerRow struct {
	VServerName         string
	ResourceID          string
	ListenType          string
	Protocol            string
	Port                int
	LBMethod            string
	SessionMaintainMode string
	SessionMaintainKey  string
	ClientTimeout       string
	HealthCheckMode     string
	HealthCheckDomain   string
	HealthCheckPath     string
}

//NewCmdULBVServerList ucloud ulb-vserver list
func NewCmdULBVServerList(out io.Writer) *cobra.Command {
	req := base.BizClient.NewDescribeVServerRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List ULB Vserver instances",
		Long:  "List ULB Vserver instances",
		Run: func(c *cobra.Command, args []string) {
			req.ProjectId = sdk.String(base.PickResourceID(*req.ProjectId))
			req.ULBId = sdk.String(base.PickResourceID(*req.ULBId))
			resp, err := base.BizClient.DescribeVServer(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			list := []ULBVServerRow{}
			for _, vs := range resp.DataSet {
				row := ULBVServerRow{}
				row.VServerName = vs.VServerName
				row.ResourceID = vs.VServerId
				row.ListenType = vs.ListenType
				row.Protocol = vs.Protocol
				row.Port = vs.FrontendPort
				row.LBMethod = vs.Method
				row.ClientTimeout = fmt.Sprintf("%ds", vs.ClientTimeout)
				row.SessionMaintainMode = vs.PersistenceType
				row.SessionMaintainKey = vs.PersistenceInfo
				row.HealthCheckMode = vs.MonitorType
				row.HealthCheckDomain = vs.Domain
				row.HealthCheckPath = vs.Path
				list = append(list, row)
			}
			base.PrintList(list, out)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	bindRegion(req, flags)
	bindProjectID(req, flags)
	req.ULBId = flags.String("ulb-id", "", "Required. Resource ID of ULB")
	req.VServerId = flags.String("vserver-id", "", "Optional. Resource ID of vserver to list")

	flags.SetFlagValuesFunc("ulb-id", func() []string {
		return getAllULBIDNames(*req.ProjectId, *req.Region)
	})

	cmd.MarkFlagRequired("ulb-id")

	return cmd
}

//NewCmdULBVServerCreate ucloud ulb-vserver create
func NewCmdULBVServerCreate(out io.Writer) *cobra.Command {
	sslID := ""
	req := base.BizClient.NewCreateVServerRequest()
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create ULB VServer instance",
		Long:  "Create ULB VServer instance",
		Run: func(c *cobra.Command, args []string) {
			if *req.ListenType == "RequestProxy" && (*req.ClientTimeout <= 0 || *req.ClientTimeout > 86400) {
				fmt.Println("Error, client-timeout-seconds in the range of (0,86400]")
				return
			}
			if *req.ListenType == "PacketsTransmit" && (*req.ClientTimeout <= 0 || *req.ClientTimeout > 86400) {
				fmt.Println("Error, client-timeout-seconds in the range of [60，900]")
				return
			}
			if *req.Protocol == "HTTPS" && sslID == "" {
				fmt.Println("Error, SSL Certificate is needed when you choose HTTPS")
				return
			}
			req.ProjectId = sdk.String(base.PickResourceID(*req.ProjectId))
			req.ULBId = sdk.String(base.PickResourceID(*req.ULBId))
			resp, err := base.BizClient.CreateVServer(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			fmt.Fprintf(out, "ulb-vserver[%s] created\n", resp.VServerId)
			if *req.Protocol == "HTTPS" && sslID != "" {
				bindReq := base.BizClient.NewBindSSLRequest()
				bindReq.Region = req.Region
				bindReq.ProjectId = req.ProjectId
				bindReq.SSLId = sdk.String(base.PickResourceID(sslID))
				bindReq.VServerId = sdk.String(resp.VServerId)
				bindReq.ULBId = req.ULBId
				_, err := base.BizClient.BindSSL(bindReq)
				if err != nil {
					base.HandleError(err)
					return
				}
				fmt.Fprintf(out, "ssl certificate[%s] bind with vserver[%s] of ulb[%s]\n", sslID, *bindReq.VServerId, *bindReq.ULBId)
			}
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.ULBId = flags.String("ulb-id", "", "Required. Resource ID of ULB instance which the VServer to create belongs to")
	bindRegion(req, flags)
	bindProjectID(req, flags)
	req.VServerName = flags.String("name", "", "Optional. Name of VServer to create")
	req.ListenType = flags.String("listen-type", "RequestProxy", "Optional. Listen type, 'RequestProxy' or 'PacketsTransmit'")
	req.Protocol = flags.String("protocol", "HTTP", "Optional. Protocol of VServer instance, 'HTTP','HTTPS','TCP' for listen type 'RequestProxy' and 'TCP','UDP' for listen type 'PacketsTransmit'")
	req.FrontendPort = flags.Int("port", 80, "Optional. Port of VServer instance")
	flags.StringVar(&sslID, "ssl-id", "", "Optional. Required if you choose HTTPS, Resource ID of SSL Certificate")
	req.Method = flags.String("lb-method", "Roundrobin", "Optional. LB methods, accept values:Roundrobin,Source,ConsistentHash,SourcePort,ConsistentHashPort,WeightRoundrobin and Leastconn. \nConsistentHash,SourcePort and ConsistentHashPort are effective for listen type PacketsTransmit only;\nLeastconn is effective for listen type RequestProxy only;\nRoundrobin,Source and WeightRoundrobin are effective for both listen types")
	req.PersistenceType = flags.String("session-maintain-mode", "None", "Optional. The method of maintaining user's session. Accept values: 'None','ServerInsert' and 'UserDefined'. 'None' meaning don't maintain user's session'; 'ServerInsert' meaning auto create session key; 'UserDefined' meaning specify session key which accpeted by flag seesion-maintain-key by yourself")
	req.PersistenceInfo = flags.String("session-maintain-key", "", "Optional. Specify a key for maintaining session")
	req.ClientTimeout = flags.Int("client-timeout-seconds", 60, "Optional.Unit seconds. For 'RequestProxy', it's lifetime for idle connections, range (0，86400]. For 'PacketsTransmit', it's the duration of the connection is maintained, range [60，900]")
	req.MonitorType = flags.String("health-check-mode", "Port", "Optional. Method of checking real server's status of health. Accept values:'Port','Path'")
	req.Domain = flags.String("health-check-domain", "", "Optional. Skip this flag if health-check-mode is assigned Port")
	req.Path = flags.String("health-check-path", "", "Optional. Skip this flags if health-check-mode is assigned Port")

	flags.SetFlagValues("listen-type", "RequestProxy", "PacketsTransmit")
	flags.SetFlagValues("protocol", "HTTP", "HTTPS", "TCP", "UDP")
	flags.SetFlagValuesFunc("lb-method", func() []string {
		if *req.ListenType == "RequestProxy" {
			return []string{"Roundrobin", "Source", "WeightRoundrobin", "Leastconn"}
		} else if *req.ListenType == "PacketsTransmit" {
			return []string{"Roundrobin", "Source", "WeightRoundrobin", "ConsistentHash", "SourcePort", "ConsistentHashPort"}
		}
		return []string{"Roundrobin", "Source", "WeightRoundrobin", "ConsistentHash", "SourcePort", "ConsistentHashPort", "Leastconn"}
	})
	flags.SetFlagValues("session-maintain-mode", "None", "ServerInsert", "UserDefined")
	flags.SetFlagValues("health-check-mode", "Port", "Path")
	flags.SetFlagValuesFunc("ulb-id", func() []string {
		return getAllULBIDNames(*req.ProjectId, *req.Region)
	})
	flags.SetFlagValuesFunc("ssl-id", func() []string {
		return getAllSSLCertIDNames(*req.ProjectId, *req.Region)
	})

	cmd.MarkFlagRequired("ulb-id")

	return cmd
}

//NewCmdULBVServerUpdate ucloud ulb-vserver update
func NewCmdULBVServerUpdate(out io.Writer) *cobra.Command {
	req := base.BizClient.NewUpdateVServerAttributeRequest()
	vserverIDs := []string{}
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update attributes of VServer instances",
		Long:  "Update attributes of VServer instances",
		Run: func(c *cobra.Command, args []string) {
			if *req.VServerName == "" {
				req.VServerName = nil
			}
			if *req.Method == "" {
				req.Method = nil
			}
			if *req.PersistenceType == "" {
				req.PersistenceType = nil
			}
			if *req.PersistenceInfo == "" {
				req.PersistenceInfo = nil
			}
			if *req.ClientTimeout == -1 {
				req.ClientTimeout = nil
			}
			if *req.MonitorType == "" {
				req.MonitorType = nil
			}
			if *req.Domain == "" {
				req.Domain = nil
			}
			if *req.Path == "" {
				req.Path = nil
			}
			req.ProjectId = sdk.String(base.PickResourceID(*req.ProjectId))
			req.ULBId = sdk.String(base.PickResourceID(*req.ULBId))
			for _, idname := range vserverIDs {
				req.VServerId = sdk.String(base.PickResourceID(idname))
				_, err := base.BizClient.UpdateVServerAttribute(req)
				if err != nil {
					base.HandleError(err)
					return
				}
				fmt.Fprintf(out, "ulb-vserver[%s] updated\n", *req.VServerId)
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	req.ULBId = flags.String("ulb-id", "", "Required. Resource ID of ULB instance which the VServer to create belongs to")
	flags.StringSliceVar(&vserverIDs, "vserver-id", nil, "Required. Resource ID of Vserver to update")
	bindRegion(req, flags)
	bindProjectID(req, flags)
	req.VServerName = flags.String("name", "", "Optional. Name of VServer")
	req.Method = flags.String("lb-method", "", "Optional. LB methods, accept values:Roundrobin,Source,ConsistentHash,SourcePort,ConsistentHashPort,WeightRoundrobin and Leastconn. \nConsistentHash,SourcePort and ConsistentHashPort are effective for listen type PacketsTransmit only;\nLeastconn is effective for listen type RequestProxy only;\nRoundrobin,Source and WeightRoundrobin are effective for both listen types")
	req.PersistenceType = flags.String("session-maintain-mode", "", "Optional. The method of maintaining user's session. Accept values: 'None','ServerInsert' and 'UserDefined'. 'None' meaning don't maintain user's session'; 'ServerInsert' meaning auto create session key; 'UserDefined' meaning specify session key which accpeted by flag seesion-maintain-key by yourself")
	req.PersistenceInfo = flags.String("session-maintain-key", "", "Optional. Specify a key for maintaining session")
	req.ClientTimeout = flags.Int("client-timeout-seconds", -1, "Optional.Unit seconds. For 'RequestProxy', it's lifetime for idle connections, range (0，86400]. For 'PacketsTransmit', it's the duration of the connection is maintained, range [60，900]")
	req.MonitorType = flags.String("health-check-mode", "", "Optional. Method of checking real server's status of health. Accept values:'Port','Path'")
	req.Domain = flags.String("health-check-domain", "", "Optional. Skip this flag if health-check-mode is assigned Port")
	req.Path = flags.String("health-check-path", "", "Optional. Skip this flags if health-check-mode is assigned Port")

	flags.SetFlagValues("lb-method", "Roundrobin", "Source", "WeightRoundrobin", "ConsistentHash", "SourcePort", "ConsistentHashPort", "Leastconn")
	flags.SetFlagValues("session-maintain-mode", "None", "ServerInsert", "UserDefined")
	flags.SetFlagValues("health-check-mode", "Port", "Path")
	flags.SetFlagValuesFunc("ulb-id", func() []string {
		return getAllULBIDNames(*req.ProjectId, *req.Region)
	})
	flags.SetFlagValuesFunc("vserver-id", func() []string {
		ulbID := base.PickResourceID(*req.ULBId)
		return getAllULBVServerIDNames(ulbID, *req.ProjectId, *req.Region)
	})

	cmd.MarkFlagRequired("ulb-id")
	cmd.MarkFlagRequired("vserver-id")

	return cmd
}

//NewCmdULBVServerDelete ucloud ulb-vserver delete
func NewCmdULBVServerDelete(out io.Writer) *cobra.Command {
	vserverIDs := []string{}
	req := base.BizClient.NewDeleteVServerRequest()
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete ULB VServer instances",
		Long:  "Delete ULB VServer instances",
		Run: func(c *cobra.Command, args []string) {
			req.ProjectId = sdk.String(base.PickResourceID(*req.ProjectId))
			req.ULBId = sdk.String(base.PickResourceID(*req.ULBId))
			for _, idname := range vserverIDs {
				vsid := base.PickResourceID(idname)
				req.VServerId = sdk.String(vsid)
				_, err := base.BizClient.DeleteVServer(req)
				if err != nil {
					base.HandleError(err)
					return
				}
				fmt.Fprintf(out, "ulb-vserver[%s] deleted\n", idname)
			}
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.ULBId = flags.String("ulb-id", "", "Required. Resource ID of ULB instance which the VServer to create belongs to")
	flags.StringSliceVar(&vserverIDs, "vserver-id", nil, "Required. Resource ID of Vserver to update")
	bindRegion(req, flags)
	bindProjectID(req, flags)

	cmd.MarkFlagRequired("ulb-id")
	cmd.MarkFlagRequired("vserver-id")

	flags.SetFlagValuesFunc("ulb-id", func() []string {
		return getAllULBIDNames(*req.ProjectId, *req.Region)
	})
	flags.SetFlagValuesFunc("vserver-id", func() []string {
		ulbID := base.PickResourceID(*req.ULBId)
		return getAllULBVServerIDNames(ulbID, *req.ProjectId, *req.Region)
	})

	return cmd
}

//NewCmdULBVServerNode ucloud ulb vserver node
func NewCmdULBVServerNode() *cobra.Command {
	out := base.Cxt.GetWriter()
	cmd := &cobra.Command{
		Use:   "backend",
		Short: "List and manipulate VServer backend nodes",
		Long:  "List and manipulate VServer backend nodes",
	}
	cmd.AddCommand(NewCmdULBVServerListNode(out))
	cmd.AddCommand(NewCmdULBVServerAddNode(out))
	cmd.AddCommand(NewCmdULBVServerUpdateNode(out))
	cmd.AddCommand(NewCmdULBVServerDeleteNode(out))
	return cmd
}

//ULBVServerNode 表格行
type ULBVServerNode struct {
	Name        string
	ResourceID  string
	BackendID   string
	PrivateIP   string
	Port        int
	HealthCheck string
	NodeMode    string
	Weight      int
}

//NewCmdULBVServerListNode ucloud ulb-vserver list-node
func NewCmdULBVServerListNode(out io.Writer) *cobra.Command {
	req := base.BizClient.NewDescribeVServerRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List ULB VServer backend nodes",
		Long:  "List ULB VServer backend nodes",
		Run: func(c *cobra.Command, args []string) {
			req.ProjectId = sdk.String(base.PickResourceID(*req.ProjectId))
			req.ULBId = sdk.String(base.PickResourceID(*req.ULBId))
			req.VServerId = sdk.String(base.PickResourceID(*req.VServerId))
			resp, err := base.BizClient.DescribeVServer(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			if len(resp.DataSet) != 1 {
				fmt.Fprintf(out, "ulb[%s] or vserver[%s] may not exist\n", *req.ULBId, *req.VServerId)
				return
			}
			vs := resp.DataSet[0]
			list := []ULBVServerNode{}
			for _, node := range vs.BackendSet {
				row := ULBVServerNode{}
				row.Name = node.ResourceName
				row.ResourceID = node.ResourceId
				row.BackendID = node.BackendId
				row.PrivateIP = node.PrivateIP
				row.Weight = node.Weight
				row.Port = node.Port
				if node.Status == 0 {
					row.HealthCheck = "Normal"
				} else if node.Status == 1 {
					row.HealthCheck = "Failed"
				}
				if node.Enabled == 1 {
					row.NodeMode = "enable"
				} else if node.Enabled == 0 {
					row.NodeMode = "disable"
				}
				list = append(list, row)
			}
			base.PrintList(list, out)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	req.ULBId = flags.String("ulb-id", "", "Required. Resource ID of ULB which the backend nodes belong to")
	req.VServerId = flags.String("vserver-id", "", "Required. Resource ID of VServer which the backend nodes belong to")
	bindRegion(req, flags)
	bindProjectID(req, flags)

	cmd.MarkFlagRequired("ulb-id")
	cmd.MarkFlagRequired("vserver-id")

	flags.SetFlagValuesFunc("ulb-id", func() []string {
		return getAllULBIDNames(*req.ProjectId, *req.Region)
	})
	flags.SetFlagValuesFunc("vserver-id", func() []string {
		ulbID := base.PickResourceID(*req.ULBId)
		return getAllULBVServerIDNames(ulbID, *req.ProjectId, *req.Region)
	})

	return cmd
}

//NewCmdULBVServerAddNode ucloud ulb-vserver add-node
func NewCmdULBVServerAddNode(out io.Writer) *cobra.Command {
	var enable *string
	var weight *int
	var ids []string
	req := base.BizClient.NewAllocateBackendRequest()
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add backend nodes for ULB Vserver instance",
		Long:  "Add backend nodes for ULB Vserver instance",
		Run: func(c *cobra.Command, args []string) {
			if *enable == "enable" {
				req.Enabled = sdk.Int(1)
			} else if *enable == "disable" {
				req.Enabled = sdk.Int(0)
			} else {
				fmt.Fprintln(out, "Error, backend-mode must be enable or disable")
				return
			}
			if *weight < 0 || *weight > 100 {
				fmt.Fprintln(out, "Error, weight must be between 0 and 100")
				return
			}
			req.ProjectId = sdk.String(base.PickResourceID(*req.ProjectId))
			req.ULBId = sdk.String(base.PickResourceID(*req.ULBId))
			req.VServerId = sdk.String(base.PickResourceID(*req.VServerId))
			for _, id := range ids {
				req.ResourceId = sdk.String(id)
				resp, err := base.BizClient.AllocateBackend(req)
				if err != nil {
					base.HandleError(err)
					continue
				}
				fmt.Fprintf(out, "backend node[%s] added, backend-id:%s\n", *req.ResourceId, resp.BackendId)
			}
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	req.ULBId = flags.String("ulb-id", "", "Required. Resource ID of ULB which the backend nodes belong to")
	req.VServerId = flags.String("vserver-id", "", "Required. Resource ID of VServer which the backend nodes belong to")
	flags.StringSliceVar(&ids, "resource-id", nil, "Required. Resource ID of the backend nodes to add")
	bindRegion(req, flags)
	bindProjectID(req, flags)
	req.ResourceType = flags.String("resource-type", "UHost", "Optional. Resource type of the backend node to add. Accept values: UHost,UPM,UDHost,UDocker")
	req.Port = flags.Int("port", 80, "Optional. The port of your real server on the backend node listening on")
	enable = flags.String("backend-mode", "enable", "Optional. Enable backend node or not. Accept values: enable, disable")
	weight = flags.Int("weight", 1, "Optional. effective for lb-method WeightRoundrobin. Rnage [0,100]")

	flags.SetFlagValues("resource-type", "Uhost", "UPM", "UDHost", "UDocker")
	flags.SetFlagValues("backend-mode", "enable", "disable")
	flags.SetFlagValuesFunc("ulb-id", func() []string {
		return getAllULBIDNames(*req.ProjectId, *req.Region)
	})
	flags.SetFlagValuesFunc("vserver-id", func() []string {
		ulbID := base.PickResourceID(*req.ULBId)
		return getAllULBVServerIDNames(ulbID, *req.ProjectId, *req.Region)
	})

	cmd.MarkFlagRequired("ulb-id")
	cmd.MarkFlagRequired("vserver-id")
	cmd.MarkFlagRequired("resource-id")
	return cmd
}

//NewCmdULBVServerUpdateNode ucloud ulb-vserver update-node
func NewCmdULBVServerUpdateNode(out io.Writer) *cobra.Command {
	var mode *string
	var weight *int
	backendIDs := []string{}
	req := base.BizClient.NewUpdateBackendAttributeRequest()
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update attributes of ULB backend nodes",
		Long:  "Update attributes of ULB backend nodes",
		Run: func(c *cobra.Command, args []string) {
			if *mode == "enable" {
				req.Enabled = sdk.Int(1)
			} else if *mode == "disable" {
				req.Enabled = sdk.Int(0)
			} else if *mode == "" {
				req.Enabled = nil
			} else {
				fmt.Fprintln(out, "Error, backend-mode must be enable or disable")
				return
			}
			if *weight != -1 && (*weight < 0 || *weight > 100) {
				fmt.Fprintln(out, "Error, weight must be between 0 and 100")
				return
			}
			if *weight != -1 {
				req.Weight = weight
			}

			if *req.Port == 0 {
				req.Port = nil
			}
			req.ULBId = sdk.String(base.PickResourceID(*req.ULBId))
			req.ProjectId = sdk.String(base.PickResourceID(*req.ProjectId))
			for _, bid := range backendIDs {
				req.BackendId = sdk.String(base.PickResourceID(bid))
				_, err := base.BizClient.UpdateBackendAttribute(req)
				if err != nil {
					base.HandleError(err)
					continue
				}
				fmt.Fprintf(out, "backend node[%s] updated\n", bid)
			}
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	req.ULBId = flags.String("ulb-id", "", "Required. Resource ID of ULB which the backend nodes belong to")
	flags.StringSliceVar(&backendIDs, "backend-id", nil, "Required. BackendID of backend nodes to update")
	req.Port = flags.Int("port", 0, "Optional. Port of your real server listening on backend nodes to update. Rnage [1,65535]")
	mode = flags.String("backend-mode", "", "Optional. Enable backend node or not. Accept values: enable, disable")
	weight = flags.Int("weight", -1, "Optional. effective for lb-method WeightRoundrobin. Rnage [0,100], -1 meaning no update")

	bindRegion(req, flags)
	bindProjectID(req, flags)

	flags.SetFlagValues("backend-mode", "enable", "disable")
	flags.SetFlagValuesFunc("ulb-id", func() []string {
		return getAllULBIDNames(*req.ProjectId, *req.Region)
	})
	flags.SetFlagValuesFunc("backend-id", func() []string {
		return getAllULBVServerNodeIDNames(*req.ULBId, "", *req.ProjectId, *req.Region)
	})

	cmd.MarkFlagRequired("ulb-id")
	cmd.MarkFlagRequired("backend-id")

	return cmd
}

//NewCmdULBVServerDeleteNode ucloud ulb-vserver delete-node
func NewCmdULBVServerDeleteNode(out io.Writer) *cobra.Command {
	backendIDs := []string{}
	req := base.BizClient.NewReleaseBackendRequest()
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete ULB VServer backend nodes",
		Long:  "Delete ULB VServer backend nodes",
		Run: func(c *cobra.Command, args []string) {
			req.ProjectId = sdk.String(base.PickResourceID(*req.ProjectId))
			req.ULBId = sdk.String(base.PickResourceID(*req.ULBId))
			for _, idname := range backendIDs {
				req.BackendId = sdk.String(base.PickResourceID(idname))
				_, err := base.BizClient.ReleaseBackend(req)
				if err != nil {
					base.HandleError(err)
					continue
				}
				fmt.Fprintf(out, "backend node[%s] deleted\n", idname)
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.ULBId = flags.String("ulb-id", "", "Required. Resource ID of ULB which the backend nodes belong to")
	flags.StringSliceVar(&backendIDs, "backend-id", nil, "Required. BackendID of backend nodes to update")
	bindRegion(req, flags)
	bindProjectID(req, flags)

	cmd.MarkFlagRequired("ulb-id")
	cmd.MarkFlagRequired("backend-id")

	flags.SetFlagValuesFunc("ulb-id", func() []string {
		return getAllULBIDNames(*req.ProjectId, *req.Region)
	})
	flags.SetFlagValuesFunc("backend-id", func() []string {
		return getAllULBVServerNodeIDNames(*req.ULBId, "", *req.ProjectId, *req.Region)
	})
	return cmd
}

//NewCmdULBVServerPolicy ucloud ulb vserver policy
func NewCmdULBVServerPolicy() *cobra.Command {
	out := base.Cxt.GetWriter()
	cmd := &cobra.Command{
		Use:   "policy",
		Short: "List and manipulate forward policy for VServer",
		Long:  "List and manipulate forward policy for VServer",
	}
	cmd.AddCommand(NewCmdULBVServerCreatePolicy(out))
	cmd.AddCommand(NewCmdULBVServerListPolicy(out))
	cmd.AddCommand(NewCmdULBVServerUpdatePolicy(out))
	cmd.AddCommand(NewCmdULBVServerDeletePolicy(out))
	return cmd
}

//NewCmdULBVServerCreatePolicy ucloud ulb-vserver create-policy
func NewCmdULBVServerCreatePolicy(out io.Writer) *cobra.Command {
	backendIDs := []string{}
	req := base.BizClient.NewCreatePolicyRequest()
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add content forward policy for VServer",
		Long:  "Add content forward policy for VServer",
		Run: func(c *cobra.Command, args []string) {
			if *req.Type != "Domain" && *req.Type != "Path" {
				fmt.Fprintln(out, "Error, forward method must be Domain or Path")
				return
			}
			req.ProjectId = sdk.String(base.PickResourceID(*req.ProjectId))
			req.ULBId = sdk.String(base.PickResourceID(*req.ULBId))
			req.VServerId = sdk.String(base.PickResourceID(*req.VServerId))
			for _, idname := range backendIDs {
				req.BackendId = append(req.BackendId, base.PickResourceID(idname))
			}
			resp, err := base.BizClient.CreatePolicy(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			fmt.Fprintf(out, "policy[%s] created\n", resp.PolicyId)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.ULBId = flags.String("ulb-id", "", "Required. Resource ID of ULB")
	req.VServerId = flags.String("vserver-id", "", "Required. Resource ID of VServer")
	flags.StringSliceVar(&backendIDs, "backend-id", nil, "Required. BackendID of the VServer's backend nodes")
	req.Type = flags.String("forward-method", "", "Required. Forward method, accept values:Domain and Path; Both forwarding methods can be described by using regular expressions or wildcards")
	req.Match = flags.String("expression", "", "Required. Expression of domain or path, such as \"www.[123].demo.com\" or \"/path/img/*.jpg\"")
	bindRegion(req, flags)
	bindProjectID(req, flags)

	flags.SetFlagValues("forward-method", "Domain", "Path")
	flags.SetFlagValuesFunc("ulb-id", func() []string {
		return getAllULBIDNames(*req.ProjectId, *req.Region)
	})
	flags.SetFlagValuesFunc("vserver-id", func() []string {
		return getAllULBVServerIDNames(*req.ULBId, *req.ProjectId, *req.Region)
	})
	flags.SetFlagValuesFunc("backend-id", func() []string {
		return getAllULBVServerNodeIDNames(*req.ULBId, *req.VServerId, *req.ProjectId, *req.Region)
	})

	cmd.MarkFlagRequired("ulb-id")
	cmd.MarkFlagRequired("vserver-id")
	cmd.MarkFlagRequired("backend-id")
	cmd.MarkFlagRequired("forward-method")
	cmd.MarkFlagRequired("expression")

	return cmd
}

//ULBVServerPolicy 表格行
type ULBVServerPolicy struct {
	ForwardMethod string
	Expression    string
	PolicyID      string
	PolicyType    string
	Backends      string
}

//NewCmdULBVServerListPolicy ucloud ulb-vserver list-policy
func NewCmdULBVServerListPolicy(out io.Writer) *cobra.Command {
	var ulbID, vserverID *string
	region := base.ConfigIns.Region
	project := base.ConfigIns.ProjectID
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List content forward policies of the VServer instance",
		Long:  "List content forward policies of the VServer instance",
		Run: func(c *cobra.Command, args []string) {
			ulbID = sdk.String(base.PickResourceID(*ulbID))
			vserverID = sdk.String(base.PickResourceID(*vserverID))
			vsList, err := getAllULBVServer(*ulbID, *vserverID, project, region)
			if err != nil {
				base.HandleError(err)
				return
			}
			if len(vsList) == 1 {
				vs := vsList[0]
				list := []ULBVServerPolicy{}
				for _, p := range vs.PolicySet {
					row := ULBVServerPolicy{}
					row.ForwardMethod = p.Type
					row.Expression = p.Match
					row.PolicyID = p.PolicyId
					row.PolicyType = p.PolicyType
					nodes := []string{}
					for _, b := range p.BackendSet {
						nodes = append(nodes, fmt.Sprintf("%s|%s:%d|%s", b.BackendId, b.PrivateIP, b.Port, b.ResourceName))
					}
					row.Backends = strings.Join(nodes, ",")
					list = append(list, row)
				}
				base.PrintList(list, out)
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	bindRegionS(&region, flags)
	bindProjectIDS(&project, flags)

	ulbID = flags.String("ulb-id", "", "Required. Resource ID of ULB")
	vserverID = flags.String("vserver-id", "", "Required. Resource ID of VServer")

	flags.SetFlagValuesFunc("ulb-id", func() []string {
		return getAllULBIDNames(project, region)
	})
	flags.SetFlagValuesFunc("vserver-id", func() []string {
		ulb := base.PickResourceID(*ulbID)
		return getAllULBVServerIDNames(ulb, project, region)
	})
	cmd.MarkFlagRequired("ulb-id")
	cmd.MarkFlagRequired("vserver-id")
	return cmd
}

//NewCmdULBVServerUpdatePolicy ucloud ulb-vserver update-policy
func NewCmdULBVServerUpdatePolicy(out io.Writer) *cobra.Command {
	policyIDs := []string{}
	backendIDs := []string{}
	addBackendIDs := []string{}
	removeBackendIDs := []string{}
	req := base.BizClient.NewUpdatePolicyRequest()
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update content forward policies of ULB VServer",
		Long:  "Update content forward policies ULB VServer",
		Run: func(c *cobra.Command, args []string) {
			req.ProjectId = sdk.String(base.PickResourceID(*req.ProjectId))
			req.ULBId = sdk.String(base.PickResourceID(*req.ULBId))
			req.VServerId = sdk.String(base.PickResourceID(*req.VServerId))

			vsList, err := getAllULBVServer(*req.ULBId, *req.VServerId, *req.ProjectId, *req.Region)
			if err != nil {
				base.HandleError(err)
				return
			}
			vs := vsList[0]

			for _, policyID := range policyIDs {
				var policy *ulb.ULBPolicySet
				for _, p := range vs.PolicySet {
					if p.PolicyId == policyID {
						policy = &p
						break
					}
				}
				if policy == nil {
					fmt.Fprintf(out, "policy[%s] not found\n", *req.PolicyId)
					continue
				}
				req.PolicyId = sdk.String(policyID)
				if *req.Type == "" {
					req.Type = sdk.String(policy.Type)
				} else if *req.Type != "Domain" && *req.Type != "Path" {
					fmt.Fprintf(out, "Error, forward-method must be Domain or Path")
					continue
				}
				if *req.Match == "" {
					req.Match = sdk.String(policy.Match)
				}
				backendIDMap := map[string]bool{}
				if backendIDs == nil {
					for _, b := range policy.BackendSet {
						backendIDMap[b.BackendId] = true
					}
				} else {
					for _, bid := range backendIDs {
						backendIDMap[base.PickResourceID(bid)] = true
					}
				}
				for _, bid := range addBackendIDs {
					backendIDMap[base.PickResourceID(bid)] = true
				}
				for _, bid := range removeBackendIDs {
					backendIDMap[base.PickResourceID(bid)] = false
				}
				for bid, ok := range backendIDMap {
					if ok {
						req.BackendId = append(req.BackendId, bid)
					}
				}
				resp, err := base.BizClient.UpdatePolicy(req)
				if err != nil {
					base.HandleError(err)
					continue
				}
				fmt.Fprintf(out, "policy[%s] updated\n", resp.PolicyId)
			}
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	bindRegion(req, flags)
	bindProjectID(req, flags)
	req.ULBId = flags.String("ulb-id", "", "Required. Resource ID of ULB")
	req.VServerId = flags.String("vserver-id", "", "Required. Resource ID of VServer")
	flags.StringSliceVar(&policyIDs, "policy-id", nil, "Required. PolicyID of policies to update")
	flags.StringSliceVar(&backendIDs, "backend-id", nil, "Optional. BackendID of backend nodes. If assign this flag, it will rewrite all backend nodes of the policy")
	flags.StringSliceVar(&addBackendIDs, "add-backend-id", nil, "Optional. BackendID of backend nodes. Add backend nodes to the policy")
	flags.StringSliceVar(&removeBackendIDs, "remove-backend-id", nil, "Optional. BackendID of backend nodes. Remove those backend nodes from the policy")
	req.Type = flags.String("forward-method", "", "Optional. Forward method of policy, accept values:Domain and Path")
	req.Match = flags.String("expression", "", "Optional. Expression of domain or path, such as \"www.[123].demo.com\" or \"/path/img/*.jpg\"")

	cmd.MarkFlagRequired("ulb-id")
	cmd.MarkFlagRequired("vserver-id")
	cmd.MarkFlagRequired("policy-id")

	flags.SetFlagValues("forward-method", "Domain", "Path")
	flags.SetFlagValuesFunc("ulb-id", func() []string {
		project := base.PickResourceID(*req.ProjectId)
		return getAllULBIDNames(project, *req.Region)
	})
	flags.SetFlagValuesFunc("vserver-id", func() []string {
		return getAllULBVServerIDNames(*req.ULBId, *req.ProjectId, *req.Region)
	})
	flags.SetFlagValuesFunc("backend-id", func() []string {
		return getAllULBVServerNodeIDNames(*req.ULBId, *req.VServerId, *req.ProjectId, *req.Region)
	})
	flags.SetFlagValuesFunc("add-backend-id", func() []string {
		return getAllULBVServerNodeIDNames(*req.ULBId, *req.VServerId, *req.ProjectId, *req.Region)
	})
	flags.SetFlagValuesFunc("remove-backend-id", func() []string {
		return getAllULBVServerNodeIDNames(*req.ULBId, *req.VServerId, *req.ProjectId, *req.Region)
	})

	return cmd
}

//NewCmdULBVServerDeletePolicy ucloud ulb-vserver delete-policy
func NewCmdULBVServerDeletePolicy(out io.Writer) *cobra.Command {
	policyIDs := []string{}
	req := base.BizClient.NewDeletePolicyRequest()
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete content forward policies of ULB VServer",
		Long:  "Delete content forward policies of ULB VServer",
		Run: func(c *cobra.Command, args []string) {
			for _, p := range policyIDs {
				req.PolicyId = sdk.String(p)
				_, err := base.BizClient.DeletePolicy(req)
				if err != nil {
					base.HandleError(err)
					continue
				}
				fmt.Fprintf(out, "policy[%s] deleted\n", p)
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	bindRegion(req, flags)
	bindProjectID(req, flags)

	flags.StringSliceVar(&policyIDs, "policy-id", nil, "Required. PolicyID of policies to delete")
	req.VServerId = flags.String("vserver-id", "", "Optional. Resource ID of VServer")

	cmd.MarkFlagRequired("policy-id")

	return cmd
}

//NewCmdULBSSL ucloud ulb-ssl-certificate
func NewCmdULBSSL() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ssl",
		Short: "List and manipulate SSL Certificates for ULB",
		Long:  "List and manipulate SSL Certificates for ULB",
	}
	out := base.Cxt.GetWriter()
	cmd.AddCommand(NewCmdSSLList(out))
	cmd.AddCommand(NewCmdSSLDescribe(out))
	cmd.AddCommand(NewCmdSSLAdd(out))
	cmd.AddCommand(NewCmdSSLDelete(out))
	cmd.AddCommand(NewCmdSSLBind(out))
	cmd.AddCommand(NewCmdSSLUnbind(out))
	return cmd
}

//SSLCertificate 表格行
type SSLCertificate struct {
	Name         string
	ResourceID   string
	MD5          string
	BindResource string
	UploadTime   string
}

//NewCmdSSLList ucloud ulb-ssl-certificate list
func NewCmdSSLList(out io.Writer) *cobra.Command {
	req := base.BizClient.NewDescribeSSLRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List SSL Certificates",
		Long:  "List SSL Certificates",
		Run: func(c *cobra.Command, args []string) {
			req.ProjectId = sdk.String(base.PickResourceID(*req.ProjectId))
			resp, err := base.BizClient.DescribeSSL(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			rows := []SSLCertificate{}
			for _, ssl := range resp.DataSet {
				row := SSLCertificate{}
				row.Name = ssl.SSLName
				row.ResourceID = ssl.SSLId
				row.MD5 = ssl.HashValue
				row.UploadTime = base.FormatDateTime(ssl.CreateTime)
				targets := []string{}
				for _, t := range ssl.BindedTargetSet {
					item := fmt.Sprintf("%s/%s(%s/%s)", t.VServerId, t.VServerName, t.ULBId, t.ULBName)
					targets = append(targets, item)
				}
				row.BindResource = strings.Join(targets, ",")
				rows = append(rows, row)
			}
			base.PrintList(rows, out)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	bindRegion(req, flags)
	bindProjectID(req, flags)
	req.SSLId = flags.String("ssl-id", "", "Optional. ResouceID of ssl certificate to list")
	bindLimit(req, flags)
	bindOffset(req, flags)

	return cmd
}

//NewCmdSSLDescribe ucloud ulb-ssl-certificate describe
func NewCmdSSLDescribe(out io.Writer) *cobra.Command {
	req := base.BizClient.NewDescribeSSLRequest()
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Display all data associated with SSL Certificate",
		Long:  "Display all data associated with SSL Certificate",
		Run: func(c *cobra.Command, args []string) {
			req.SSLId = sdk.String(base.PickResourceID(*req.SSLId))
			req.ProjectId = sdk.String(base.PickResourceID(*req.ProjectId))
			resp, err := base.BizClient.DescribeSSL(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			if len(resp.DataSet) <= 0 {
				fmt.Fprintf(out, "ssl certificate[%s] is not exists\n", *req.SSLId)
				return
			}

			sslcf := resp.DataSet[0]
			targets := []string{}
			for _, t := range sslcf.BindedTargetSet {
				item := fmt.Sprintf("%s/%s-%s/%s", t.ULBId, t.ULBName, t.VServerId, t.VServerName)
				targets = append(targets, item)
			}
			rows := []base.DescribeTableRow{
				base.DescribeTableRow{
					Attribute: "ResourceID",
					Content:   sslcf.SSLId,
				},
				base.DescribeTableRow{
					Attribute: "Name",
					Content:   sslcf.SSLName,
				},
				base.DescribeTableRow{
					Attribute: "Type",
					Content:   sslcf.SSLType,
				},
				base.DescribeTableRow{
					Attribute: "UploadTime",
					Content:   base.FormatDateTime(sslcf.CreateTime),
				},
				base.DescribeTableRow{
					Attribute: "BindResource",
					Content:   strings.Join(targets, ","),
				},
				base.DescribeTableRow{
					Attribute: "MD5",
					Content:   sslcf.HashValue,
				},
				base.DescribeTableRow{
					Attribute: "Content",
					Content:   sslcf.SSLContent,
				},
			}
			base.PrintDescribe(rows, global.JSON)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	req.SSLId = flags.String("ssl-id", "", "Required. ResouceID of ssl certificate to describe")
	bindRegion(req, flags)
	bindProjectID(req, flags)
	flags.SetFlagValuesFunc("ssl-id", func() []string {
		return getAllSSLCertIDNames(*req.ProjectId, *req.Region)
	})
	cmd.MarkFlagRequired("ssl-id")
	return cmd
}

//NewCmdSSLAdd ucloud ulb-ssl-certificate add
func NewCmdSSLAdd(out io.Writer) *cobra.Command {
	var allPath, sitePath, keyPath, caPath *string
	req := base.BizClient.NewCreateSSLRequest()
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add SSL Certificate",
		Long:  "Add SSL Certificate",
		Run: func(c *cobra.Command, args []string) {
			if *allPath == "" && (*sitePath == "" || *keyPath == "") {
				fmt.Fprintln(out, "if all-in-one-file is omitted, site-certificate-file and private-key-file can't be empty")
				return
			}
			if *allPath != "" {
				content, err := readFile(*allPath)
				if err != nil {
					base.HandleError(err)
					return
				}
				req.SSLContent = &content
			}
			if *sitePath != "" {
				content, err := readFile(*sitePath)
				if err != nil {
					base.HandleError(err)
					return
				}
				req.UserCert = &content
			}
			if *keyPath != "" {
				content, err := readFile(*keyPath)
				if err != nil {
					base.HandleError(err)
					return
				}
				req.PrivateKey = &content
			}
			if *caPath != "" {
				content, err := readFile(*caPath)
				if err != nil {
					base.HandleError(err)
					return
				}
				req.CaCert = &content
			}

			req.ProjectId = sdk.String(base.PickResourceID(*req.ProjectId))
			resp, err := base.BizClient.CreateSSL(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			fmt.Fprintf(out, "ssl certificate[%s] added\n", resp.SSLId)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	bindRegion(req, flags)
	bindProjectID(req, flags)
	req.SSLName = flags.String("name", "", "Required. Name of ssl certificate to add")
	req.SSLType = flags.String("format", "Pem", "Optional. Format of ssl certificate")
	allPath = flags.String("all-in-one-file", "", "Optional. Path of file which contain the complete content of the SSL certificate, including the content of site certificate, the private key which encrypted the site certificate, and the CA certificate. ")
	sitePath = flags.String("site-certificate-file", "", "Optional. Path of user's certificate file, *.crt. Required if all-in-one-file is omitted")
	keyPath = flags.String("private-key-file", "", "Optional. Path of private key file, *.key. Required if all-in-one-file is omitted")
	caPath = flags.String("ca-certificate-file", "", "Optional. Path of CA certificate file, *.crt")
	cmd.MarkFlagRequired("name")
	flags.SetFlagValuesFunc("all-in-one-file", func() []string {
		return base.GetFileList("")
	})
	flags.SetFlagValuesFunc("private-key-file", func() []string {
		return base.GetFileList(".key")
	})
	flags.SetFlagValuesFunc("ca-certificate-file", func() []string {
		return base.GetFileList(".crt")
	})
	flags.SetFlagValuesFunc("site-certificate-file", func() []string {
		return base.GetFileList(".crt")
	})
	return cmd
}

//NewCmdSSLDelete ucloud ulb-ssl-certificate delete
func NewCmdSSLDelete(out io.Writer) *cobra.Command {
	var idNames []string
	req := base.BizClient.NewDeleteSSLRequest()
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete SSL Certificates by resource id(ssl id)",
		Long:  "Delete SSL Certificates by resource id(ssl id)",
		Run: func(c *cobra.Command, args []string) {
			req.ProjectId = sdk.String(base.PickResourceID(*req.ProjectId))
			for _, idname := range idNames {
				req.SSLId = sdk.String(base.PickResourceID(idname))
				_, err := base.BizClient.DeleteSSL(req)
				if err != nil {
					base.HandleError(err)
					continue
				}
				fmt.Fprintf(out, "ssl certificate[%s] deleted\n", idname)
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	bindRegion(req, flags)
	bindProjectID(req, flags)
	flags.StringSliceVar(&idNames, "ssl-id", nil, "Required. Resource ID of SSL Certificates to delete")
	flags.SetFlagValuesFunc("ssl-id", func() []string {
		return getAllSSLCertIDNames(*req.ProjectId, *req.Region)
	})
	return cmd
}

//NewCmdSSLBind ucloud ulb-ssl-certificate bind
func NewCmdSSLBind(out io.Writer) *cobra.Command {
	req := base.BizClient.NewBindSSLRequest()
	cmd := &cobra.Command{
		Use:   "bind",
		Short: "Bind SSL Certificate with VServer",
		Long:  "Bind SSL Certificate with VServer",
		Run: func(c *cobra.Command, args []string) {
			req.ProjectId = sdk.String(base.PickResourceID(*req.ProjectId))
			req.ULBId = sdk.String(base.PickResourceID(*req.ULBId))
			req.VServerId = sdk.String(base.PickResourceID(*req.VServerId))
			req.SSLId = sdk.String(base.PickResourceID(*req.SSLId))
			_, err := base.BizClient.BindSSL(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			fmt.Fprintf(out, "ssl certificate[%s] bind with vserver[%s] of ulb[%s]\n", *req.SSLId, *req.VServerId, *req.ULBId)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	bindRegion(req, flags)
	bindProjectID(req, flags)
	req.SSLId = flags.String("ssl-id", "", "Required. Resource ID of SSL Certificate to bind")
	req.ULBId = flags.String("ulb-id", "", "Required. Resource ID of ULB")
	req.VServerId = flags.String("vserver-id", "", "Required. Resource ID of VServer")
	flags.SetFlagValuesFunc("ssl-id", func() []string {
		return getAllSSLCertIDNames(*req.ProjectId, *req.Region)
	})
	flags.SetFlagValuesFunc("ulb-id", func() []string {
		return getAllULBIDNames(*req.ProjectId, *req.Region)
	})
	flags.SetFlagValuesFunc("vserver-id", func() []string {
		return getAllULBVServerIDNames(*req.ULBId, *req.ProjectId, *req.Region)
	})
	cmd.MarkFlagRequired("ssl-id")
	cmd.MarkFlagRequired("ulb-id")
	cmd.MarkFlagRequired("vserver-id")
	return cmd
}

//NewCmdSSLUnbind ucloud ulb-ssl-certificate unbind
func NewCmdSSLUnbind(out io.Writer) *cobra.Command {
	req := base.BizClient.NewUnbindSSLRequest()
	cmd := &cobra.Command{
		Use:   "unbind",
		Short: "Unbind SSL Certificate with VServer",
		Long:  "Unbind SSL Certificate with VServer",
		Run: func(c *cobra.Command, args []string) {
			req.ProjectId = sdk.String(base.PickResourceID(*req.ProjectId))
			req.ULBId = sdk.String(base.PickResourceID(*req.ULBId))
			req.VServerId = sdk.String(base.PickResourceID(*req.VServerId))
			req.SSLId = sdk.String(base.PickResourceID(*req.SSLId))
			_, err := base.BizClient.UnbindSSL(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			fmt.Fprintf(out, "ssl certificate[%s] unbind with vserver[%s] of ulb[%s]\n", *req.SSLId, *req.VServerId, *req.ULBId)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	bindRegion(req, flags)
	bindProjectID(req, flags)
	req.SSLId = flags.String("ssl-id", "", "Required. Resource ID of SSL Certificate to unbind")
	req.ULBId = flags.String("ulb-id", "", "Required. Resource ID of ULB")
	req.VServerId = flags.String("vserver-id", "", "Required. Resource ID of VServer")
	flags.SetFlagValuesFunc("ssl-id", func() []string {
		return getAllSSLCertIDNames(*req.ProjectId, *req.Region)
	})
	flags.SetFlagValuesFunc("ulb-id", func() []string {
		if *req.SSLId == "" {
			return getAllULBIDNames(*req.ProjectId, *req.Region)
		}
		cert, err := getSSLCertByID(*req.SSLId, *req.ProjectId, *req.Region)
		if err != nil {
			return nil
		}
		ulbs := []string{}
		for _, b := range cert.BindedTargetSet {
			ulbs = append(ulbs, fmt.Sprintf("%s/%s", b.ULBId, b.ULBName))
		}
		return ulbs
	})
	flags.SetFlagValuesFunc("vserver-id", func() []string {
		if *req.SSLId == "" {
			return getAllULBVServerIDNames(*req.ULBId, *req.ProjectId, *req.Region)
		}
		cert, err := getSSLCertByID(*req.SSLId, *req.ProjectId, *req.Region)
		if err != nil {
			return nil
		}
		vservers := []string{}
		for _, b := range cert.BindedTargetSet {
			vservers = append(vservers, fmt.Sprintf("%s/%s", b.VServerId, b.VServerName))
		}
		return vservers
	})
	cmd.MarkFlagRequired("ssl-id")
	cmd.MarkFlagRequired("ulb-id")
	cmd.MarkFlagRequired("vserver-id")
	return cmd
}

func readFile(file string) (string, error) {
	byts, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}
	return string(byts), nil
}

func getAllULBVServerNodes(ulbID, vserverID, project, region string) ([]ulb.ULBBackendSet, error) {
	vsList, err := getAllULBVServer(ulbID, vserverID, project, region)
	if err != nil {
		return nil, err
	}
	nodeList := []ulb.ULBBackendSet{}
	for _, vs := range vsList {
		nodeList = append(nodeList, vs.BackendSet...)
	}
	return nodeList, nil
}

func getAllULBVServerNodeIDNames(ulbID, vserverID, project, region string) []string {
	nodeList, err := getAllULBVServerNodes(ulbID, vserverID, project, region)
	if err != nil {
		return nil
	}
	idNames := []string{}
	for _, node := range nodeList {
		idNames = append(idNames, fmt.Sprintf("%s/%s", node.BackendId, node.ResourceName))
	}
	return idNames
}

func getAllSSLCertIDNames(project, region string) []string {
	sslcs, err := getAllSSLCerts(project, region)
	if err != nil {
		return nil
	}
	idNames := []string{}
	for _, ssl := range sslcs {
		idNames = append(idNames, fmt.Sprintf("%s/%s", ssl.SSLId, ssl.SSLName))
	}
	return idNames
}

func getAllSSLCerts(project, region string) ([]ulb.ULBSSLSet, error) {
	req := base.BizClient.NewDescribeSSLRequest()
	req.ProjectId = sdk.String(base.PickResourceID(project))
	req.Region = sdk.String(region)
	list := []ulb.ULBSSLSet{}
	for offset, limit := 0, 50; ; offset += limit {
		req.Offset = sdk.Int(offset)
		req.Limit = sdk.Int(limit)
		resp, err := base.BizClient.DescribeSSL(req)
		if err != nil {
			return nil, err
		}
		list = append(list, resp.DataSet...)
		if resp.TotalCount <= offset+limit {
			break
		}
	}
	return list, nil
}

func getSSLCertByID(sslID, project, region string) (*ulb.ULBSSLSet, error) {
	if sslID == "" {
		return nil, fmt.Errorf("ssl certificate resource id can't be empty")
	}
	req := base.BizClient.NewDescribeSSLRequest()
	req.ProjectId = sdk.String(base.PickResourceID(project))
	req.Region = sdk.String(region)
	req.SSLId = sdk.String(base.PickResourceID(sslID))
	resp, err := base.BizClient.DescribeSSL(req)
	if err != nil {
		return nil, err
	}
	if len(resp.DataSet) <= 0 {
		return nil, fmt.Errorf("ssl certificate[%s] is not exists", sslID)
	}
	return &resp.DataSet[0], nil
}

func getAllULBVServer(ulbID, vserverID, project, region string) ([]ulb.ULBVServerSet, error) {
	req := base.BizClient.NewDescribeVServerRequest()
	req.ULBId = sdk.String(base.PickResourceID(ulbID))
	req.ProjectId = sdk.String(base.PickResourceID(project))
	req.Region = &region
	if vserverID != "" {
		req.VServerId = sdk.String(base.PickResourceID(vserverID))
	}
	resp, err := base.BizClient.DescribeVServer(req)
	if err != nil {
		return nil, err
	}
	if vserverID != "" {
		if len(resp.DataSet) < 1 {
			return nil, fmt.Errorf("VServer[%s] may not exist", vserverID)
		} else if len(resp.DataSet) > 1 {
			return nil, fmt.Errorf("Internal Error, too many vserver:%#v", resp.DataSet)
		}
	}
	return resp.DataSet, nil
}

func getAllULBVServerIDNames(ulbID, project, region string) []string {
	vservers, err := getAllULBVServer(ulbID, "", project, region)
	if err != nil {
		return nil
	}
	idNames := []string{}
	for _, vs := range vservers {
		idNames = append(idNames, fmt.Sprintf("%s/%s", vs.VServerId, vs.VServerName))
	}
	return idNames
}
