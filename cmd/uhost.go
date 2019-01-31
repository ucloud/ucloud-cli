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
	"strings"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/uhost"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/base"
	"github.com/ucloud/ucloud-cli/model/cli"
	"github.com/ucloud/ucloud-cli/model/status"
	"github.com/ucloud/ucloud-cli/ux"
)

//NewCmdUHost ucloud uhost
func NewCmdUHost() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uhost",
		Short: "List,create,delete,stop,restart,poweroff or resize UHost instance",
		Long:  `List,create,delete,stop,restart,poweroff or resize UHost instance`,
		Args:  cobra.NoArgs,
	}
	writer := base.Cxt.GetWriter()
	cmd.AddCommand(NewCmdUHostList())
	cmd.AddCommand(NewCmdUHostCreate(writer))
	cmd.AddCommand(NewCmdUHostDelete(writer))
	cmd.AddCommand(NewCmdUHostStop(writer))
	cmd.AddCommand(NewCmdUHostStart(writer))
	cmd.AddCommand(NewCmdUHostReboot(writer))
	cmd.AddCommand(NewCmdUHostPoweroff())
	cmd.AddCommand(NewCmdUHostResize(writer))
	cmd.AddCommand(NewCmdUHostClone(writer))
	cmd.AddCommand(NewCmdUhostResetPassword(writer))
	cmd.AddCommand(NewCmdUhostReinstallOS(writer))
	cmd.AddCommand(NewCmdUhostCreateImage(writer))

	return cmd
}

//UHostRow UHost表格行
type UHostRow struct {
	UHostName    string
	ResourceID   string
	Group        string
	PrivateIP    string
	PublicIP     string
	Config       string
	Type         string
	State        string
	CreationTime string
}

//NewCmdUHostList [ucloud uhost list]
func NewCmdUHostList() *cobra.Command {
	req := base.BizClient.NewDescribeUHostInstanceRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all UHost Instances",
		Long:  `List all UHost Instances`,
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := base.BizClient.DescribeUHostInstance(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			list := make([]UHostRow, 0)
			for _, host := range resp.UHostSet {
				row := UHostRow{}
				row.UHostName = host.Name
				row.ResourceID = host.UHostId
				row.Group = host.Tag
				for _, ip := range host.IPSet {
					if row.PublicIP != "" {
						row.PublicIP += " | "
					}
					if ip.Type == "Private" {
						row.PrivateIP = ip.IP
					} else {
						row.PublicIP += fmt.Sprintf("%s %s", ip.IP, ip.Type)
					}
				}
				osName := strings.SplitN(host.OsName, " ", 2)
				cupCore := host.CPU
				memorySize := host.Memory / 1024
				diskSize := 0
				for _, disk := range host.DiskSet {
					if disk.Type == "Data" || disk.Type == "Udisk" {
						diskSize += disk.Size
					}
				}
				row.Config = fmt.Sprintf("%s cpu:%d memory:%dG disk:%dG", osName[0], cupCore, memorySize, diskSize)
				row.CreationTime = base.FormatDate(host.CreateTime)
				row.State = host.State
				row.Type = host.UHostType + "/" + host.HostType
				list = append(list, row)
			}
			base.PrintList(list, global.json)
		},
	}
	cmd.Flags().SortFlags = false
	req.ProjectId = cmd.Flags().String("project-id", base.ConfigIns.ProjectID, "Optional. Assign project-id")
	req.Region = cmd.Flags().String("region", base.ConfigIns.Region, "Optional. Assign region")
	req.Zone = cmd.Flags().String("zone", "", "Optional. Assign availability zone")
	cmd.Flags().StringSliceVar(&req.UHostIds, "uhost-id", make([]string, 0), "Optional. UHost Instance ID, multiple values separated by comma(without space)")
	// req.Tag = cmd.Flags().String("group", "", "Optional. Business group")
	req.Offset = cmd.Flags().Int("offset", 0, "Optional. Offset default 0")
	req.Limit = cmd.Flags().Int("limit", 50, "Optional. Limit default 50, max value 100")
	bindGroup(req, cmd.Flags())

	return cmd
}

//NewCmdUHostCreate [ucloud uhost create]
func NewCmdUHostCreate(out io.Writer) *cobra.Command {
	var bindEipID *string
	var async *bool

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

			resp, err := base.BizClient.CreateUHostInstance(req)
			if err != nil {
				base.HandleError(err)
				return
			}

			if len(resp.UHostIds) == 1 {
				text := fmt.Sprintf("uhost[%s] is initializing", resp.UHostIds[0])
				if *async {
					fmt.Fprintln(out, text)
				} else {
					poller := base.NewPoller(describeUHostByID, out)
					poller.Poll(resp.UHostIds[0], *req.ProjectId, *req.Region, *req.Zone, text, []string{status.HOST_RUNNING, status.HOST_FAIL})
				}
			} else {
				fmt.Fprintf(out, "expect uhost count 1 , accept %d", len(resp.UHostIds))
				return
			}
			bindEipID = sdk.String(base.PickResourceID(*bindEipID))
			if *bindEipID != "" && len(resp.UHostIds) == 1 {
				bindEIP(sdk.String(resp.UHostIds[0]), sdk.String("uhost"), bindEipID, req.ProjectId, req.Region)
			} else if *eipReq.OperatorName != "" && *eipReq.Bandwidth != 0 {
				eipReq.ChargeType = req.ChargeType
				eipReq.Tag = req.Tag
				eipReq.Quantity = req.Quantity
				eipReq.Region = req.Region
				eipReq.ProjectId = req.ProjectId
				eipResp, err := base.BizClient.AllocateEIP(eipReq)

				if err != nil {
					base.HandleError(err)
				} else {
					for _, eip := range eipResp.EIPSet {
						base.Cxt.Printf("allocate EIP[%s] ", eip.EIPId)
						for _, ip := range eip.EIPAddr {
							base.Cxt.Printf("IP:%s  Line:%s \n", ip.IP, ip.OperatorName)
						}
						if len(resp.UHostIds) == 1 {
							bindEIP(sdk.String(resp.UHostIds[0]), sdk.String("uhost"), sdk.String(eip.EIPId), req.ProjectId, req.Region)
						}
					}
				}
			}
		},
	}

	n1Zone := map[string]bool{
		"cn-bj2-01": true,
		"cn-bj2-03": true,
		"cn-sh2-01": true,
		"hk-01":     true,
	}
	defaultUhostType := "N2"
	if _, ok := n1Zone[base.ConfigIns.Zone]; ok {
		defaultUhostType = "N1"
	}

	req.Disks = make([]uhost.UHostDisk, 2)
	req.Disks[0].IsBoot = sdk.String("True")
	req.Disks[1].IsBoot = sdk.String("False")

	flags := cmd.Flags()
	flags.SortFlags = false
	async = flags.Bool("async", false, "Optional. Do not wait for the long-running operation to finish.")
	req.CPU = flags.Int("cpu", 4, "Required. The count of CPU cores. Optional parameters: {1, 2, 4, 8, 12, 16, 24, 32}")
	req.Memory = flags.Int("memory-gb", 8, "Required. Memory size. Unit: GB. Range: [1, 128], multiple of 2")
	req.Password = flags.String("password", "", "Required. Password of the uhost user(root/ubuntu)")
	req.ImageId = flags.String("image-id", "", "Required. The ID of image. see 'ucloud image list'")
	req.VPCId = flags.String("vpc-id", "", "Optional. VPC ID. This field is required under VPC2.0. See 'ucloud vpc list'")
	req.SubnetId = flags.String("subnet-id", "", "Optional. Subnet ID. This field is required under VPC2.0. See 'ucloud subnet list'")
	req.Name = flags.String("name", "UHost", "Optional. UHost instance name")
	bindEipID = flags.String("bind-eip", "", "Optional. Resource ID or IP Address of eip that will be bound to the new created uhost")
	eipReq.OperatorName = flags.String("create-eip-line", "", "Optional. Required if you want to create new EIP. Line of the created eip to be bound with the new created uhost")
	eipReq.Bandwidth = cmd.Flags().Int("create-eip-bandwidth-mb", 0, "Optional. Required if you want to create new EIP. Bandwidth(Unit:Mbps).The range of value related to network charge mode. By traffic [1, 300]; by bandwidth [1,800] (Unit: Mbps); it could be 0 if the eip belong to the shared bandwidth")
	eipReq.PayMode = cmd.Flags().String("create-eip-traffic-mode", "Bandwidth", "Optional. 'Traffic','Bandwidth' or 'ShareBandwidth'")
	eipReq.Name = flags.String("create-eip-name", "", "Optional. Name of created eip to bind with the uhost")
	eipReq.Remark = cmd.Flags().String("create-eip-remark", "", "Optional.Remark of your EIP.")

	req.ChargeType = flags.String("charge-type", "Month", "Optional.'Year',pay yearly;'Month',pay monthly;'Dynamic', pay hourly")
	req.Quantity = flags.Int("quantity", 1, "Optional. The duration of the instance. N years/months.")
	bindProjectID(req, flags)
	bindRegion(req, flags)
	// bindZone(req, flags)
	req.Zone = flags.String("zone", base.ConfigIns.Zone, "Optional. Override default available zone, see 'ucloud region'")

	req.UHostType = flags.String("type", defaultUhostType, "Optional. Default is 'N2' of which cpu is V4 and sata disk. also support 'N1' means V3 cpu and sata disk;'I2' means V4 cpu and ssd disk;'D1' means big data model;'G1' means GPU type, model for K80;'G2' model for P40; 'G3' model for V100")
	req.NetCapability = flags.String("net-capability", "Normal", "Optional. Default is 'Normal', also support 'Super' which will enhance multiple times network capability as before")
	req.Disks[0].Type = flags.String("os-disk-type", "LOCAL_NORMAL", "Optional. Enumeration value. 'LOCAL_NORMAL', Ordinary local disk; 'CLOUD_NORMAL', Ordinary cloud disk; 'LOCAL_SSD',local ssd disk; 'CLOUD_SSD',cloud ssd disk; 'EXCLUSIVE_LOCAL_DISK',big data. The disk only supports a limited combination.")
	req.Disks[0].Size = flags.Int("os-disk-size-gb", 20, "Optional. Default 20G. Windows should be bigger than 40G Unit GB")
	req.Disks[0].BackupType = flags.String("os-disk-backup-type", "NONE", "Optional. Enumeration value, 'NONE' or 'DATAARK'. DataArk supports real-time backup, which can restore the disk back to any moment within the last 12 hours. (Normal Local Disk and Normal Cloud Disk Only)")
	req.Disks[1].Type = flags.String("data-disk-type", "LOCAL_NORMAL", "Optional. Enumeration value. 'LOCAL_NORMAL', Ordinary local disk; 'CLOUD_NORMAL', Ordinary cloud disk; 'LOCAL_SSD',local ssd disk; 'CLOUD_SSD',cloud ssd disk; 'EXCLUSIVE_LOCAL_DISK',big data. The disk only supports a limited combination.")
	req.Disks[1].Size = flags.Int("data-disk-size-gb", 20, "Optional. Disk size. Unit GB")
	req.Disks[1].BackupType = flags.String("data-disk-backup-type", "NONE", "Optional. Enumeration value, 'NONE' or 'DATAARK'. DataArk supports real-time backup, which can restore the disk back to any moment within the last 12 hours. (Normal Local Disk and Normal Cloud Disk Only)")
	req.SecurityGroupId = flags.String("firewall-id", "", "Optional. Firewall Id, default: Web recommended firewall. see 'ucloud firewall list'.")
	req.Tag = flags.String("group", "Default", "Optional. Business group")

	flags.SetFlagValues("charge-type", "Month", "Year", "Dynamic", "Trial")
	flags.SetFlagValues("cpu", "1", "2", "4", "8", "12", "16", "24", "32")
	flags.SetFlagValues("type", "N2", "N1", "I2", "D1", "G1", "G2", "G3")
	flags.SetFlagValues("net-capability", "Normal", "Super")
	flags.SetFlagValues("os-disk-type", "LOCAL_NORMAL", "CLOUD_NORMAL", "LOCAL_SSD", "CLOUD_SSD", "EXCLUSIVE_LOCAL_DISK")
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

	cmd.MarkFlagRequired("cpu")
	cmd.MarkFlagRequired("memory")
	cmd.MarkFlagRequired("password")
	cmd.MarkFlagRequired("image-id")

	return cmd
}

//NewCmdUHostDelete ucloud uhost delete
func NewCmdUHostDelete(out io.Writer) *cobra.Command {
	var uhostIDs *[]string
	var isDestory = sdk.Bool(false)
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
			if *isDestory {
				req.Destroy = sdk.Int(1)
			} else {
				req.Destroy = sdk.Int(0)
			}
			for _, id := range *uhostIDs {
				id = base.PickResourceID(id)
				req.UHostId = &id
				hostIns, err := describeUHostByID(*req.UHostId, *req.ProjectId, *req.Region, *req.Zone)
				if err != nil {
					base.HandleError(err)
				} else if hostIns != nil {
					ins := hostIns.(*uhost.UHostInstanceSet)
					if ins.State == "Running" {
						_req := base.BizClient.NewStopUHostInstanceRequest()
						_req.ProjectId = req.ProjectId
						_req.Region = req.Region
						_req.Zone = req.Zone
						_req.UHostId = req.UHostId
						stopUhostIns(_req, false, out)
					}
				}
				resp, err := base.BizClient.TerminateUHostInstance(req)
				if err != nil {
					base.HandleError(err)
				} else {
					base.Cxt.Printf("uhost:[%s] deleted\n", resp.UHostId)
				}
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	uhostIDs = cmd.Flags().StringSlice("uhost-id", nil, "Requried. ResourceIDs(UhostIds) of the uhost instance")
	bindRegion(req, flags)
	bindProjectID(req, flags)
	req.Zone = cmd.Flags().String("zone", "", "Optional. availability zone")
	isDestory = cmd.Flags().Bool("destory", false, "Optional. false,the uhost instance will be thrown to UHost recycle if you have permission; true,the uhost instance will be deleted directly")
	req.ReleaseEIP = cmd.Flags().Bool("release-eip", false, "Optional. false,Unbind EIP only; true, Unbind EIP and release it")
	req.ReleaseUDisk = cmd.Flags().Bool("delete-cloud-disk", false, "Optional. false,Detach cloud disk only; true, Detach cloud disk and delete it")
	yes = cmd.Flags().BoolP("yes", "y", false, "Optional. Do not prompt for confirmation.")
	cmd.Flags().SetFlagValues("destory", "true", "false")
	cmd.Flags().SetFlagValues("release-eip", "true", "false")
	cmd.Flags().SetFlagValues("delete-cloud-disk", "true", "false")
	cmd.Flags().SetFlagValuesFunc("uhost-id", func() []string {
		return getUhostList([]string{status.HOST_RUNNING, status.HOST_FAIL, status.HOST_FAIL}, *req.ProjectId, *req.Region, *req.Zone)
	})
	cmd.MarkFlagRequired("uhost-id")

	return cmd
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
		text := fmt.Sprintf("uhost:[%v] is shutting down", resp.UhostId)
		if async {
			fmt.Fprintln(out, text)
		} else {
			poller := base.NewPoller(describeUHostByID, out)
			poller.Poll(resp.UhostId, *req.ProjectId, *req.Region, *req.Zone, text, []string{status.HOST_STOPPED, status.HOST_FAIL})
		}
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
					text := fmt.Sprintf("uhost:[%v] is starting", resp.UhostId)
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
					text := fmt.Sprintf("UHost:[%v] is restarting", resp.UhostId)
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
func NewCmdUHostPoweroff() *cobra.Command {
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
					base.Cxt.Println(err)
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
					base.Cxt.Printf("UHost:[%v] is power off\n", resp.UhostId)
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
	cmd.MarkFlagRequired("uhost-id")
	return cmd
}

//NewCmdUHostResize ucloud uhost resize
func NewCmdUHostResize(out io.Writer) *cobra.Command {
	var yes, async *bool
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
			if *req.DiskSpace == 0 {
				req.DiskSpace = nil
			}
			if *req.BootDiskSpace == 0 {
				req.BootDiskSpace = nil
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
						confirmText := "Resize uhost must be after stop it. Do you want to stop this uhost?"
						if len(*uhostIDs) > 1 {
							confirmText = "Resize uhost must be after stop it. Do you want to stop those uhosts?"
						}
						agreeClose, err := ux.Prompt(confirmText)
						if err != nil {
							base.Cxt.Println(err)
							return
						}
						if !agreeClose {
							continue
						}
					}
					_req := base.BizClient.NewStopUHostInstanceRequest()
					_req.ProjectId = req.ProjectId
					_req.Region = req.Region
					_req.Zone = req.Zone
					_req.UHostId = &id
					stopUhostIns(_req, false, out)
				}

				resp, err := base.BizClient.ResizeUHostInstance(req)
				if err != nil {
					base.HandleError(err)
				} else {
					text := fmt.Sprintf("UHost:[%v] resized", resp.UhostId)
					if *async {
						fmt.Fprintln(out, text)
					} else {
						poller := base.NewPoller(describeUHostByID, out)
						poller.Poll(resp.UhostId, *req.ProjectId, *req.Region, *req.Zone, text, []string{status.HOST_RUNNING, status.HOST_STOPPED, status.HOST_FAIL})
					}
				}
			}
		},
	}
	cmd.Flags().SortFlags = false
	uhostIDs = cmd.Flags().StringSlice("uhost-id", nil, "Required. ResourceIDs(or UhostIDs) of the uhost instances")
	req.ProjectId = cmd.Flags().String("project-id", base.ConfigIns.ProjectID, "Optional. Assign project-id")
	req.Region = cmd.Flags().String("region", base.ConfigIns.Region, "Optional. Assign region")
	req.Zone = cmd.Flags().String("zone", "", "Optional. Assign availability zone")
	req.CPU = cmd.Flags().Int("cpu", 0, "Optional. The number of virtual CPU cores. Series1 {1, 2, 4, 8, 12, 16, 24, 32}. Series2 {1,2,4,8,16}")
	req.Memory = cmd.Flags().Int("memory-gb", 0, "Optional. memory size. Unit: GB. Range: [1, 128], multiple of 2")
	req.DiskSpace = cmd.Flags().Int("data-disk-size-gb", 0, "Optional. Data disk size,unit GB. Range[10,1000], SSD disk range[100,500]. Step 10")
	req.BootDiskSpace = cmd.Flags().Int("system-disk-size-gb", 0, "Optional. System disk size, unit GB. Range[20,100]. Step 10. System disk does not support shrinkage")
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
		for _, s := range states {
			if host.State == s {
				list = append(list, host.UHostId+"/"+strings.Replace(host.Name, " ", "-", -1))
			}
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
		Short: "Create an uhost with the same configuration as another uhost",
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

			for index := 0; index < 2; index++ {
				disk := uhostIns.DiskSet[index]
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
