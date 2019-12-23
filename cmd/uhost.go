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
	"encoding/base64"
	"fmt"
	"io"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/uhost"
	"github.com/ucloud/ucloud-sdk-go/services/unet"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/base"
	"github.com/ucloud/ucloud-cli/model/cli"
	"github.com/ucloud/ucloud-cli/model/status"
	"github.com/ucloud/ucloud-cli/ux"
)

var uhostSpoller = base.NewSpoller(sdescribeUHostByID, base.Cxt.GetWriter())

//NewCmdUHost ucloud uhost
func NewCmdUHost() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uhost",
		Short: "List,create,delete,stop,restart,poweroff or resize UHost instance",
		Long:  `List,create,delete,stop,restart,poweroff or resize UHost instance`,
		Args:  cobra.NoArgs,
	}
	out := base.Cxt.GetWriter()
	cmd.AddCommand(NewCmdUHostList(out))
	cmd.AddCommand(NewCmdUHostCreate())
	cmd.AddCommand(NewCmdUHostDelete(out))
	cmd.AddCommand(NewCmdUHostStop(out))
	cmd.AddCommand(NewCmdUHostStart(out))
	cmd.AddCommand(NewCmdUHostReboot(out))
	cmd.AddCommand(NewCmdUHostPoweroff(out))
	cmd.AddCommand(NewCmdUHostResize(out))
	cmd.AddCommand(NewCmdUHostClone(out))
	cmd.AddCommand(NewCmdUhostResetPassword(out))
	cmd.AddCommand(NewCmdUhostReinstallOS(out))
	cmd.AddCommand(NewCmdUhostCreateImage(out))
	cmd.AddCommand(NewCmdIsolation(out))
	cmd.AddCommand(NewCmdUhostLeaveIsolationGroup(out))

	return cmd
}

//UHostRow UHost表格行
type UHostRow struct {
	UHostName    string
	Remark       string
	ResourceID   string
	Group        string
	PrivateIP    string
	PublicIP     string
	Config       string
	DiskSet      string
	Zone         string
	Image        string
	VPC          string
	Subnet       string
	Type         string
	State        string
	CreationTime string
}

func listUhost(uhosts []uhost.UHostInstanceSet, out io.Writer, output string) {
	list := make([]UHostRow, 0)
	for _, host := range uhosts {
		row := UHostRow{}
		row.UHostName = host.Name
		row.Remark = host.Remark
		row.ResourceID = host.UHostId
		row.Group = host.Tag
		for _, ip := range host.IPSet {
			if row.PublicIP != "" {
				row.PublicIP += " | "
			}
			if ip.Type == "Private" {
				row.PrivateIP = ip.IP
				row.VPC = ip.VPCId
				row.Subnet = ip.SubnetId
			} else {
				row.PublicIP += fmt.Sprintf("%s", ip.IP)
			}
		}
		cupCore := host.CPU
		memorySize := host.Memory / 1024
		diskSize := 0
		var disks []string
		for _, disk := range host.DiskSet {
			if disk.Type == "Data" || disk.Type == "Udisk" {
				diskSize += disk.Size
			}
			disks = append(disks, fmt.Sprintf("%s:%s:%dG", disk.Type, disk.DiskType, disk.Size))
		}
		row.Zone = host.Zone
		row.DiskSet = strings.Join(disks, "|")
		row.Config = fmt.Sprintf("cpu:%d memory:%dG disk:%dG", cupCore, memorySize, diskSize)
		row.Image = fmt.Sprintf("%s|%s", host.BasicImageId, host.BasicImageName)
		row.CreationTime = base.FormatDate(host.CreateTime)
		row.State = host.State
		row.Type = host.MachineType + "/" + host.HostType
		if host.HotplugFeature {
			row.Type += "/HotPlug"
		}
		list = append(list, row)
	}
	if global.JSON {
		base.PrintJSON(list, out)
	} else {
		var cols []string
		if output == "wide" {
			cols = []string{"UHostName", "Remark", "ResourceID", "Group", "PrivateIP", "PublicIP", "Config", "DiskSet", "Zone", "Image", "VPC", "Subnet", "Type", "State", "CreationTime"}
		} else {
			cols = []string{"UHostName", "ResourceID", "Group", "PrivateIP", "PublicIP", "Config", "Image", "Type", "State", "CreationTime"}
		}
		base.PrintTable(list, cols)
	}
}

func listUhostID(uhosts []uhost.UHostInstanceSet, out io.Writer) {
	ids := make([]string, 0)
	for _, u := range uhosts {
		ids = append(ids, u.UHostId)
	}
	fmt.Fprintln(out, strings.Join(ids, ","))
}

func fetchUHosts(req *uhost.DescribeUHostInstanceRequest) ([]uhost.UHostInstanceSet, int, error) {
	resp, err := base.BizClient.DescribeUHostInstance(req)
	if err != nil {
		return nil, 0, err
	}
	return resp.UHostSet, resp.TotalCount, nil
}

func fetchUHostsPageOff(req *uhost.DescribeUHostInstanceRequest) ([]uhost.UHostInstanceSet, error) {
	_req := *req
	result := make([]uhost.UHostInstanceSet, 0)
	for limit, offset := 50, 0; ; offset += limit {
		_req.Offset = sdk.Int(offset)
		_req.Limit = sdk.Int(limit)
		uhosts, total, err := fetchUHosts(&_req)
		if err != nil {
			return nil, err
		}
		result = append(result, uhosts...)
		if offset+limit >= total {
			break
		}
	}
	return result, nil
}

func getAllUHosts(req *uhost.DescribeUHostInstanceRequest, pageOff bool, allRegion bool) ([]uhost.UHostInstanceSet, error) {
	if allRegion {
		result := make([]uhost.UHostInstanceSet, 0)
		regions, err := getAllRegions()
		if err != nil {
			return nil, err
		}
		for _, region := range regions {
			_req := *req
			_req.Region = sdk.String(region)
			//如果要获取所有region的主机，则不分页
			uhosts, err := fetchUHostsPageOff(&_req)
			if err != nil {
				return nil, err
			}
			result = append(result, uhosts...)
		}
		return result, nil
	}

	if pageOff {
		_req := *req
		uhosts, err := fetchUHostsPageOff(&_req)
		if err != nil {
			return nil, err
		}
		return uhosts, nil
	}

	uhosts, _, err := fetchUHosts(req)
	if err != nil {
		return nil, err
	}
	return uhosts, nil
}

//NewCmdUHostList [ucloud uhost list]
func NewCmdUHostList(out io.Writer) *cobra.Command {
	var allRegion, pageOff, idOnly bool
	var output string
	var uhostIds []string
	req := base.BizClient.NewDescribeUHostInstanceRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all UHost Instances",
		Long:  `List all UHost Instances`,
		Run: func(cmd *cobra.Command, args []string) {
			*req.VPCId = base.PickResourceID(*req.VPCId)
			*req.SubnetId = base.PickResourceID(*req.SubnetId)
			*req.IsolationGroup = base.PickResourceID(*req.IsolationGroup)
			for _, uhost := range uhostIds {
				req.UHostIds = append(req.UHostIds, base.PickResourceID(uhost))
			}

			uhosts, err := getAllUHosts(req, pageOff, allRegion)
			if err != nil {
				base.HandleError(err)
				return
			}
			if idOnly {
				listUhostID(uhosts, out)
			} else {
				listUhost(uhosts, out, output)
			}
		},
	}
	cmd.Flags().SortFlags = false
	req.ProjectId = cmd.Flags().String("project-id", base.ConfigIns.ProjectID, "Optional. Assign project-id")
	req.Region = cmd.Flags().String("region", base.ConfigIns.Region, "Optional. Assign region.")
	req.Zone = cmd.Flags().String("zone", "", "Optional. Assign availability zone")
	req.Offset = cmd.Flags().Int("offset", 0, "Optional. Offset default 0")
	req.Limit = cmd.Flags().Int("limit", 50, "Optional. Limit default 50, max value 100")
	req.VPCId = cmd.Flags().String("vpc-id", "", "Optional. Resource ID of VPC. List uhost instances of the specified VPC")
	req.SubnetId = cmd.Flags().String("subnet-id", "", "Optional. Resource ID of Subnet. List uhost instances of the specified Subnet")
	req.IsolationGroup = cmd.Flags().String("isolation-group", "", "Optional. Resource ID of isolation group. List uhost instances of the specified isolation group")
	cmd.Flags().StringSliceVar(&uhostIds, "uhost-id", make([]string, 0), "Optional. Resource ID of uhost instances, multiple values separated by comma(without space)")
	cmd.Flags().BoolVar(&allRegion, "all-region", false, "Optional. Accpet values: true or false. List uhost instances of all regions when assigned true")
	cmd.Flags().BoolVar(&pageOff, "page-off", false, "Optional. Paging or not. If all-region is specified this flag will be true. Accept values: true or false. If assigned, the limit flag will be disabled and list all uhost instances")
	cmd.Flags().BoolVar(&idOnly, "uhost-id-only", false, "Optional. Just display resource id of uhost")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Optional. Accept values: wide. Display more information about uhost such as DiskSet and Zone")
	bindGroup(req, cmd.Flags())

	cmd.Flags().SetFlagValues("page-off", "true", "false")
	cmd.Flags().SetFlagValues("uhost-id-only", "true", "false")
	cmd.Flags().SetFlagValues("output", "wide")
	cmd.Flags().SetFlagValuesFunc("project-id", getProjectList)
	cmd.Flags().SetFlagValuesFunc("region", getRegionList)
	cmd.Flags().SetFlagValuesFunc("zone", func() []string {
		return getZoneList(req.GetRegion())
	})

	flags := cmd.Flags()
	flags.SetFlagValuesFunc("vpc-id", func() []string {
		return getAllVPCIdNames(*req.ProjectId, *req.Region)
	})
	flags.SetFlagValuesFunc("subnet-id", func() []string {
		return getAllSubnetIDNames(*req.VPCId, *req.ProjectId, *req.Region)
	})
	flags.SetFlagValuesFunc("isolation-group", func() []string {
		return getIsolationGroupList(*req.ProjectId, *req.Region)
	})
	flags.SetFlagValuesFunc("uhost-id", func() []string {
		return getUhostList(nil, *req.ProjectId, *req.Region, *req.Zone)
	})

	return cmd
}

//NewCmdUHostCreate [ucloud uhost create]
func NewCmdUHostCreate() *cobra.Command {
	var bindEipIDs []string
	var hotPlug string
	var async bool
	var count int
	var hotPlugImageFlag bool

	req := base.BizClient.NewCreateUHostInstanceRequest()
	eipReq := base.BizClient.NewAllocateEIPRequest()
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create UHost instance",
		Long:  "Create UHost instance",
		Run: func(cmd *cobra.Command, args []string) {
			*req.Memory *= 1024
			req.LoginMode = sdk.String("Password")
			req.ImageId = sdk.String(base.PickResourceID(*req.ImageId))
			req.VPCId = sdk.String(base.PickResourceID(*req.VPCId))
			req.SubnetId = sdk.String(base.PickResourceID(*req.SubnetId))
			req.SecurityGroupId = sdk.String(base.PickResourceID(*req.SecurityGroupId))
			req.IsolationGroup = sdk.String(base.PickResourceID(*req.IsolationGroup))
			if hotPlug == "true" {
				req.HotplugFeature = sdk.Bool(true)
				any, err := describeImageByID(*req.ImageId, *req.ProjectId, *req.Region, *req.Zone)
				if err != nil {
					base.LogError(fmt.Sprintf("check image support hot-plug failed: %v", err))
				} else {
					image, ok := any.(*uhost.UHostImageSet)
					if !ok {
						base.LogError(fmt.Sprintf("check image support hot-plug failed, image %s may not exist", *req.ImageId))
					}
					for _, feature := range image.Features {
						if feature == "HotPlug" {
							hotPlugImageFlag = true
						}
					}
				}
				if !hotPlugImageFlag {
					base.LogWarn(fmt.Sprintf("warning. image %s does not support hot-plug", *req.ImageId))
					req.HotplugFeature = sdk.Bool(false)
				}
			}

			wg := &sync.WaitGroup{}
			tokens := make(chan struct{}, 10)
			wg.Add(count)
			if count <= 5 {
				for i := 0; i < count; i++ {
					bindEipID := ""
					if len(bindEipIDs) > i {
						bindEipID = bindEipIDs[i]
					}
					go createUhostWrapper(req, eipReq, bindEipID, async, make(chan bool, count), wg, tokens, i)
				}
			} else {
				retCh := make(chan bool, count)
				ux.Doc.Disable()
				refresh := ux.NewRefresh()

				go func() {
					for i := 0; i < count; i++ {
						bindEipID := ""
						if len(bindEipIDs) > i {
							bindEipID = bindEipIDs[i]
						}
						go createUhostWrapper(req, eipReq, bindEipID, async, retCh, wg, tokens, i)
					}
				}()

				go func() {
					var success, fail int
					refresh.Do(fmt.Sprintf("uhost creating, total:%d, success:%d, fail:%d", count, success, fail))
					for ret := range retCh {
						if ret {
							success++
						} else {
							fail++
						}
						refresh.Do(fmt.Sprintf("uhost creating, total:%d, success:%d, fail:%d", count, success, fail))
						if count == success+fail && fail > 0 {
							fmt.Printf("Check logs in %s\n", base.GetLogFilePath())
						}
					}
				}()
			}
			wg.Wait()
		},
	}

	req.Disks = make([]uhost.UHostDisk, 2)
	req.Disks[0].IsBoot = sdk.String("True")
	req.Disks[1].IsBoot = sdk.String("False")

	flags := cmd.Flags()
	flags.SortFlags = false
	req.CPU = flags.Int("cpu", 4, "Required. The count of CPU cores. Optional parameters: {1, 2, 4, 8, 12, 16, 24, 32}")
	req.Memory = flags.Int("memory-gb", 8, "Required. Memory size. Unit: GB. Range: [1, 128], multiple of 2")
	req.Password = flags.String("password", "", "Required. Password of the uhost user(root/ubuntu)")
	req.ImageId = flags.String("image-id", "", "Required. The ID of image. see 'ucloud image list'")
	flags.BoolVar(&async, "async", false, "Optional. Do not wait for the long-running operation to finish.")
	flags.IntVar(&count, "count", 1, "Optional. Number of uhost to create.")
	req.VPCId = flags.String("vpc-id", "", "Optional. VPC ID. This field is required under VPC2.0. See 'ucloud vpc list'")
	req.SubnetId = flags.String("subnet-id", "", "Optional. Subnet ID. This field is required under VPC2.0. See 'ucloud subnet list'")
	req.Name = flags.String("name", "UHost", "Optional. UHost instance name")
	flags.StringSliceVar(&bindEipIDs, "bind-eip", nil, "Optional. Resource ID or IP Address of eip that will be bound to the new created uhost")
	eipReq.OperatorName = flags.String("create-eip-line", "", "Optional. BGP for regions in the chinese mainland and International for overseas regions")
	eipReq.Bandwidth = flags.Int("create-eip-bandwidth-mb", 0, "Optional. Required if you want to create new EIP. Bandwidth(Unit:Mbps).The range of value related to network charge mode. By traffic [1, 300]; by bandwidth [1,800] (Unit: Mbps); it could be 0 if the eip belong to the shared bandwidth")
	eipReq.PayMode = flags.String("create-eip-traffic-mode", "Bandwidth", "Optional. 'Traffic','Bandwidth' or 'ShareBandwidth'")
	eipReq.ShareBandwidthId = flags.String("shared-bw-id", "", "Optional. Resource ID of shared bandwidth. It takes effect when create-eip-traffic-mode is ShareBandwidth ")
	eipReq.Name = flags.String("create-eip-name", "", "Optional. Name of created eip to bind with the uhost")
	eipReq.Remark = flags.String("create-eip-remark", "", "Optional.Remark of your EIP.")

	req.ChargeType = flags.String("charge-type", "Month", "Optional.'Year',pay yearly;'Month',pay monthly;'Dynamic', pay hourly")
	req.Quantity = flags.Int("quantity", 1, "Optional. The duration of the instance. N years/months.")
	bindProjectID(req, flags)
	bindRegion(req, flags)
	bindZone(req, flags)

	req.MachineType = flags.String("machine-type", "", "Optional. Accept values: N, C, G, O. Forward to https://docs.ucloud.cn/api/uhost-api/uhost_type for details")
	req.MinimalCpuPlatform = flags.String("minimal-cpu-platform", "", "Optional. Accpet values: Intel/Auto, Intel/IvyBridge, Intel/Haswell, Intel/Broadwell, Intel/Skylake, Intel/Cascadelake")
	req.UHostType = flags.String("type", "", "Optional. Accept values: N1, N2, N3, G1, G2, G3, I1, I2, C1. Forward to https://docs.ucloud.cn/api/uhost-api/uhost_type for details")
	req.GPU = flags.Int("gpu", 0, "Optional. The count of GPU cores.")
	req.NetCapability = flags.String("net-capability", "Normal", "Optional. Default is 'Normal', also support 'Super' which will enhance multiple times network capability as before")
	flags.StringVar(&hotPlug, "hot-plug", "true", "Optional. Enable hot plug feature or not. Accept values: true or false")
	req.Disks[0].Type = flags.String("os-disk-type", "CLOUD_SSD", "Optional. Enumeration value. 'LOCAL_NORMAL', Ordinary local disk; 'CLOUD_NORMAL', Ordinary cloud disk; 'LOCAL_SSD',local ssd disk; 'CLOUD_SSD',cloud ssd disk; 'EXCLUSIVE_LOCAL_DISK',big data. The disk only supports a limited combination.")
	req.Disks[0].Size = flags.Int("os-disk-size-gb", 20, "Optional. Default 20G. Windows should be bigger than 40G Unit GB")
	req.Disks[0].BackupType = flags.String("os-disk-backup-type", "NONE", "Optional. Enumeration value, 'NONE' or 'DATAARK'. DataArk supports real-time backup, which can restore the disk back to any moment within the last 12 hours. (Normal Local Disk and Normal Cloud Disk Only)")
	req.Disks[1].Type = flags.String("data-disk-type", "CLOUD_SSD", "Optional. Enumeration value. 'LOCAL_NORMAL', Ordinary local disk; 'CLOUD_NORMAL', Ordinary cloud disk; 'LOCAL_SSD',local ssd disk; 'CLOUD_SSD',cloud ssd disk; 'EXCLUSIVE_LOCAL_DISK',big data. The disk only supports a limited combination.")
	req.Disks[1].Size = flags.Int("data-disk-size-gb", 20, "Optional. Disk size. Unit GB")
	req.Disks[1].BackupType = flags.String("data-disk-backup-type", "NONE", "Optional. Enumeration value, 'NONE' or 'DATAARK'. DataArk supports real-time backup, which can restore the disk back to any moment within the last 12 hours. (Normal Local Disk and Normal Cloud Disk Only)")
	req.SecurityGroupId = flags.String("firewall-id", "", "Optional. Firewall Id, default: Web recommended firewall. see 'ucloud firewall list'.")
	req.Tag = flags.String("group", "Default", "Optional. Business group")
	req.IsolationGroup = flags.String("isolation-group", "", "Optional. Resource ID of isolation group. see 'ucloud uhost isolation-group list")

	flags.MarkDeprecated("type", "please use --machine-type instead")
	flags.SetFlagValues("charge-type", "Month", "Year", "Dynamic", "Trial")
	flags.SetFlagValues("hot-plug", "true", "false")
	flags.SetFlagValues("cpu", "1", "2", "4", "8", "12", "16", "24", "32")
	flags.SetFlagValues("type", "N2", "N1", "N3", "I2", "I1", "C1", "G1", "G2", "G3")
	flags.SetFlagValues("machine-type", "N", "C", "G", "O")
	flags.SetFlagValues("minimal-cpu-platform", "Intel/Auto", "Intel/IvyBridge", "Intel/Haswell", "Intel/Broadwell", "Intel/Skylake", "Intel/Cascadelake")
	flags.SetFlagValues("net-capability", "Normal", "Super")
	flags.SetFlagValues("os-disk-type", "LOCAL_NORMAL", "CLOUD_NORMAL", "LOCAL_SSD", "CLOUD_SSD", "CLOUD_RSSD", "EXCLUSIVE_LOCAL_DISK")
	flags.SetFlagValues("os-disk-backup-type", "NONE", "DATAARK")
	flags.SetFlagValues("data-disk-type", "LOCAL_NORMAL", "CLOUD_NORMAL", "LOCAL_SSD", "CLOUD_SSD", "EXCLUSIVE_LOCAL_DISK")
	flags.SetFlagValues("data-disk-backup-type", "NONE", "DATAARK")
	flags.SetFlagValues("create-eip-line", "BGP", "International")
	flags.SetFlagValues("create-eip-traffic-mode", "Bandwidth", "Traffic", "ShareBandwidth")

	flags.SetFlagValuesFunc("image-id", func() []string {
		return getImageList([]string{status.IMAGE_AVAILABLE}, cli.IMAGE_BASE, *req.ProjectId, *req.Region, *req.Zone)
	})
	flags.SetFlagValuesFunc("vpc-id", func() []string {
		return getAllVPCIdNames(*req.ProjectId, *req.Region)
	})
	flags.SetFlagValuesFunc("bind-eip", func() []string {
		return getAllEip(*req.ProjectId, *req.Region, []string{status.EIP_FREE}, nil)
	})
	flags.SetFlagValuesFunc("firewall-id", func() []string {
		return getFirewallIDNames(*req.ProjectId, *req.Region)
	})
	flags.SetFlagValuesFunc("subnet-id", func() []string {
		return getAllSubnetIDNames(*req.VPCId, *req.ProjectId, *req.Region)
	})
	flags.SetFlagValuesFunc("isolation-group", func() []string {
		return getIsolationGroupList(*req.ProjectId, *req.Region)
	})

	cmd.MarkFlagRequired("cpu")
	cmd.MarkFlagRequired("memory-gb")
	cmd.MarkFlagRequired("password")
	cmd.MarkFlagRequired("image-id")

	return cmd
}

//createUhostWrapper 处理UI和并发控制
func createUhostWrapper(req *uhost.CreateUHostInstanceRequest, eipReq *unet.AllocateEIPRequest, bindEipID string, async bool, retCh chan<- bool, wg *sync.WaitGroup, tokens chan struct{}, idx int) {
	//控制并发数量
	tokens <- struct{}{}
	defer func() {
		<-tokens
		//设置延时，使报错能渲染出来
		time.Sleep(time.Second / 5)
		wg.Done()
	}()

	success, logs := createUhost(req, eipReq, bindEipID, async)
	retCh <- success
	logs = append(logs, fmt.Sprintf("index:%d, result:%t", idx, success))
	base.LogInfo(logs...)
}

func createUhost(req *uhost.CreateUHostInstanceRequest, eipReq *unet.AllocateEIPRequest, bindEipID string, async bool) (bool, []string) {
	resp, err := base.BizClient.CreateUHostInstance(req)
	block := ux.NewBlock()
	ux.Doc.Append(block)
	logs := []string{"=================================================="}
	logs = append(logs, fmt.Sprintf("api:CreateUHostInstance, request:%v", base.ToQueryMap(req)))
	if err != nil {
		logs = append(logs, fmt.Sprintf("err:%v", err))
		block.Append(base.ParseError(err))
		return false, logs
	}

	logs = append(logs, fmt.Sprintf("resp:%#v", resp))
	if len(resp.UHostIds) != 1 {
		block.Append(fmt.Sprintf("expect uhost count 1 , accept %d", len(resp.UHostIds)))
		return false, logs
	}

	text := fmt.Sprintf("uhost[%s] is initializing", resp.UHostIds[0])
	if async {
		block.Append(text)
	} else {
		uhostSpoller.Sspoll(resp.UHostIds[0], text, []string{status.HOST_RUNNING, status.HOST_FAIL}, block)
	}

	if bindEipID != "" {
		eip := base.PickResourceID(bindEipID)
		logs = append(logs, fmt.Sprintf("bind eip: %s", eip))
		eipLogs, err := sbindEIP(sdk.String(resp.UHostIds[0]), sdk.String("uhost"), &eip, req.ProjectId, req.Region)
		logs = append(logs, eipLogs...)
		if err != nil {
			block.Append(fmt.Sprintf("bind eip[%s] with uhost[%s] failed: %v", eip, resp.UHostIds[0], err))
			return false, logs
		}
		block.Append(fmt.Sprintf("bind eip[%s] with uhost[%s] successfully", eip, resp.UHostIds[0]))
	} else if *eipReq.Bandwidth != 0 {
		eipReq.ChargeType = req.ChargeType
		eipReq.Tag = req.Tag
		eipReq.Quantity = req.Quantity
		eipReq.Region = req.Region
		eipReq.ProjectId = req.ProjectId
		logs = append(logs, fmt.Sprintf("create eip request: %v", base.ToQueryMap(eipReq)))
		if *eipReq.OperatorName == "" {
			*eipReq.OperatorName = getEIPLine(*req.Region)
		}
		eipResp, err := base.BizClient.AllocateEIP(eipReq)

		if err != nil {
			logs = append(logs, fmt.Sprintf("create eip error: %#v", err))
			block.Append(base.ParseError(err))
		} else {
			logs = append(logs, fmt.Sprintf("create eip resp: %#v", eipResp))
			for _, eip := range eipResp.EIPSet {
				block.Append(fmt.Sprintf("allocate EIP[%s] ", eip.EIPId))
				for _, ip := range eip.EIPAddr {
					block.Append(fmt.Sprintf("IP:%s  Line:%s", ip.IP, ip.OperatorName))
				}
				if len(resp.UHostIds) == 1 {
					eipLogs, err := sbindEIP(sdk.String(resp.UHostIds[0]), sdk.String("uhost"), sdk.String(eip.EIPId), req.ProjectId, req.Region)
					logs = append(logs, eipLogs...)
					if err != nil {
						block.Append(fmt.Sprintf("bind eip[%s] with uhost[%s] failed: %v", eip, resp.UHostIds[0], err))
						return false, logs
					}
					block.Append(fmt.Sprintf("bind eip[%s] with uhost[%s] successfully", eip, resp.UHostIds[0]))
				}
			}
		}
	}
	return true, logs
}

//NewCmdUHostDelete ucloud uhost delete
func NewCmdUHostDelete(out io.Writer) *cobra.Command {
	var uhostIDs *[]string
	var isDestroy = sdk.Bool(false)
	var yes *bool

	req := base.BizClient.NewTerminateUHostInstanceRequest()
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete Uhost instance",
		Long:  "Delete Uhost instance",
		Run: func(cmd *cobra.Command, args []string) {
			if !*yes {
				sure, err := ux.Prompt("Are you sure you want to delete the host(s)?")
				if err != nil {
					base.Cxt.Println(err)
					return
				}
				if !sure {
					return
				}
			}
			if *isDestroy {
				req.Destroy = sdk.Int(1)
			} else {
				req.Destroy = sdk.Int(0)
			}

			reqs := make([]request.Common, len(*uhostIDs))
			for idx, id := range *uhostIDs {
				_req := *req
				id = base.PickResourceID(id)
				_req.UHostId = sdk.String(id)
				reqs[idx] = &_req
			}
			coAction := newConcurrentAction(reqs, deleteUHost)
			coAction.Do()
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	uhostIDs = cmd.Flags().StringSlice("uhost-id", nil, "Requried. ResourceIDs(UhostIds) of the uhost instance")
	bindRegion(req, flags)
	bindProjectID(req, flags)
	req.Zone = cmd.Flags().String("zone", "", "Optional. availability zone")
	isDestroy = cmd.Flags().Bool("destroy", false, "Optional. false,the uhost instance will be thrown to UHost recycle if you have permission; true,the uhost instance will be deleted directly")
	req.ReleaseEIP = cmd.Flags().Bool("release-eip", true, "Optional. false,Unbind EIP only; true, Unbind EIP and release it")
	req.ReleaseUDisk = cmd.Flags().Bool("delete-cloud-disk", false, "Optional. false, detach cloud disk only; true, detach cloud disk and delete it")
	yes = cmd.Flags().BoolP("yes", "y", false, "Optional. Do not prompt for confirmation.")
	cmd.Flags().SetFlagValues("destroy", "true", "false")
	cmd.Flags().SetFlagValues("release-eip", "true", "false")
	cmd.Flags().SetFlagValues("delete-cloud-disk", "true", "false")
	cmd.Flags().SetFlagValuesFunc("uhost-id", func() []string {
		return getUhostList([]string{status.HOST_RUNNING, status.HOST_STOPPED, status.HOST_FAIL}, *req.ProjectId, *req.Region, *req.Zone)
	})
	cmd.MarkFlagRequired("uhost-id")

	return cmd
}

func deleteUHost(creq request.Common) (bool, []string) {
	req := creq.(*uhost.TerminateUHostInstanceRequest)
	block := ux.NewBlock()
	ux.Doc.Append(block)
	logs := []string{}
	hostIns, err := sdescribeUHostByID(*req.UHostId)
	if err != nil {
		logs = append(logs, fmt.Sprintf("describe uhost[%s] failed: %s", *req.UHostId, base.ParseError(err)))
		return false, logs
	}

	if hostIns == nil {
		logs = append(logs, fmt.Sprintf("uhost[%s] does not exist", *req.UHostId))
		return false, logs
	}

	ins := hostIns.(*uhost.UHostInstanceSet)
	if ins.State == "Running" {
		_req := base.BizClient.NewStopUHostInstanceRequest()
		_req.ProjectId = req.ProjectId
		_req.Region = req.Region
		_req.Zone = req.Zone
		_req.UHostId = req.UHostId
		stopUhostInsV2(_req, false, block)
	}

	logs = append(logs, fmt.Sprintf("api:TerminateUHostInstance, request:%v", base.ToQueryMap(req)))
	resp, err := base.BizClient.TerminateUHostInstance(req)
	if err != nil {
		block.Append(base.ParseError(err))
		logs = append(logs, fmt.Sprintf("delete uhost[%s] failed: %s", *req.UHostId, base.ParseError(err)))
		return false, logs
	}
	text := fmt.Sprintf("uhost[%s] deleted", resp.UHostId)
	logs = append(logs, text)
	block.Append(text)
	return true, logs
}

//NewCmdUHostStop ucloud uhost stop
func NewCmdUHostStop(out io.Writer) *cobra.Command {
	var uhostIDs *[]string
	var async *bool
	req := base.BizClient.NewStopUHostInstanceRequest()
	cmd := &cobra.Command{
		Use:     "stop",
		Short:   "Shut down uhost instance",
		Long:    "Shut down uhost instance",
		Example: "ucloud uhost stop --uhost-id uhost-xxx1,uhost-xxx2",
		Run: func(cmd *cobra.Command, args []string) {
			for _, id := range *uhostIDs {
				id = base.PickResourceID(id)
				req.UHostId = &id
				stopUhostIns(req, *async, out)
			}
		},
	}
	cmd.Flags().SortFlags = false
	uhostIDs = cmd.Flags().StringSlice("uhost-id", nil, "Required. ResourceIDs(UHostIds) of the uhost instances")
	req.ProjectId = cmd.Flags().String("project-id", base.ConfigIns.ProjectID, "Optional. Assign project-id")
	req.Region = cmd.Flags().String("region", base.ConfigIns.Region, "Optional. Assign region")
	req.Zone = cmd.Flags().String("zone", "", "Optional. Assign availability zone")
	async = cmd.Flags().Bool("async", false, "Optional. Do not wait for the long-running operation to finish.")
	cmd.Flags().SetFlagValuesFunc("uhost-id", func() []string {
		return getUhostList([]string{status.HOST_RUNNING}, *req.ProjectId, *req.Region, *req.Zone)
	})
	cmd.MarkFlagRequired("uhost-id")

	return cmd
}

func stopUhostIns(req *uhost.StopUHostInstanceRequest, async bool, out io.Writer) {
	resp, err := base.BizClient.StopUHostInstance(req)
	if err != nil {
		base.HandleError(err)
	} else {
		text := fmt.Sprintf("uhost[%v] is shutting down", resp.UhostId)
		if async {
			fmt.Fprintln(out, text)
		} else {
			poller := base.NewPoller(describeUHostByID, out)
			poller.Poll(resp.UhostId, *req.ProjectId, *req.Region, *req.Zone, text, []string{status.HOST_STOPPED, status.HOST_FAIL})
		}
	}
}

//可并发调用版本
func stopUhostInsV2(req *uhost.StopUHostInstanceRequest, async bool, block *ux.Block) {
	resp, err := base.BizClient.StopUHostInstance(req)
	if err != nil {
		block.Append(base.ParseError(err))
		return
	}

	text := fmt.Sprintf("uhost[%v] is shutting down", resp.UhostId)
	if async {
		block.Append(text)
	} else {
		uhostSpoller.Sspoll(resp.UhostId, text, []string{status.HOST_STOPPED, status.HOST_FAIL}, block)
	}
}

//NewCmdUHostStart ucloud uhost start
func NewCmdUHostStart(out io.Writer) *cobra.Command {
	var async *bool
	var uhostIDs *[]string
	req := base.BizClient.NewStartUHostInstanceRequest()
	cmd := &cobra.Command{
		Use:     "start",
		Short:   "Start Uhost instance",
		Long:    "Start Uhost instance",
		Example: "ucloud uhost start --uhost-id uhost-xxx1,uhost-xxx2",
		Run: func(cmd *cobra.Command, args []string) {
			for _, id := range *uhostIDs {
				id := base.PickResourceID(id)
				req.UHostId = &id
				resp, err := base.BizClient.StartUHostInstance(req)
				if err != nil {
					base.HandleError(err)
				} else {
					text := fmt.Sprintf("uhost[%v] is starting", resp.UhostId)
					if *async {
						fmt.Fprintln(out, text)
					} else {
						poller := base.NewPoller(describeUHostByID, out)
						poller.Poll(resp.UhostId, *req.ProjectId, *req.Region, *req.Zone, text, []string{status.HOST_RUNNING, status.HOST_FAIL})
					}
				}
			}
		},
	}
	cmd.Flags().SortFlags = false
	uhostIDs = cmd.Flags().StringSlice("uhost-id", nil, "Requried. ResourceIDs(UHostIds) of the uhost instance")
	req.ProjectId = cmd.Flags().String("project-id", base.ConfigIns.ProjectID, "Optional. Assign project-id")
	req.Region = cmd.Flags().String("region", base.ConfigIns.Region, "Optional. Assign region")
	req.Zone = cmd.Flags().String("zone", "", "Optional. Assign availability zone")
	async = cmd.Flags().Bool("async", false, "Optional. Do not wait for the long-running operation to finish.")
	cmd.Flags().SetFlagValuesFunc("uhost-id", func() []string {
		return getUhostList([]string{status.HOST_STOPPED}, *req.ProjectId, *req.Region, *req.Zone)
	})
	cmd.MarkFlagRequired("uhost-id")
	return cmd
}

//NewCmdUHostReboot ucloud uhost restart
func NewCmdUHostReboot(out io.Writer) *cobra.Command {
	var uhostIDs *[]string
	var async *bool
	req := base.BizClient.NewRebootUHostInstanceRequest()
	cmd := &cobra.Command{
		Use:     "restart",
		Short:   "Restart uhost instance",
		Long:    "Restart uhost instance",
		Example: "ucloud uhost restart --uhost-id uhost-xxx1,uhost-xxx2",
		Run: func(cmd *cobra.Command, args []string) {
			for _, id := range *uhostIDs {
				id = base.PickResourceID(id)
				req.UHostId = &id
				resp, err := base.BizClient.RebootUHostInstance(req)
				if err != nil {
					base.HandleError(err)
				} else {
					text := fmt.Sprintf("uhost[%v] is restarting", resp.UhostId)
					if *async {
						fmt.Fprintln(out, text)
					} else {
						poller := base.NewPoller(describeUHostByID, out)
						poller.Poll(resp.UhostId, *req.ProjectId, *req.Region, *req.Zone, text, []string{status.HOST_RUNNING, status.HOST_FAIL})
					}
				}
			}
		},
	}
	cmd.Flags().SortFlags = false
	uhostIDs = cmd.Flags().StringSlice("uhost-id", nil, "Required. ResourceIDs(UHostIds) of the uhost instance")
	req.ProjectId = cmd.Flags().String("project-id", base.ConfigIns.ProjectID, "Optional. Assign project-id")
	req.Region = cmd.Flags().String("region", base.ConfigIns.Region, "Optional. Assign region")
	req.Zone = cmd.Flags().String("zone", "", "Optional. Assign availability zone")
	req.DiskPassword = cmd.Flags().String("disk-password", "", "Optional. Encrypted disk password")
	async = cmd.Flags().Bool("async", false, "Optional. Do not wait for the long-running operation to finish.")
	cmd.Flags().SetFlagValuesFunc("uhost-id", func() []string {
		return getUhostList([]string{status.HOST_FAIL, status.HOST_RUNNING, status.HOST_STOPPED}, *req.ProjectId, *req.Region, *req.Zone)
	})
	cmd.MarkFlagRequired("uhost-id")
	return cmd
}

//NewCmdUHostPoweroff ucloud uhost poweroff
func NewCmdUHostPoweroff(out io.Writer) *cobra.Command {
	var yes *bool
	var uhostIDs *[]string
	req := base.BizClient.NewPoweroffUHostInstanceRequest()
	cmd := &cobra.Command{
		Use:     "poweroff",
		Short:   "Analog power off Uhost instnace",
		Long:    "Analog power off Uhost instnace",
		Example: "ucloud uhost poweroff --uhost-id uhost-xxx1,uhost-xxx2",
		Run: func(cmd *cobra.Command, args []string) {
			if !*yes {
				confirmText := "Danger, it may affect data integrity. Are you sure you want to poweroff this uhost?"
				if len(*uhostIDs) > 1 {
					confirmText = "Danger, it may affect data integrity. Are you sure you want to poweroff those uhosts?"
				}
				sure, err := ux.Prompt(confirmText)
				if err != nil {
					fmt.Fprintln(out, err)
					return
				}
				if !sure {
					return
				}
			}
			for _, id := range *uhostIDs {
				id = base.PickResourceID(id)
				req.UHostId = &id
				resp, err := base.BizClient.PoweroffUHostInstance(req)
				if err != nil {
					base.HandleError(err)
				} else {
					fmt.Fprintf(out, "uhost[%v] is power off\n", resp.UhostId)
				}
			}
		},
	}
	cmd.Flags().SortFlags = false
	uhostIDs = cmd.Flags().StringSlice("uhost-id", nil, "ResourceIDs(UHostIds) of the uhost instance")
	req.ProjectId = cmd.Flags().String("project-id", base.ConfigIns.ProjectID, "Assign project-id")
	req.Region = cmd.Flags().String("region", base.ConfigIns.Region, "Assign region")
	req.Zone = cmd.Flags().String("zone", "", "Assign availability zone")
	yes = cmd.Flags().BoolP("yes", "y", false, "Optional. Do not prompt for confirmation.")

	cmd.Flags().SetFlagValuesFunc("uhost-id", func() []string {
		return getUhostList([]string{status.HOST_FAIL, status.HOST_RUNNING, status.HOST_STOPPED}, *req.ProjectId, *req.Region, *req.Zone)
	})
	cmd.MarkFlagRequired("uhost-id")

	return cmd
}

func resizeUhost(req *uhost.ResizeUHostInstanceRequest) {

}

//NewCmdUHostResize ucloud uhost resize
func NewCmdUHostResize(out io.Writer) *cobra.Command {
	var yes, async *bool
	var bootDiskSize, dataDiskSize int
	var dataDiskID string
	var uhostIDs *[]string
	req := base.BizClient.NewResizeUHostInstanceRequest()
	cmd := &cobra.Command{
		Use:     "resize",
		Short:   "Resize uhost instance,such as cpu core count, memory size and disk size",
		Long:    "Resize uhost instance,such as cpu core count, memory size and disk size",
		Example: "ucloud uhost resize --uhost-id uhost-xxx1,uhost-xxx2 --cpu 4 --memory-gb 8",
		Run: func(cmd *cobra.Command, args []string) {
			if *req.CPU == 0 {
				req.CPU = nil
			}
			if *req.Memory == 0 {
				req.Memory = nil
			} else {
				*req.Memory *= 1024
			}
			for _, id := range *uhostIDs {
				id = base.PickResourceID(id)
				req.UHostId = &id
				host, err := describeUHostByID(id, *req.ProjectId, *req.Region, *req.Zone)
				if err != nil {
					base.Cxt.Println(err)
					return
				}
				inst := host.(*uhost.UHostInstanceSet)
				if inst.State == "Running" {
					if !*yes {
						confirmText := "Resize uhost must be done after the uhost is stopped. Do you want to stop this uhost?"
						if len(*uhostIDs) > 1 {
							confirmText = "Resize uhost must be done after the uhost is stopped. Do you want to stop those uhosts?"
						}
						agreeClose, err := ux.Prompt(confirmText)
						if err != nil {
							base.Cxt.Println(err)
							return
						}
						if !agreeClose {
							return
						}
					}
					_req := base.BizClient.NewStopUHostInstanceRequest()
					_req.ProjectId = req.ProjectId
					_req.Region = req.Region
					_req.Zone = req.Zone
					_req.UHostId = &id
					stopUhostIns(_req, false, out)
				}
				if req.CPU != nil || req.Memory != nil || *req.NetCapValue != 0 {
					resp, err := base.BizClient.ResizeUHostInstance(req)
					if err != nil {
						base.HandleError(err)
					} else {
						text := fmt.Sprintf("uhost [%v] cpu, memory resized", resp.UhostId)
						if *async {
							fmt.Fprintln(out, text)
						} else {
							poller := base.NewPoller(describeUHostByID, out)
							poller.Poll(resp.UhostId, *req.ProjectId, *req.Region, *req.Zone, text, []string{status.HOST_RUNNING, status.HOST_STOPPED, status.HOST_FAIL})
						}
					}
				}

				if dataDiskSize != 0 || bootDiskSize != 0 {
					_req := base.BizClient.NewResizeAttachedDiskRequest()
					var bootDisk uhost.UHostDiskSet
					var dataDisks = map[string]uhost.UHostDiskSet{}
					for _, disk := range inst.DiskSet {
						if disk.IsBoot == "True" {
							bootDisk = disk
						} else if disk.IsBoot == "False" {
							dataDisks[disk.DiskId] = disk
						}
					}
					if bootDiskSize != 0 {
						if bootDiskSize <= bootDisk.Size {
							base.LogError(fmt.Sprintf("Error, disk does not support shrinkage. current system-disk-size %dg", bootDisk.Size))
							continue
						} else {
							_req.DiskSpace = &bootDiskSize
							_req.DiskId = &bootDisk.DiskId
						}
					} else if dataDiskSize != 0 {
						var dataDisk uhost.UHostDiskSet
						if len(dataDisks) > 1 {
							if dataDiskID == "" {
								base.LogError(fmt.Sprintf("Error, the uhost %s have %d data disks. data-disk-id should be assigned", id, len(dataDisks)))
								continue
							}
							var ok bool
							dataDisk, ok = dataDisks[dataDiskID]
							if !ok {
								base.LogError(fmt.Sprintf("Error, the disk %s does not exist", dataDiskID))
								continue
							}
						} else if len(dataDisks) == 1 {
							for _, disk := range dataDisks {
								dataDisk = disk
							}
						} else if len(dataDisks) == 0 {
							base.LogError(fmt.Sprintf("Error, the uhost %s have no data disk. data-disk-id should be assigned", id))
							continue
						}
						if dataDiskSize <= dataDisk.Size {
							base.LogError(fmt.Sprintf("Error, disk does not support shrinkage. current data-disk-size %dg", dataDisk.Size))
							continue
						}
						_req.DiskSpace = &dataDiskSize
						_req.DiskId = &dataDisk.DiskId
					}
					_req.ProjectId = req.ProjectId
					_req.Region = req.Region
					_req.Zone = req.Zone
					_req.UHostId = &id
					_, err := base.BizClient.ResizeAttachedDisk(_req)
					if err != nil {
						base.HandleError(err)
					} else {
						text := fmt.Sprintf("uhost [%v] disk resized", id)
						if *async {
							fmt.Fprintln(out, text)
						} else {
							poller := base.NewPoller(describeUHostByID, out)
							poller.Poll(id, *req.ProjectId, *req.Region, *req.Zone, text, []string{status.HOST_RUNNING, status.HOST_STOPPED, status.HOST_FAIL})
						}
					}
				}
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	uhostIDs = cmd.Flags().StringSlice("uhost-id", nil, "Required. ResourceIDs(or UhostIDs) of the uhost instances")
	bindProjectID(req, flags)
	bindRegion(req, flags)
	bindZone(req, flags)
	req.CPU = cmd.Flags().Int("cpu", 0, "Optional. The number of virtual CPU cores. Series1 {1, 2, 4, 8, 12, 16, 24, 32}. Series2 {1,2,4,8,16}")
	req.Memory = cmd.Flags().Int("memory-gb", 0, "Optional. memory size. Unit: GB. Range: [1, 128], multiple of 2")
	cmd.Flags().IntVar(&bootDiskSize, "system-disk-size-gb", 0, "Optional. System disk size, unit GB. Range[20,100]. Step 10. System disk does not support shrinkage")
	cmd.Flags().IntVar(&dataDiskSize, "data-disk-size-gb", 0, "Optional. Data disk size,unit GB. Step 10. disk does not support shrinkage")
	cmd.Flags().StringVar(&dataDiskID, "data-disk-id", "", "Optional. If the uhost specified has two or more data disks, this parameter should be assigned")
	req.NetCapValue = cmd.Flags().Int("net-cap", 0, "Optional. NIC scale. 1,upgrade; 2,downgrade; 0,unchanged")
	yes = cmd.Flags().BoolP("yes", "y", false, "Optional. Do not prompt for confirmation.")
	async = cmd.Flags().BoolP("async", "a", false, "Optional. Do not wait for the long-running operation to finish.")
	cmd.Flags().SetFlagValuesFunc("uhost-id", func() []string {
		return getUhostList([]string{status.HOST_RUNNING, status.HOST_STOPPED, status.HOST_FAIL}, *req.ProjectId, *req.Region, *req.Zone)
	})
	cmd.MarkFlagRequired("uhost-id")
	return cmd
}

func describeUHostByID(uhostID, projectID, region, zone string) (interface{}, error) {
	req := base.BizClient.NewDescribeUHostInstanceRequest()
	req.UHostIds = []string{uhostID}
	req.ProjectId = &projectID
	req.Region = &region
	req.Zone = &zone

	resp, err := base.BizClient.DescribeUHostInstance(req)
	if err != nil {
		return nil, err
	}
	if len(resp.UHostSet) < 1 {
		return nil, nil
	}

	return &resp.UHostSet[0], nil
}

func sdescribeUHostByID(uhostID string) (interface{}, error) {
	req := base.BizClient.NewDescribeUHostInstanceRequest()
	req.UHostIds = []string{uhostID}

	resp, err := base.BizClient.DescribeUHostInstance(req)
	if err != nil {
		return nil, err
	}
	if len(resp.UHostSet) < 1 {
		return nil, nil
	}

	return &resp.UHostSet[0], nil
}

func getUhostList(states []string, project, region, zone string) []string {
	req := base.BizClient.NewDescribeUHostInstanceRequest()
	req.ProjectId = sdk.String(project)
	req.Region = sdk.String(region)
	req.Zone = sdk.String(zone)
	req.Limit = sdk.Int(50)
	resp, err := base.BizClient.DescribeUHostInstance(req)
	if err != nil {
		//todo runtime log
		return nil
	}
	list := []string{}
	for _, host := range resp.UHostSet {
		if states != nil {
			for _, s := range states {
				if host.State == s {
					list = append(list, host.UHostId+"/"+strings.Replace(host.Name, " ", "-", -1))
				}
			}
		} else {
			list = append(list, host.UHostId+"/"+strings.Replace(host.Name, " ", "-", -1))
		}
	}
	return list
}

//NewCmdUHostClone ucloud uhost clone
func NewCmdUHostClone(out io.Writer) *cobra.Command {
	var uhostID *string
	var async *bool
	req := base.BizClient.NewCreateUHostInstanceRequest()
	cmd := &cobra.Command{
		Use:   "clone",
		Short: "Create an uhost with the same configuration as another uhost, excluding bound eip and udisk",
		Long:  "Create an uhost with the same configuration as another uhost, excluding bound eip and udisk",
		Run: func(com *cobra.Command, args []string) {
			*uhostID = base.PickResourceID(*uhostID)
			queryReq := base.BizClient.NewDescribeUHostInstanceRequest()
			queryReq.ProjectId = req.ProjectId
			queryReq.Region = req.Region
			queryReq.Zone = req.Zone
			queryReq.UHostIds = []string{*uhostID}
			queryResp, err := base.BizClient.DescribeUHostInstance(queryReq)
			if err != nil {
				base.HandleError(err)
				return
			}
			if len(queryResp.UHostSet) < 1 {
				base.Cxt.PrintErr(fmt.Errorf("uhost[%s] not exist", *uhostID))
				return
			}
			queryFirewallReq := base.BizClient.NewDescribeFirewallRequest()
			queryFirewallReq.ProjectId = req.ProjectId
			queryFirewallReq.Region = req.Region
			queryFirewallReq.ResourceId = uhostID
			queryFirewallReq.ResourceType = sdk.String("uhost")

			firewallResp, err := base.BizClient.DescribeFirewall(queryFirewallReq)
			if err != nil {
				base.HandleError(err)
				return
			}

			if len(firewallResp.DataSet) == 1 {
				req.SecurityGroupId = &firewallResp.DataSet[0].FWId
			}

			uhostIns := queryResp.UHostSet[0]

			req.ImageId = &uhostIns.BasicImageId
			req.CPU = &uhostIns.CPU
			req.Memory = &uhostIns.Memory
			for _, ip := range uhostIns.IPSet {
				if ip.Type == "Private" {
					req.VPCId = &ip.VPCId
					req.SubnetId = &ip.SubnetId
				}
			}
			req.ChargeType = &uhostIns.ChargeType
			req.UHostType = &uhostIns.UHostType
			req.NetCapability = &uhostIns.NetCapability

			for _, disk := range uhostIns.DiskSet {
				item := uhost.UHostDisk{
					Size:   sdk.Int(disk.Size),
					Type:   sdk.String(disk.DiskType),
					IsBoot: sdk.String(disk.IsBoot),
				}
				if disk.BackupType != "" {
					item.BackupType = sdk.String(disk.BackupType)
				}
				req.Disks = append(req.Disks, item)
			}
			req.Tag = &uhostIns.Tag
			req.LoginMode = sdk.String("Password")
			resp, err := base.BizClient.CreateUHostInstance(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			if len(resp.UHostIds) == 1 {
				text := fmt.Sprintf("cloned uhost:[%s] is initializing", resp.UHostIds[0])
				if *async {
					fmt.Fprintln(out, text)
				} else {
					poller := base.NewPoller(describeUHostByID, out)
					poller.Poll(resp.UHostIds[0], *req.ProjectId, *req.Region, *req.Zone, text, []string{status.HOST_RUNNING, status.HOST_FAIL})
				}
			} else {
				base.HandleError(fmt.Errorf("expect uhost count 1, accept %d", len(resp.UHostIds)))
				return
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	uhostID = flags.String("uhost-id", "", "Required. Resource ID of the uhost to clone from")
	req.Password = flags.String("password", "", "Required. Password of the uhost user(root/ubuntu)")
	req.Name = flags.String("name", "", "Optional. Name of the uhost to clone")
	req.ProjectId = flags.String("project-id", base.ConfigIns.ProjectID, "Optional. Assign project-id")
	req.Region = flags.String("region", base.ConfigIns.Region, "Optional. Assign region")
	req.Zone = flags.String("zone", base.ConfigIns.Zone, "Optional. Assign availability zone")
	async = flags.Bool("async", false, "Optional. Do not wait for the long-running operation to finish.")
	flags.SetFlagValuesFunc("uhost-id", func() []string {
		return getUhostList([]string{status.HOST_RUNNING, status.HOST_STOPPED}, *req.ProjectId, *req.Region, *req.Zone)
	})
	cmd.MarkFlagRequired("uhost-id")
	cmd.MarkFlagRequired("password")
	return cmd
}

//NewCmdUhostCreateImage ucloud uhost create-image
func NewCmdUhostCreateImage(out io.Writer) *cobra.Command {
	var async *bool
	req := base.BizClient.NewCreateCustomImageRequest()
	cmd := &cobra.Command{
		Use:   "create-image",
		Short: "Create image from an uhost instance",
		Long:  "Create image from an uhost instance",
		Run: func(cmd *cobra.Command, args []string) {
			req.UHostId = sdk.String(base.PickResourceID(*req.UHostId))
			resp, err := base.BizClient.CreateCustomImage(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			text := fmt.Sprintf("iamge[%s] is making", resp.ImageId)
			if *async {
				fmt.Fprintln(out, text)
			} else {
				poller := base.NewPoller(describeImageByID, out)
				poller.Poll(resp.ImageId, *req.ProjectId, *req.Region, *req.Zone, text, []string{status.IMAGE_AVAILABLE, status.IMAGE_UNAVAILABLE})
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	req.UHostId = flags.String("uhost-id", "", "Resource ID of uhost to create image from")
	req.ImageName = flags.String("image-name", "", "Required. Name of the image to create")
	req.ImageDescription = flags.String("image-desc", "", "Optional. Description of the image to create")
	req.ProjectId = flags.String("project-id", base.ConfigIns.ProjectID, "Optional. Assign project-id")
	req.Region = flags.String("region", base.ConfigIns.Region, "Optional. Assign region")
	req.Zone = flags.String("zone", base.ConfigIns.Zone, "Optional. Assign availability zone")
	async = flags.BoolP("async", "a", false, "Optional. Do not wait for the long-running operation to finish.")

	flags.SetFlagValuesFunc("uhost-id", func() []string {
		return getUhostList([]string{status.HOST_RUNNING, status.HOST_STOPPED}, *req.ProjectId, *req.Region, *req.Zone)
	})

	cmd.MarkFlagRequired("uhost-id")
	cmd.MarkFlagRequired("image-name")
	return cmd
}

//NewCmdUhostResetPassword ucloud uhost reset-password
func NewCmdUhostResetPassword(out io.Writer) *cobra.Command {
	var yes *bool
	var uhostIDs *[]string
	req := base.BizClient.NewResetUHostInstancePasswordRequest()
	cmd := &cobra.Command{
		Use:   "reset-password",
		Short: "Reset the administrator password for the UHost instances.",
		Long:  "Reset the administrator password for the UHost instances.",
		Run: func(cmd *cobra.Command, args []string) {
			for _, id := range *uhostIDs {
				id = base.PickResourceID(id)
				req.UHostId = &id
				err := checkAndCloseUhost(*yes, false, id, *req.ProjectId, *req.Region, *req.Zone, out)
				if err != nil {
					base.Cxt.Println(err)
					continue
				}
				host, err := describeUHostByID(id, *req.ProjectId, *req.Region, *req.Zone)
				inst, ok := host.(*uhost.UHostInstanceSet)
				if !ok {
					return
				}
				if inst.BootDiskState == "Initializing" {
					fmt.Fprintf(out, "uhost[%s] boot disk in initializing, wait 10 minutes\n", id)
					return
				}
				resp, err := base.BizClient.ResetUHostInstancePassword(req)
				if err != nil {
					base.HandleError(err)
					return
				}
				fmt.Fprintf(out, "uhost[%s] reset password\n", resp.UhostId)
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	uhostIDs = flags.StringSlice("uhost-id", nil, "Required. Resource IDs of the uhosts to reset the administrator's password")
	req.Password = flags.String("password", "", "Required. New Password")
	req.ProjectId = flags.String("project-id", base.ConfigIns.ProjectID, "Optional. Assign project-id")
	req.Region = flags.String("region", base.ConfigIns.Region, "Optional. Assign region")
	req.Zone = flags.String("zone", base.ConfigIns.Zone, "Optional. Assign availability zone")
	yes = cmd.Flags().BoolP("yes", "y", false, "Optional. Do not prompt for confirmation.")
	flags.SetFlagValuesFunc("uhost-id", func() []string {
		return getUhostList([]string{status.HOST_RUNNING, status.HOST_STOPPED}, *req.ProjectId, *req.Region, *req.Zone)
	})
	cmd.MarkFlagRequired("uhost-id")
	cmd.MarkFlagRequired("password")
	return cmd
}

func checkAndCloseUhost(yes, async bool, uhostID, project, region, zone string, out io.Writer) error {
	host, err := describeUHostByID(uhostID, project, region, zone)
	if err != nil {
		return err
	}
	inst, ok := host.(*uhost.UHostInstanceSet)
	if ok {
		if inst.State == "Running" {
			if !yes {
				confirmText := fmt.Sprintf("uhost[%s] will be stopped, can we do this?", uhostID)
				agreeClose, err := ux.Prompt(confirmText)
				if err != nil {
					return err
				}
				if !agreeClose {
					return fmt.Errorf("skip, you do not agree to stop uhost")
				}
			}
			_req := base.BizClient.NewStopUHostInstanceRequest()
			_req.ProjectId = &project
			_req.Region = &region
			_req.Zone = &zone
			_req.UHostId = &uhostID
			stopUhostIns(_req, async, out)
		}
	} else {
		return fmt.Errorf("Something wrong, uhost[%s] may not exist", uhostID)
	}
	return nil
}

//NewCmdUhostReinstallOS ucloud uhost reinstall-os
func NewCmdUhostReinstallOS(out io.Writer) *cobra.Command {
	var isReserveDataDisk, yes, async *bool
	req := base.BizClient.NewReinstallUHostInstanceRequest()
	cmd := &cobra.Command{
		Use:   "reinstall-os",
		Short: "Reinstall the operating system of the UHost instance",
		Long:  "Reinstall the operating system of the UHost instance. we will detach all udisk disks if the uhost attached some, and then stop the uhost if it's running",
		Run: func(cmd *cobra.Command, args []string) {
			if *isReserveDataDisk {
				req.ReserveDisk = sdk.String("Yes")
			} else {
				req.ReserveDisk = sdk.String("No")
			}
			req.UHostId = sdk.String(base.PickResourceID(*req.UHostId))
			req.Password = sdk.String(base64.StdEncoding.EncodeToString([]byte(sdk.StringValue(req.Password))))

			any, err := describeUHostByID(*req.UHostId, *req.ProjectId, *req.Region, *req.Zone)
			if err != nil {
				base.Cxt.Println(err)
				return
			}
			uhostIns, ok := any.(*uhost.UHostInstanceSet)
			if ok {
				for _, disk := range uhostIns.DiskSet {
					if disk.Type == "Udisk" {
						sure := false
						if !*yes {
							text := fmt.Sprintf("udisk[%s/%s] will be detached, can we do this?", disk.DiskId, disk.Name)
							sure, err = ux.Prompt(text)
							if err != nil {
								base.Cxt.PrintErr(err)
								return
							}
							if !sure {
								base.Cxt.Printf("you don't agree to detach udisk\n")
								return
							}
						}
						if *yes || sure {
							err := detachUdisk(false, disk.DiskId, out)
							if err != nil {
								base.Cxt.Println(err)
								return
							}
						}
					}
				}
			} else {
				base.Cxt.Printf("Something wrong, uhost[%s] may not exist\n", *req.UHostId)
				return
			}

			err = checkAndCloseUhost(*yes, *async, *req.UHostId, *req.ProjectId, *req.Region, *req.Zone, out)
			if err != nil {
				base.Cxt.Println(err)
				return
			}
			resp, err := base.BizClient.ReinstallUHostInstance(req)
			if err != nil {
				base.Cxt.Println(err)
				return
			}
			text := fmt.Sprintf("uhost[%s] is reinstalling OS", *req.UHostId)
			if *async {
				fmt.Fprintln(out, text)
			} else {
				poller := base.NewPoller(describeUHostByID, out)
				poller.Poll(resp.UhostId, *req.ProjectId, *req.Region, *req.Zone, text, []string{status.HOST_RUNNING, status.HOST_FAIL})
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.UHostId = flags.String("uhost-id", "", "Required. Resource ID of the uhost to reinstall operating system")
	req.Password = flags.String("password", "", "Required. Password of the administrator")
	req.ImageId = flags.String("image-id", "", "Optional. Resource ID the image to install. See 'ucloud image list'. Default is original image of the uhost")
	req.ProjectId = flags.String("project-id", base.ConfigIns.ProjectID, "Optional. Assign project-id")
	req.Region = flags.String("region", base.ConfigIns.Region, "Optional. Assign region")
	req.Zone = flags.String("zone", base.ConfigIns.Zone, "Optional. Assign availability zone")
	isReserveDataDisk = flags.Bool("keep-data-disk", false, "Keep data disk or not. If you keep data disk, you can't change OS type(Linux->Window,e.g.)")
	yes = cmd.Flags().BoolP("yes", "y", false, "Optional. Do not prompt for confirmation.")
	async = flags.BoolP("async", "a", false, "Optional. Do not wait for the long-running operation to finish.")
	flags.SetFlagValuesFunc("uhost-id", func() []string {
		return getUhostList([]string{status.HOST_RUNNING, status.HOST_STOPPED}, *req.ProjectId, *req.Region, *req.Zone)
	})
	cmd.MarkFlagRequired("uhost-id")
	cmd.MarkFlagRequired("password")
	return cmd
}

//NewCmdUhostLeaveIsolationGroup ucloud uhost leave-isolation-group
func NewCmdUhostLeaveIsolationGroup(out io.Writer) *cobra.Command {
	var uhostIds []string
	req := base.BizClient.NewLeaveIsolationGroupRequest()
	cmd := &cobra.Command{
		Use:   "leave-isolation-group",
		Short: "Detach uhost from its isolation group",
		Run: func(c *cobra.Command, args []string) {
			for _, idname := range uhostIds {
				id := base.PickResourceID(idname)
				any, err := describeUHostByID(id, *req.ProjectId, *req.Region, *req.Zone)
				if err != nil {
					base.LogError(fmt.Sprintf("fetch uhost %s failed: %v", idname, err))
					continue
				}
				ins, ok := any.(*uhost.UHostInstanceSet)
				if !ok {
					base.LogError(fmt.Sprintf("uhost %s may not exist", idname))
					continue
				}
				if ins.IsolationGroup == "" {
					base.LogPrint(fmt.Sprintf("uhost %s doesn't attached any isolation group", idname))
					continue
				}
				req.GroupId = sdk.String(ins.IsolationGroup)
				req.UHostId = &id
				_, err = base.BizClient.LeaveIsolationGroup(req)
				if err != nil {
					base.HandleError(err)
					continue
				}
				base.LogPrint(fmt.Sprintf("uhost %s detached from isolation group %s", idname, ins.IsolationGroup))
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	flags.StringSliceVar(&uhostIds, "uhost-id", nil, "Required. Resource ID of uhosts to be detech from its isolation group")
	bindRegion(req, flags)
	bindProjectID(req, flags)
	bindZone(req, flags)
	cmd.MarkFlagRequired("uhost-id")
	flags.SetFlagValuesFunc("uhost-id", func() []string {
		return getUhostList(nil, *req.ProjectId, *req.Region, *req.Zone)
	})
	return cmd
}

//NewCmdIsolation ucloud uhost isolation-gorup
func NewCmdIsolation(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "isolation-group",
		Short: "List and manipulate isolation group of uhost",
		Long:  "List and manipulate isolation group of uhost",
	}
	cmd.AddCommand(NewCmdIsolationList(out))
	cmd.AddCommand(NewCmdIsolationCreate(out))
	cmd.AddCommand(NewCmdIsolationDelete(out))
	return cmd
}

//NewCmdIsolationCreate ucloud uhost isolation-group create
func NewCmdIsolationCreate(out io.Writer) *cobra.Command {
	req := base.BizClient.NewCreateIsolationGroupRequest()
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create isolation group instance",
		Long:  "Create isolation group instance",
		Run: func(c *cobra.Command, args []string) {
			re := regexp.MustCompile(cli.REGEXP_NAME)
			if !re.Match([]byte(*req.GroupName)) {
				base.LogError(fmt.Sprintf("group-name %s is invalid! Length 1~63, only English,Chinese,number and '-_.' are allowed", *req.GroupName))
				return
			}
			resp, err := base.BizClient.CreateIsolationGroup(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			base.LogPrint(fmt.Sprintf("isolation group %s created", resp.GroupId))
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	req.GroupName = flags.String("group-name", "", "Required. Name of isolation group. Length 1~63, only English,Chinese,number and '-_.' are allowed")
	bindRegion(req, flags)
	bindProjectID(req, flags)
	req.Remark = flags.String("remark", "", "Optional. Remark ok isolation group")

	cmd.MarkFlagRequired("group-name")
	return cmd
}

//NewCmdIsolationDelete ucloud uhost
func NewCmdIsolationDelete(out io.Writer) *cobra.Command {
	var ids []string
	req := base.BizClient.NewDeleteIsolationGroupRequest()
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete isolation group instances",
		Run: func(c *cobra.Command, args []string) {
			for _, idname := range ids {
				id := base.PickResourceID(idname)
				req.GroupId = &id
				_, err := base.BizClient.DeleteIsolationGroup(req)
				if err != nil {
					base.HandleError(err)
					continue
				}
				base.LogPrint(fmt.Sprintf("isolation group %s deleted", idname))
			}
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	flags.StringSliceVar(&ids, "group-id", nil, "Required. Resource ID of isolation groups to be deleted")
	bindRegion(req, flags)
	bindProjectID(req, flags)

	cmd.MarkFlagRequired("group-id")
	flags.SetFlagValuesFunc("group-id", func() []string {
		return getIsolationGroupList(*req.ProjectId, *req.Region)
	})

	return cmd
}

type isolationGroupRow struct {
	ResourceID string
	Name       string
	Remark     string
	UHostCount string
}

//NewCmdIsolationList ucloud uhost isolation-group list
func NewCmdIsolationList(out io.Writer) *cobra.Command {
	req := base.BizClient.NewDescribeIsolationGroupRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List isolation group of uhost",
		Run: func(c *cobra.Command, args []string) {
			resp, err := base.BizClient.DescribeIsolationGroup(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			var list []isolationGroupRow
			for _, group := range resp.IsolationGroupSet {
				row := isolationGroupRow{
					ResourceID: group.GroupId,
					Name:       group.GroupName,
					Remark:     group.Remark,
				}
				var zones []string
				for _, item := range group.SpreadInfoSet {
					zones = append(zones, fmt.Sprintf("%s:%d", item.Zone, item.UHostCount))
				}
				row.UHostCount = strings.Join(zones, " ")
				list = append(list, row)
			}
			base.PrintList(list, out)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	req.GroupId = flags.String("group-id", "", "Optional. Resource ID of isolation group to describe")
	bindRegion(req, flags)
	bindProjectID(req, flags)
	bindLimit(req, flags)
	bindOffset(req, flags)

	flags.SetFlagValuesFunc("group-id", func() []string {
		return getIsolationGroupList(*req.ProjectId, *req.Region)
	})

	return cmd
}

func getIsolationGroupList(project, region string) []string {
	req := base.BizClient.NewDescribeIsolationGroupRequest()
	req.ProjectId = sdk.String(project)
	req.Region = sdk.String(region)
	req.Limit = sdk.Int(50)
	resp, err := base.BizClient.DescribeIsolationGroup(req)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	list := []string{}
	for _, group := range resp.IsolationGroupSet {
		list = append(list, group.GroupId+"/"+strings.Replace(group.GroupName, " ", "-", -1))
	}
	return list
}
