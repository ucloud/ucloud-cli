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
	"net"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/uhost"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/base"
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
	cmd.AddCommand(NewCmdUHostList())
	cmd.AddCommand(NewCmdUHostCreate())
	cmd.AddCommand(NewCmdUHostDelete())
	cmd.AddCommand(NewCmdUHostStop())
	cmd.AddCommand(NewCmdUHostStart())
	cmd.AddCommand(NewCmdUHostReboot())
	cmd.AddCommand(NewCmdUHostPoweroff())
	cmd.AddCommand(NewCmdUHostResize())

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
			if global.json {
				base.PrintJSON(resp.UHostSet)
			} else {
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
				base.PrintTableS(list)
			}
		},
	}
	cmd.Flags().SortFlags = false
	req.ProjectId = cmd.Flags().String("project-id", base.ConfigInstance.ProjectID, "Optional. Assign project-id")
	req.Region = cmd.Flags().String("region", base.ConfigInstance.Region, "Optional. Assign region")
	req.Zone = cmd.Flags().String("zone", "", "Optional. Assign availability zone")
	cmd.Flags().StringSliceVar(&req.UHostIds, "resource-id", make([]string, 0), "Optional. UHost Instance ID, multiple values separated by comma(without space)")
	req.Tag = cmd.Flags().String("group", "", "Optional. Group")
	req.Offset = cmd.Flags().Int("offset", 0, "Optional. Offset default 0")
	req.Limit = cmd.Flags().Int("limit", 50, "Optional. Limit default 50, max value 100")

	return cmd
}

//NewCmdUHostCreate [ucloud uhost create]
func NewCmdUHostCreate() *cobra.Command {
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
			images := strings.SplitN(*req.ImageId, "/", 2)
			if len(images) >= 2 {
				*req.ImageId = images[0]
			}

			resp, err := base.BizClient.CreateUHostInstance(req)
			if err != nil {
				base.HandleError(err)
				return
			}

			if !*async {
				if len(resp.UHostIds) == 1 {
					text := fmt.Sprintf("UHost:[%s] is initializing", resp.UHostIds[0])
					done := pollUhost(resp.UHostIds[0], *req.ProjectId, *req.Region, *req.Zone, []string{status.HOST_RUNNING, status.HOST_FAIL})
					ux.DotSpinner.Start(text)
					<-done
					ux.DotSpinner.Stop()
				}
			} else {
				base.Cxt.Printf("UHost:%v created\n", resp.UHostIds)
			}

			if *bindEipID != "" && len(resp.UHostIds) == 1 {
				ip := net.ParseIP(*bindEipID)
				if ip != nil {
					eipID, err := getEIPIDbyIP(ip, *req.ProjectId, *req.Region)
					if err != nil {
						base.HandleError(err)
					} else {
						*bindEipID = eipID
					}
				}
				bindEIP(sdk.String(resp.UHostIds[0]), sdk.String("uhost"), bindEipID, req.ProjectId, req.Region)
			}

			if *eipReq.OperatorName != "" && *eipReq.Bandwidth != 0 {
				if *eipReq.OperatorName == "BGP" {
					*eipReq.OperatorName = "Bgp"
				}
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
	if _, ok := n1Zone[base.ConfigInstance.Zone]; ok {
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
	bindEipID = flags.String("bind-eip", "", "Optional. Bind eip to uhost. Value could be resource id or IP Address")
	eipReq.OperatorName = flags.String("create-eip-line", "", "Optional. Required if you want to create new EIP. Line of created eip to bind with the uhost")
	eipReq.Bandwidth = cmd.Flags().Int("create-eip-bandwidth-mb", 0, "Optional. Required if you want to create new EIP. Bandwidth(Unit:Mbps).The range of value related to network charge mode. By traffic [1, 200]; by bandwidth [1,800] (Unit: Mbps); it could be 0 if the eip belong to the shared bandwidth")
	eipReq.PayMode = cmd.Flags().String("create-eip-charge-mode", "Bandwidth", "Optional. 'Traffic','Bandwidth' or 'ShareBandwidth'")
	eipReq.Name = flags.String("create-eip-name", "", "Optional. Name of created eip to bind with the uhost")
	eipReq.Remark = cmd.Flags().String("create-eip-remark", "", "Optional.Remark of your EIP.")
	eipReq.CouponId = cmd.Flags().String("create-eip-coupon-id", "", "Optional.Coupon ID, The Coupon can deducte part of the payment,see https://accountv2.ucloud.cn")

	req.ChargeType = flags.String("charge-type", "Month", "Optional.'Year',pay yearly;'Month',pay monthly;'Dynamic', pay hourly(requires access)")
	req.Quantity = flags.Int("quantity", 1, "Optional. The duration of the instance. N years/months.")
	req.ProjectId = flags.String("project-id", base.ConfigInstance.ProjectID, "Optional. Assign project-id")
	req.Region = flags.String("region", base.ConfigInstance.Region, "Optional. Assign region")
	req.Zone = flags.String("zone", base.ConfigInstance.Zone, "Optional. Assign availability zone")
	req.UHostType = flags.String("type", defaultUhostType, "Optional. Default is 'N2' of which cpu is V4 and sata disk. also support 'N1' means V3 cpu and sata disk;'I2' means V4 cpu and ssd disk;'D1' means big data model;'G1' means GPU type, model for K80;'G2' model for P40; 'G3' model for V100")
	req.NetCapability = flags.String("net-capability", "Normal", "Optional. Default is 'Normal', also support 'Super' which will enhance multiple times network capability as before")
	req.Disks[0].Type = flags.String("os-disk-type", "LOCAL_NORMAL", "Optional. Enumeration value. 'LOCAL_NORMAL', Ordinary local disk; 'CLOUD_NORMAL', Ordinary cloud disk; 'LOCAL_SSD',local ssd disk; 'CLOUD_SSD',cloud ssd disk; 'EXCLUSIVE_LOCAL_DISK',big data. The disk only supports a limited combination.")
	req.Disks[0].Size = flags.Int("os-disk-size-gb", 20, "Optional. Default 20G. Windows should be bigger than 40G Unit GB")
	req.Disks[0].BackupType = flags.String("os-disk-backup-type", "NONE", "Optional. Enumeration value, 'NONE' or 'DATAARK'. DataArk supports real-time backup, which can restore the disk back to any moment within the last 12 hours. (Normal Local Disk and Normal Cloud Disk Only)")
	req.Disks[1].Type = flags.String("data-disk-type", "LOCAL_NORMAL", "Optional. Enumeration value. 'LOCAL_NORMAL', Ordinary local disk; 'CLOUD_NORMAL', Ordinary cloud disk; 'LOCAL_SSD',local ssd disk; 'CLOUD_SSD',cloud ssd disk; 'EXCLUSIVE_LOCAL_DISK',big data. The disk only supports a limited combination.")
	req.Disks[1].Size = flags.Int("data-disk-size-gb", 20, "Optional. Disk size. Unit GB")
	req.Disks[1].BackupType = flags.String("data-disk-backup-type", "NONE", "Optional. Enumeration value, 'NONE' or 'DATAARK'. DataArk supports real-time backup, which can restore the disk back to any moment within the last 12 hours. (Normal Local Disk and Normal Cloud Disk Only)")
	req.NetworkId = flags.String("network-id", "", "Optional. Network ID (no need to fill in the case of VPC2.0). In the case of VPC1.0, if not filled in, we will choose the basic network; if it is filled in, we will choose the subnet. See 'ucloud subnet list'.")
	req.SecurityGroupId = flags.String("firewall-id", "", "Optional. Firewall Id, default: Web recommended firewall. see 'ucloud firewall list'.")
	req.Tag = flags.String("group", "Default", "Optional. Business group")
	req.CouponId = flags.String("coupon-id", "", "Optional. Coupon ID, The Coupon can deduct part of the payment,see https://accountv2.ucloud.cn")

	cmd.Flags().SetFlagValues("charge-type", "Month", "Year", "Dynamic", "Trial")
	cmd.Flags().SetFlagValues("cpu", "1", "2", "4", "8", "12", "16", "24", "32")
	cmd.Flags().SetFlagValues("type", "N2", "N1", "I2", "D1", "G1", "G2", "G3")
	cmd.Flags().SetFlagValues("net-capability", "Normal", "Super")
	cmd.Flags().SetFlagValues("os-disk-type", "LOCAL_NORMAL", "CLOUD_NORMAL", "LOCAL_SSD", "CLOUD_SSD", "EXCLUSIVE_LOCAL_DISK")
	cmd.Flags().SetFlagValues("os-disk-backup-type", "NONE", "DATAARK")
	cmd.Flags().SetFlagValues("data-disk-type", "LOCAL_NORMAL", "CLOUD_NORMAL", "LOCAL_SSD", "CLOUD_SSD", "EXCLUSIVE_LOCAL_DISK")
	cmd.Flags().SetFlagValues("data-disk-backup-type", "NONE", "DATAARK")
	cmd.Flags().SetFlagValues("create-eip-line", "BGP", "International")
	cmd.Flags().SetFlagValues("create-eip-charge-mode", "Bandwidth", "Traffic", "ShareBandwidth")

	cmd.Flags().SetFlagValuesFunc("image-id", func() []string {
		req := base.BizClient.NewDescribeImageRequest()
		projectID, _ := flags.GetString("project-id")
		if projectID == "" {
			projectID = base.ConfigInstance.ProjectID
		}
		req.ProjectId = sdk.String(projectID)

		region, _ := flags.GetString("region")
		if region == "" {
			region = base.ConfigInstance.Region
		}
		req.Region = sdk.String(region)

		zone, _ := flags.GetString("zone")
		if zone == "" {
			zone = base.ConfigInstance.Zone
		}
		req.Zone = sdk.String(zone)
		req.ImageType = sdk.String("Base")
		req.Limit = sdk.Int(1000)
		result := make([]string, 0)
		resp, err := base.BizClient.DescribeImage(req)
		if err == nil {
			for _, image := range resp.ImageSet {
				if image.State == "Available" {
					imageName := strings.Replace(image.ImageName, " ", "-", -1)
					result = append(result, fmt.Sprintf("%s/%s", image.ImageId, imageName))
				}
			}
		}
		return result
	})

	cmd.MarkFlagRequired("cpu")
	cmd.MarkFlagRequired("memory")
	cmd.MarkFlagRequired("password")
	cmd.MarkFlagRequired("image-id")

	return cmd
}

//NewCmdUHostDelete ucloud uhost delete
func NewCmdUHostDelete() *cobra.Command {
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
				sure, err := ux.Prompt("Are you sure you want to delete this host?")
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
						stopUhostIns(_req, false)
					}
				}
				resp, err := base.BizClient.TerminateUHostInstance(req)
				if err != nil {
					base.HandleError(err)
				} else {
					base.Cxt.Printf("UHost:[%v] deleted\n", resp.UHostId)
				}
			}
		},
	}
	cmd.Flags().SortFlags = false
	uhostIDs = cmd.Flags().StringSlice("resource-id", nil, "Requried. ResourceIDs(UhostIds) of the uhost instance")
	req.ProjectId = cmd.Flags().String("project-id", base.ConfigInstance.ProjectID, "Optional. Assign project-id")
	req.Region = cmd.Flags().String("region", base.ConfigInstance.Region, "Optional. Assign region")
	req.Zone = cmd.Flags().String("zone", "", "Optional. availability zone")
	isDestory = cmd.Flags().Bool("destory", false, "Optional. false,the uhost instance will be thrown to UHost recycle If you have permission; true,the uhost instance will be deleted directly")
	req.ReleaseEIP = cmd.Flags().Bool("release-eip", false, "Optional. false,Unbind EIP only; true, Unbind EIP and release it")
	req.ReleaseUDisk = cmd.Flags().Bool("delete-cloud-disk", false, "Optional.false,Detach cloud disk only; true, Detach cloud disk and delete it")
	yes = cmd.Flags().BoolP("yes", "y", false, "Optional. Do not prompt for confirmation.")
	cmd.Flags().SetFlagValues("destory", "true", "false")
	cmd.Flags().SetFlagValues("release-eip", "true", "false")
	cmd.Flags().SetFlagValues("delete-cloud-disk", "true", "false")
	cmd.Flags().SetFlagValuesFunc("resource-id", func() []string {
		return getUhostList([]string{status.HOST_RUNNING, status.HOST_FAIL, status.HOST_FAIL}, *req.ProjectId, *req.Region, *req.Zone)
	})
	cmd.MarkFlagRequired("resource-id")

	return cmd
}

//NewCmdUHostStop ucloud uhost stop
func NewCmdUHostStop() *cobra.Command {
	var uhostIDs *[]string
	var async *bool
	req := base.BizClient.NewStopUHostInstanceRequest()
	cmd := &cobra.Command{
		Use:     "stop",
		Short:   "Shut down uhost instance",
		Long:    "Shut down uhost instance",
		Example: "ucloud uhost stop --resource-id uhost-xxx1,uhost-xxx2",
		Run: func(cmd *cobra.Command, args []string) {
			for _, id := range *uhostIDs {
				id = base.PickResourceID(id)
				req.UHostId = &id
				stopUhostIns(req, *async)
			}
		},
	}
	cmd.Flags().SortFlags = false
	uhostIDs = cmd.Flags().StringSlice("resource-id", nil, "Required. ResourceIDs(UHostIds) of the uhost instances")
	req.ProjectId = cmd.Flags().String("project-id", base.ConfigInstance.ProjectID, "Optional. Assign project-id")
	req.Region = cmd.Flags().String("region", base.ConfigInstance.Region, "Optional. Assign region")
	req.Zone = cmd.Flags().String("zone", "", "Optional. Assign availability zone")
	async = cmd.Flags().Bool("async", false, "Optional. Do not wait for the long-running operation to finish.")
	cmd.Flags().SetFlagValuesFunc("resource-id", func() []string {
		return getUhostList([]string{status.HOST_RUNNING}, *req.ProjectId, *req.Region, *req.Zone)
	})
	cmd.MarkFlagRequired("resource-id")

	return cmd
}

func stopUhostIns(req *uhost.StopUHostInstanceRequest, async bool) {
	resp, err := base.BizClient.StopUHostInstance(req)
	if err != nil {
		base.HandleError(err)
	} else {
		text := fmt.Sprintf("UHost:[%v] is shutting down", resp.UhostId)
		if async {
			base.Cxt.Println(text)
		} else {
			done := pollUhost(resp.UhostId, *req.ProjectId, *req.Region, *req.Zone, []string{status.HOST_STOPPED, status.HOST_FAIL})
			ux.DotSpinner.Start(text)
			<-done
			ux.DotSpinner.Stop()
		}
	}
}

//NewCmdUHostStart ucloud uhost start
func NewCmdUHostStart() *cobra.Command {
	var async *bool
	var uhostIDs *[]string
	req := base.BizClient.NewStartUHostInstanceRequest()
	cmd := &cobra.Command{
		Use:     "start",
		Short:   "Start Uhost instance",
		Long:    "Start Uhost instance",
		Example: "ucloud uhost start --resource-id uhost-xxx1,uhost-xxx2",
		Run: func(cmd *cobra.Command, args []string) {
			for _, id := range *uhostIDs {
				id := base.PickResourceID(id)
				req.UHostId = &id
				resp, err := base.BizClient.StartUHostInstance(req)
				if err != nil {
					base.HandleError(err)
				} else {
					text := fmt.Sprintf("UHost:[%v] is starting", resp.UhostId)
					if *async {
						base.Cxt.Println(text)
					} else {
						done := pollUhost(resp.UhostId, *req.ProjectId, *req.Region, *req.Zone, []string{status.HOST_RUNNING, status.HOST_FAIL})
						dotSpinner := ux.NewDotSpinner()
						dotSpinner.Start(text)
						<-done
						dotSpinner.Stop()
					}
				}
			}
		},
	}
	cmd.Flags().SortFlags = false
	uhostIDs = cmd.Flags().StringSlice("resource-id", nil, "Requried. ResourceIDs(UHostIds) of the uhost instance")
	req.ProjectId = cmd.Flags().String("project-id", base.ConfigInstance.ProjectID, "Optional. Assign project-id")
	req.Region = cmd.Flags().String("region", base.ConfigInstance.Region, "Optional. Assign region")
	req.Zone = cmd.Flags().String("zone", "", "Optional. Assign availability zone")
	req.DiskPassword = cmd.Flags().String("disk-password", "", "Optional. Encrypted disk password")
	async = cmd.Flags().Bool("async", false, "Optional. Do not wait for the long-running operation to finish.")
	cmd.Flags().SetFlagValuesFunc("resource-id", func() []string {
		return getUhostList([]string{status.HOST_STOPPED}, *req.ProjectId, *req.Region, *req.Zone)
	})
	cmd.MarkFlagRequired("resource-id")
	return cmd
}

//NewCmdUHostReboot ucloud uhost restart
func NewCmdUHostReboot() *cobra.Command {
	var uhostIDs *[]string
	var async *bool
	req := base.BizClient.NewRebootUHostInstanceRequest()
	cmd := &cobra.Command{
		Use:     "restart",
		Short:   "Restart uhost instance",
		Long:    "Restart uhost instance",
		Example: "ucloud uhost restart --resource-id uhost-xxx1,uhost-xxx2",
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
						base.Cxt.Println(text)
					} else {
						done := pollUhost(resp.UhostId, *req.ProjectId, *req.Region, *req.Zone, []string{status.HOST_RUNNING, status.HOST_FAIL})
						ux.DotSpinner.Start(text)
						<-done
						ux.DotSpinner.Stop()
					}
				}
			}
		},
	}
	cmd.Flags().SortFlags = false
	uhostIDs = cmd.Flags().StringSlice("resource-id", nil, "Required. ResourceIDs(UHostIds) of the uhost instance")
	req.ProjectId = cmd.Flags().String("project-id", base.ConfigInstance.ProjectID, "Optional. Assign project-id")
	req.Region = cmd.Flags().String("region", base.ConfigInstance.Region, "Optional. Assign region")
	req.Zone = cmd.Flags().String("zone", "", "Optional. Assign availability zone")
	req.DiskPassword = cmd.Flags().String("disk-password", "", "Optional. Encrypted disk password")
	async = cmd.Flags().Bool("async", false, "Optional. Do not wait for the long-running operation to finish.")
	cmd.Flags().SetFlagValuesFunc("resource-id", func() []string {
		return getUhostList([]string{status.HOST_FAIL, status.HOST_RUNNING, status.HOST_STOPPED}, *req.ProjectId, *req.Region, *req.Zone)
	})
	cmd.MarkFlagRequired("resource-id")
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
		Example: "ucloud uhost poweroff --resource-id uhost-xxx1,uhost-xxx2",
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
	uhostIDs = cmd.Flags().StringSlice("resource-id", nil, "ResourceIDs(UHostIds) of the uhost instance")
	req.ProjectId = cmd.Flags().String("project-id", base.ConfigInstance.ProjectID, "Assign project-id")
	req.Region = cmd.Flags().String("region", base.ConfigInstance.Region, "Assign region")
	req.Zone = cmd.Flags().String("zone", "", "Assign availability zone")
	yes = cmd.Flags().BoolP("yes", "y", false, "Optional. Do not prompt for confirmation.")
	cmd.MarkFlagRequired("resource-id")
	return cmd
}

//NewCmdUHostResize ucloud uhost resize
func NewCmdUHostResize() *cobra.Command {
	var yes *bool
	var uhostIDs *[]string
	req := base.BizClient.NewResizeUHostInstanceRequest()
	cmd := &cobra.Command{
		Use:     "resize",
		Short:   "Resize uhost instance,such as cpu core count, memory size and disk size",
		Long:    "Resize uhost instance,such as cpu core count, memory size and disk size",
		Example: "ucloud uhost resize --resource-id uhost-xxx1,uhost-xxx2 --cpu 4 --memory-gb 8",
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
					stopUhostIns(_req, false)
				}

				resp, err := base.BizClient.ResizeUHostInstance(req)
				if err != nil {
					base.HandleError(err)
				} else {
					base.Cxt.Printf("UHost:[%v] resized\n", resp.UhostId)
				}
			}
		},
	}
	cmd.Flags().SortFlags = false
	uhostIDs = cmd.Flags().StringSlice("resource-id", nil, "Required. ResourceIDs(or UhostIDs) of the uhost instances")
	req.ProjectId = cmd.Flags().String("project-id", base.ConfigInstance.ProjectID, "Optional. Assign project-id")
	req.Region = cmd.Flags().String("region", base.ConfigInstance.Region, "Optional. Assign region")
	req.Zone = cmd.Flags().String("zone", "", "Optional. Assign availability zone")
	req.CPU = cmd.Flags().Int("cpu", 0, "Optional. The number of virtual CPU cores. Series1 {1, 2, 4, 8, 12, 16, 24, 32}. Series2 {1,2,4,8,16}")
	req.Memory = cmd.Flags().Int("memory-gb", 0, "Optional. memory size. Unit: GB. Range: [1, 128], multiple of 2")
	req.DiskSpace = cmd.Flags().Int("data-disk-size-gb", 0, "Optional. Data disk size,unit GB. Range[10,1000], SSD disk range[100,500]. Step 10")
	req.BootDiskSpace = cmd.Flags().Int("system-disk-size-gb", 0, "Optional. System disk size, unit GB. Range[20,100]. Step 10. System disk does not support shrinkage")
	req.NetCapValue = cmd.Flags().Int("net-cap", 0, "Optional. NIC scale. 1,upgrade; 2,downgrade; 0,unchanged")
	yes = cmd.Flags().BoolP("yes", "y", false, "Optional. Do not prompt for confirmation.")
	cmd.Flags().SetFlagValuesFunc("resource-id", func() []string {
		return getUhostList([]string{status.HOST_RUNNING, status.HOST_STOPPED, status.HOST_FAIL}, *req.ProjectId, *req.Region, *req.Zone)
	})
	cmd.MarkFlagRequired("resource-id")
	return cmd
}

var pollUhost = base.Poll(describeUHostByID)

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
