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
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/unet"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/base"
	"github.com/ucloud/ucloud-cli/model/status"
)

//NewCmdEIP ucloud eip
func NewCmdEIP() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "eip",
		Short: "List,allocate and release EIP",
		Long:  `Manipulate EIP, such as list,allocate and release`,
		Args:  cobra.NoArgs,
	}
	out := base.Cxt.GetWriter()
	cmd.AddCommand(NewCmdEIPList(out))
	cmd.AddCommand(NewCmdEIPAllocate())
	cmd.AddCommand(NewCmdEIPRelease())
	cmd.AddCommand(NewCmdEIPBind())
	cmd.AddCommand(NewCmdEIPUnbind())
	cmd.AddCommand(NewCmdEIPModifyBandwidth())
	cmd.AddCommand(NewCmdEIPSetChargeMode())
	cmd.AddCommand(NewCmdEIPJoinSharedBW())
	cmd.AddCommand(NewCmdEIPLeaveSharedBW())
	return cmd
}

//EIPRow 表格行
type EIPRow struct {
	Name           string
	IP             string
	ResourceID     string
	Group          string
	ChargeMode     string
	Bandwidth      string
	BindResource   string
	Status         string
	ExpirationTime string
}

//NewCmdEIPList ucloud eip list
func NewCmdEIPList(out io.Writer) *cobra.Command {
	req := base.BizClient.NewDescribeEIPRequest()
	fetchAll := false
	pageOff := false
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List all EIP instances",
		Long:    `List all EIP instances`,
		Example: "ucloud eip list",
		Run: func(cmd *cobra.Command, args []string) {
			var eipList []unet.UnetEIPSet
			if fetchAll || pageOff {
				list, err := fetchAllEip(*req.ProjectId, *req.Region)
				if err != nil {
					base.HandleError(err)
					return
				}
				eipList = list
			} else {
				resp, err := base.BizClient.DescribeEIP(req)
				if err != nil {
					base.HandleError(err)
					return
				}
				eipList = resp.EIPSet
			}

			list := make([]EIPRow, 0)
			for _, eip := range eipList {
				row := EIPRow{}
				row.Name = eip.Name
				for _, ip := range eip.EIPAddr {
					row.IP += ip.IP + " " + ip.OperatorName + "   "
				}
				row.ResourceID = eip.EIPId
				row.Group = eip.Tag
				row.ChargeMode = eip.PayMode
				row.Bandwidth = strconv.Itoa(eip.Bandwidth) + "Mb"
				if eip.Resource.ResourceId != "" {
					row.BindResource = fmt.Sprintf("%s|%s(%s)", eip.Resource.ResourceName, eip.Resource.ResourceId, eip.Resource.ResourceType)
				}
				row.Status = eip.Status
				row.ExpirationTime = time.Unix(int64(eip.ExpireTime), 0).Format("2006-01-02")
				list = append(list, row)
			}
			base.PrintList(list, out)
		},
	}

	flags := cmd.Flags()
	bindRegion(req, flags)
	bindProjectID(req, flags)
	req.Offset = flags.Int("offset", 0, "Optional. Offset default 0")
	req.Limit = flags.Int("limit", 50, "Optional. Limit default 50, max value 100")
	flags.BoolVar(&fetchAll, "list-all", false, "List all eip")
	flags.BoolVar(&pageOff, "page-off", false, "Optional. Paging or not. Accept values: true or false")
	flags.SetFlagValues("list-all", "true", "false")
	flags.MarkDeprecated("list-all", "please use '--page-off' instead")

	return cmd
}

func getEIPIDbyIP(ip net.IP, projectID, region string) (string, error) {
	eipList, err := fetchAllEip(projectID, region)
	if err != nil {
		return "", err
	}
	for _, eip := range eipList {
		for _, addr := range eip.EIPAddr {
			if addr.IP == ip.String() {
				return eip.EIPId, nil
			}
		}
	}
	return "", fmt.Errorf("IP[%s] not exist", ip.String())
}

func fetchAllEip(projectID, region string) ([]unet.UnetEIPSet, error) {
	req := base.BizClient.NewDescribeEIPRequest()
	list := []unet.UnetEIPSet{}
	req.ProjectId = sdk.String(projectID)
	req.Region = sdk.String(region)
	for offset, step := 0, 100; ; offset += step {
		req.Offset = &offset
		req.Limit = &step
		resp, err := base.BizClient.DescribeEIP(req)
		if err != nil {
			return nil, err
		}
		for i, size := 0, len(resp.EIPSet); i < size; i++ {
			list = append(list, resp.EIPSet[i])
		}
		if resp.TotalCount <= offset+step {
			break
		}
	}
	return list, nil
}

//states,paymodes 为nil时，不作为过滤条件
func getAllEip(projectID, region string, states, paymodes []string) []string {
	list, err := fetchAllEip(projectID, region)
	if err != nil {
		return nil
	}
	strs := []string{}
	for _, item := range list {
		rightState := false
		if states == nil {
			rightState = true
		} else {
			for _, s := range states {
				if item.Status == s {
					rightState = true
				}
			}
		}

		rightPayMode := false
		if paymodes == nil {
			rightPayMode = true
		} else {
			for _, m := range paymodes {
				if item.PayMode == m {
					rightPayMode = true
				}
			}
		}
		if !rightPayMode || !rightState {
			continue
		}

		ips := []string{}
		for _, ip := range item.EIPAddr {
			ips = append(ips, ip.IP)
		}
		strs = append(strs, item.EIPId+"/"+strings.Join(ips, ","))
	}
	return strs
}

func getEIP(eipID string) (*unet.UnetEIPSet, error) {
	req := base.BizClient.NewDescribeEIPRequest()
	req.EIPIds = append(req.EIPIds, eipID)
	resp, err := base.BizClient.DescribeEIP(req)
	if err != nil {
		return nil, err
	}
	if len(resp.EIPSet) == 1 {
		return &resp.EIPSet[0], nil
	}
	return nil, fmt.Errorf("eip[%s] may not exist", eipID)
}

//NewCmdEIPAllocate ucloud eip allocate
func NewCmdEIPAllocate() *cobra.Command {
	var count *int
	var req = base.BizClient.NewAllocateEIPRequest()
	var cmd = &cobra.Command{
		Use:     "allocate",
		Short:   "Allocate EIP",
		Long:    "Allocate EIP",
		Example: "ucloud eip allocate --line BGP --bandwidth-mb 2",
		Run: func(cmd *cobra.Command, args []string) {
			if *req.OperatorName == "" {
				*req.OperatorName = getEIPLine(*req.Region)
			}
			for i := 0; i < *count; i++ {
				resp, err := base.BizClient.AllocateEIP(req)
				if err != nil {
					base.HandleError(err)
					continue
				}
				for _, eip := range resp.EIPSet {
					base.Cxt.Printf("allocate EIP[%s] ", eip.EIPId)
					for _, ip := range eip.EIPAddr {
						base.Cxt.Printf("IP:%s  Line:%s \n", ip.IP, ip.OperatorName)
					}
				}
			}
		},
	}
	cmd.Flags().SortFlags = false
	req.Bandwidth = cmd.Flags().Int("bandwidth-mb", 0, "Required. Bandwidth(Unit:Mbps).The range of value related to network charge mode. By traffic [1, 200]; by bandwidth [1,800] (Unit: Mbps); it could be 0 if the eip belong to the shared bandwidth")
	req.OperatorName = cmd.Flags().String("line", "", "Optional. 'BGP' or 'International'. 'BGP' could be set in China mainland regions, such as cn-bj2 etc. 'International' could be set in the regions beyond mainland, such as hk, tw-kh, us-ws etc.")
	bindProjectID(req, cmd.Flags())
	bindRegion(req, cmd.Flags())
	req.PayMode = cmd.Flags().String("traffic-mode", "Bandwidth", "Optional. traffic-mode is an enumeration value. 'Traffic','Bandwidth' or 'ShareBandwidth'")
	req.ShareBandwidthId = cmd.Flags().String("share-bandwidth-id", "", "Optional. ShareBandwidthId, required only when traffic-mode is 'ShareBandwidth'")
	req.Quantity = cmd.Flags().Int("quantity", 1, "Optional. The duration of the instance. N years/months.")
	req.ChargeType = cmd.Flags().String("charge-type", "Month", "Optional. Enumeration value.'Year',pay yearly;'Month',pay monthly;'Dynamic', pay hourly(requires permission),'Trial', free trial(need permission)")
	req.Tag = cmd.Flags().String("group", "Default", "Optional. Group of your EIP.")
	req.Name = cmd.Flags().String("name", "EIP", "Optional. Name of your EIP.")
	req.Remark = cmd.Flags().String("remark", "", "Optional. Remark of your EIP.")
	count = cmd.Flags().Int("count", 1, "Optional. Count of EIP to allocate")

	cmd.Flags().SetFlagValues("line", "BGP", "International")
	cmd.Flags().SetFlagValues("traffic-mode", "Bandwidth", "Traffic", "ShareBandwidth")
	cmd.Flags().SetFlagValues("charge-type", "Month", "Year", "Dynamic", "Trial")
	cmd.MarkFlagRequired("bandwidth-mb")
	return cmd
}

//NewCmdEIPBind ucloud eip bind
func NewCmdEIPBind() *cobra.Command {
	var projectID, region, resourceID, resourceType *string
	var eipIDs []string
	cmd := &cobra.Command{
		Use:     "bind",
		Short:   "Bind EIP with uhost",
		Long:    "Bind EIP with uhost",
		Example: "ucloud eip bind --eip-id eip-xxx --resource-id uhost-xxx",
		Run: func(cmd *cobra.Command, args []string) {
			for _, eipID := range eipIDs {
				bindEIP(resourceID, resourceType, &eipID, projectID, region)
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	cmd.Flags().StringSliceVar(&eipIDs, "eip-id", nil, "Required. EIPId to bind")
	resourceID = cmd.Flags().String("resource-id", "", "Required. ResourceID , which is the UHostId of uhost")
	resourceType = cmd.Flags().String("resource-type", "uhost", "Requried. ResourceType, type of resource to bind with eip. 'uhost','vrouter','ulb','upm','hadoophost'.eg..")
	projectID = cmd.Flags().String("project-id", base.ConfigIns.ProjectID, "Optional. Assign project-id")
	region = cmd.Flags().String("region", base.ConfigIns.Region, "Optional. Assign region")

	cmd.Flags().SetFlagValues("resource-type", "uhost", "vrouter", "ulb", "upm", "hadoophost", "fortresshost", "udockhost", "udhost", "natgw", "udb", "vpngw", "ucdr", "dbaudit")
	cmd.Flags().SetFlagValuesFunc("eip-id", func() []string {
		return getAllEip(*projectID, *region, []string{status.EIP_FREE}, nil)
	})

	cmd.MarkFlagRequired("eip-id")
	cmd.MarkFlagRequired("resource-id")

	return cmd
}

func bindEIP(resourceID, resourceType, eipID, projectID, region *string) {
	ip := net.ParseIP(*eipID)
	if ip != nil {
		id, err := getEIPIDbyIP(ip, *projectID, *region)
		if err != nil {
			base.HandleError(err)
		} else {
			*eipID = id
		}
	}
	req := base.BizClient.NewBindEIPRequest()
	req.ResourceId = resourceID
	req.ResourceType = resourceType
	req.EIPId = sdk.String(base.PickResourceID(*eipID))
	req.ProjectId = sdk.String(base.PickResourceID(*projectID))
	req.Region = region
	_, err := base.BizClient.BindEIP(req)
	if err != nil {
		base.HandleError(err)
	} else {
		base.Cxt.Printf("bind EIP[%s] with %s[%s]\n", *req.EIPId, *req.ResourceType, *req.ResourceId)
	}
}

func sbindEIP(resourceID, resourceType, eipID, projectID, region *string) ([]string, error) {
	logs := make([]string, 0)
	ip := net.ParseIP(*eipID)
	if ip != nil {
		id, err := getEIPIDbyIP(ip, *projectID, *region)
		if err != nil {
			base.HandleError(err)
		} else {
			*eipID = id
		}
	}
	req := base.BizClient.NewBindEIPRequest()
	req.ResourceId = resourceID
	req.ResourceType = resourceType
	req.EIPId = sdk.String(base.PickResourceID(*eipID))
	req.ProjectId = sdk.String(base.PickResourceID(*projectID))
	req.Region = region
	logs = append(logs, fmt.Sprintf("api: BindEIP, request: %v", base.ToQueryMap(req)))
	_, err := base.BizClient.BindEIP(req)
	if err != nil {
		logs = append(logs, fmt.Sprintf("bind eip failed: %v", err))
		return logs, err
	}
	logs = append(logs, fmt.Sprintf("bind eip[%s] with %s[%s] successfully", *req.EIPId, *req.ResourceType, *req.ResourceId))
	return logs, nil
}

//NewCmdEIPUnbind ucloud eip unbind
func NewCmdEIPUnbind() *cobra.Command {
	eipIDs := []string{}
	req := base.BizClient.NewUnBindEIPRequest()
	cmd := &cobra.Command{
		Use:     "unbind",
		Short:   "Unbind EIP with uhost",
		Long:    "Unbind EIP with uhost",
		Example: "ucloud eip unbind --eip-id eip-xxx",
		Run: func(cmd *cobra.Command, args []string) {
			req.ProjectId = sdk.String(base.PickResourceID(*req.ProjectId))
			for _, eip := range eipIDs {
				eipIns, err := getEIP(base.PickResourceID(eip))
				if err != nil {
					base.HandleError(err)
					return
				}
				req.EIPId = sdk.String(base.PickResourceID(eip))
				req.ResourceId = sdk.String(eipIns.Resource.ResourceId)
				req.ResourceType = sdk.String(eipIns.Resource.ResourceType)
				_, err = base.BizClient.UnBindEIP(req)
				if err != nil {
					base.HandleError(err)
					return
				}
				base.Cxt.Printf("unbind EIP[%s] with %s[%s]\n", *req.EIPId, *req.ResourceType, *req.ResourceId)
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&eipIDs, "eip-id", nil, "Required. Resource ID of eips to unbind with some resource")
	bindRegion(req, flags)
	bindProjectID(req, flags)

	cmd.MarkFlagRequired("eip-id")
	cmd.Flags().SetFlagValuesFunc("eip-id", func() []string {
		return getAllEip(*req.ProjectId, *req.Region, []string{status.EIP_USED}, nil)
	})

	return cmd
}

func unbindEIP(resourceID, resourceType, eipID, projectID, region string) ([]string, error) {
	logs := make([]string, 0)
	eipID = base.PickResourceID(eipID)
	ip := net.ParseIP(eipID)
	if ip != nil {
		id, err := getEIPIDbyIP(ip, projectID, region)
		if err != nil {
			base.HandleError(err)
		} else {
			eipID = id
		}
	}
	req := base.BizClient.NewUnBindEIPRequest()
	req.ResourceId = &resourceID
	req.ResourceType = &resourceType
	req.EIPId = &eipID
	req.ProjectId = sdk.String(base.PickResourceID(projectID))
	req.Region = &region
	logs = append(logs, fmt.Sprintf("api: UnBindEIP, request: %v", base.ToQueryMap(req)))
	_, err := base.BizClient.UnBindEIP(req)
	if err != nil {
		logs = append(logs, fmt.Sprintf("unbind eip failed: %v", err))
		return logs, err
	}
	logs = append(logs, fmt.Sprintf("unbind eip[%s] with %s[%s] successfully", *req.EIPId, *req.ResourceType, *req.ResourceId))
	return logs, nil
}

//NewCmdEIPRelease ucloud eip release
func NewCmdEIPRelease() *cobra.Command {
	var ids []string
	req := base.BizClient.NewReleaseEIPRequest()
	cmd := &cobra.Command{
		Use:     "release",
		Short:   "Release EIP",
		Long:    "Release EIP",
		Example: "ucloud eip release --eip-id eip-xx1,eip-xx2",
		Run: func(cmd *cobra.Command, args []string) {
			req.ProjectId = sdk.String(base.PickResourceID(*req.ProjectId))
			for _, id := range ids {
				req.EIPId = sdk.String(base.PickResourceID(id))
				_, err := base.BizClient.ReleaseEIP(req)
				if err != nil {
					base.HandleError(err)
				} else {
					base.Cxt.Printf("eip[%s] released\n", *req.EIPId)
				}
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	flags.StringSliceVarP(&ids, "eip-id", "", nil, "Required. Resource ID of the EIPs you want to release")
	bindProjectID(req, flags)
	bindRegion(req, flags)
	cmd.MarkFlagRequired("eip-id")
	flags.SetFlagValuesFunc("eip-id", func() []string {
		return getAllEip(*req.ProjectId, *req.Region, []string{status.EIP_FREE}, nil)
	})

	return cmd
}

//NewCmdEIPModifyBandwidth ucloud eip modify-bw
func NewCmdEIPModifyBandwidth() *cobra.Command {
	ids := []string{}
	req := base.BizClient.NewModifyEIPBandwidthRequest()
	cmd := &cobra.Command{
		Use:     "modify-bw",
		Short:   "Modify bandwith of EIP instances",
		Long:    "Modify bandwith of EIP instances",
		Example: "ucloud eip modify-bw --eip-id eip-xxx --bandwidth-mb 20",
		// Deprecated: "use 'ucloud eip modiy'",
		Run: func(cmd *cobra.Command, args []string) {
			for _, id := range ids {
				id = base.PickResourceID(id)
				req.EIPId = &id
				_, err := base.BizClient.ModifyEIPBandwidth(req)
				if err != nil {
					base.HandleError(err)
				} else {
					base.Cxt.Printf("eip[%s]'s bandwidth modified\n", id)
				}
			}
		},
	}
	cmd.Flags().SortFlags = false
	cmd.Flags().StringSliceVarP(&ids, "eip-id", "", nil, "Required, Resource ID of EIPs to modify bandwidth")
	req.Bandwidth = cmd.Flags().Int("bandwidth-mb", 0, "Required. Bandwidth of EIP after modifed. Charge by traffic, range [1,300]; charge by bandwidth, range [1,800]")
	req.ProjectId = cmd.Flags().String("project-id", base.ConfigIns.ProjectID, "Optional. Assign project-id")
	req.Region = cmd.Flags().String("region", base.ConfigIns.Region, "Optional. Assign region")
	cmd.Flags().SetFlagValuesFunc("eip-id", func() []string {
		return getAllEip(*req.ProjectId, *req.Region, nil, nil)
	})
	cmd.MarkFlagRequired("eip-id")
	cmd.MarkFlagRequired("bandwidth-mb")
	return cmd
}

//NewCmdEIPSetChargeMode ucloud eip modify-traffic-mode
func NewCmdEIPSetChargeMode() *cobra.Command {
	ids := []string{}
	req := base.BizClient.NewSetEIPPayModeRequest()
	cmd := &cobra.Command{
		Use:     "modify-traffic-mode",
		Short:   "Modify charge mode of EIP instances",
		Long:    "Modify charge mode of EIP instances",
		Example: "ucloud eip modify-traffic-mode --eip-id eip-xx1,eip-xx2 --traffic-mode Traffic",
		Run: func(cmd *cobra.Command, args []string) {
			for _, id := range ids {
				id = base.PickResourceID(id)
				req.EIPId = &id
				eipIns, err := getEIP(id)
				if err != nil {
					base.HandleError(err)
					return
				}
				req.Bandwidth = sdk.Int(eipIns.Bandwidth)
				_, err = base.BizClient.SetEIPPayMode(req)
				if err != nil {
					base.HandleError(err)
				} else {
					base.Cxt.Printf("eip[%s]'s charge mode was modified to %s\n", id, *req.PayMode)
				}
			}
		},
	}

	cmd.Flags().SortFlags = false
	cmd.Flags().StringSliceVarP(&ids, "eip-id", "", nil, "Required, Resource ID of EIPs to modify charge mode")
	req.PayMode = cmd.Flags().String("traffic-mode", "", "Required, Charge mode of eip, 'Traffic','Bandwidth' or 'PostAccurateBandwidth'")
	req.ProjectId = cmd.Flags().String("project-id", base.ConfigIns.ProjectID, "Optional. Assign project-id")
	req.Region = cmd.Flags().String("region", base.ConfigIns.Region, "Optional. Assign region")
	cmd.Flags().SetFlagValues("traffic-mode", "Bandwidth", "Traffic", "PostAccurateBandwidth")
	cmd.Flags().SetFlagValuesFunc("eip-id", func() []string {
		return getAllEip(*req.ProjectId, *req.Region, nil, nil)
	})
	cmd.MarkFlagRequired("eip-id")
	cmd.MarkFlagRequired("traffic-mode")
	return cmd
}

//NewCmdEIPJoinSharedBW ucloud eip join-shared-bw
func NewCmdEIPJoinSharedBW() *cobra.Command {
	eipIDs := []string{}
	req := base.BizClient.NewAssociateEIPWithShareBandwidthRequest()
	cmd := &cobra.Command{
		Use:     "join-shared-bw",
		Short:   "Join shared bandwidth",
		Long:    "Join shared bandwidth",
		Example: "ucloud eip join-shared-bw --eip-id eip-xxx --shared-bw-id bwshare-xxx",
		Run: func(c *cobra.Command, args []string) {
			for _, eip := range eipIDs {
				req.EIPIds = append(req.EIPIds, base.PickResourceID(eip))
			}
			req.ShareBandwidthId = sdk.String(base.PickResourceID(*req.ShareBandwidthId))
			_, err := base.BizClient.AssociateEIPWithShareBandwidth(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			base.Cxt.Printf("eip%v joined shared bandwidth[%s]\n", req.EIPIds, *req.ShareBandwidthId)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	flags.StringSliceVar(&eipIDs, "eip-id", nil, "Required. Resource ID of EIPs to join shared bandwdith")
	req.ShareBandwidthId = flags.String("shared-bw-id", "", "Required. Resource ID of shared bandwidth to be joined")
	req.Region = flags.String("region", base.ConfigIns.Region, "Optional. Region, see 'ucloud region'")
	req.ProjectId = flags.String("project-id", base.ConfigIns.ProjectID, "Optional. Project-id, see 'ucloud project list'")
	flags.SetFlagValuesFunc("eip-id", func() []string {
		return getAllEip(*req.ProjectId, *req.Region, nil, []string{status.EIP_CHARGE_BANDWIDTH, status.EIP_CHARGE_TRAFFIC})
	})
	flags.SetFlagValuesFunc("shared-bw-id", func() []string {
		list, _ := getAllSharedBW(*req.ProjectId, *req.Region)
		return list
	})
	cmd.MarkFlagRequired("eip-id")
	cmd.MarkFlagRequired("shared-bw-id")

	return cmd
}

//NewCmdEIPLeaveSharedBW ucloud eip leave-shared-bw
func NewCmdEIPLeaveSharedBW() *cobra.Command {
	eipIDs := []string{}
	req := base.BizClient.NewDisassociateEIPWithShareBandwidthRequest()
	cmd := &cobra.Command{
		Use:     "leave-shared-bw",
		Short:   "Leave shared bandwidth",
		Long:    "Leave shared bandwidth",
		Example: "ucloud eip leave-shared-bw --eip-id eip-b2gvu3",
		Run: func(c *cobra.Command, args []string) {
			if *req.ShareBandwidthId == "" {
				for _, eipID := range eipIDs {
					eipIns, err := getEIP(base.PickResourceID(eipID))
					if err != nil {
						base.HandleError(err)
						continue
					}
					sharedBWID := eipIns.ShareBandwidthSet.ShareBandwidthId
					if sharedBWID == "" {
						base.Cxt.Printf("eip[%s] doesn't join any shared bandwidth\n", eipID)
						continue
					}
					req.ShareBandwidthId = sdk.String(sharedBWID)
					req.EIPIds = []string{base.PickResourceID(eipID)}
					_, err = base.BizClient.DisassociateEIPWithShareBandwidth(req)
					if err != nil {
						base.HandleError(err)
						continue
					}
					base.Cxt.Printf("eip[%s] left shared bandwidth[%s]\n", eipID, sharedBWID)
				}
			} else {
				for _, id := range eipIDs {
					req.EIPIds = append(req.EIPIds, base.PickResourceID(id))
				}
				*req.ShareBandwidthId = base.PickResourceID(*req.ShareBandwidthId)
				_, err := base.BizClient.DisassociateEIPWithShareBandwidth(req)
				if err != nil {
					base.HandleError(err)
					return
				}
				base.Cxt.Printf("eip%v left shared bandwidth[%s]\n", eipIDs, *req.ShareBandwidthId)
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	flags.StringSliceVar(&eipIDs, "eip-id", nil, "Required. Resource ID of EIPs to leave shared bandwidth")
	req.Bandwidth = flags.Int("bandwidth-mb", 1, "Required. Bandwidth of EIP after leaving shared bandwidth, ranging [1,300] for 'Traffic' charge mode, ranging [1,800] for 'Bandwidth' charge mode. Unit:Mb")
	req.PayMode = flags.String("traffic-mode", "Bandwidth", "Optional. Charge mode of the EIP after leaving shared bandwidth, 'Bandwidth' or 'Traffic'")
	req.ShareBandwidthId = flags.String("shared-bw-id", "", "Optional. Resource ID of shared bandwidth instance, assign this flag to make the operation faster")
	req.Region = flags.String("region", base.ConfigIns.Region, "Optional. Region, see 'ucloud region'")
	req.ProjectId = flags.String("project-id", base.ConfigIns.ProjectID, "Optional. Project-id, see 'ucloud project list'")

	flags.SetFlagValues("traffic-mode", "Bandwidth", "Traffic")
	flags.SetFlagValuesFunc("eip-id", func() []string {
		return getAllEip(*req.ProjectId, *req.Region, nil, []string{status.EIP_CHARGE_SHARE})
	})
	flags.SetFlagValuesFunc("shared-bw-id", func() []string {
		list, _ := getAllSharedBW(*req.ProjectId, *req.Region)
		return list
	})

	cmd.MarkFlagRequired("bandwidth")
	cmd.MarkFlagRequired("eip-id")
	return cmd
}
