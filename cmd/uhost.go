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
	"strings"

	"github.com/ucloud/ucloud-sdk-go/sdk"

	"github.com/spf13/cobra"
	"github.com/ucloud/ucloud-sdk-go/services/uhost"

	. "github.com/ucloud/ucloud-cli/util"
)

//NewCmdUHost ucloud uhost
func NewCmdUHost() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uhost",
		Short: "List,create,delete,stop,restart,poweroff or scale UHost instance",
		Long:  `List,create,delete,stop,restart,poweroff or scale UHost instance`,
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(NewCmdUHostList())
	cmd.AddCommand(NewCmdUHostCreate())
	cmd.AddCommand(NewCmdUHostDelete())
	cmd.AddCommand(NewCmdUHostStop())
	cmd.AddCommand(NewCmdUHostStart())
	cmd.AddCommand(NewCmdUHostReboot())
	cmd.AddCommand(NewCmdUHostPoweroff())
	cmd.AddCommand(NewCmdUHostScale())

	return cmd
}

//UHostRow UHost表格行
type UHostRow struct {
	UHostName      string
	ResourceID     string
	UGroup         string
	ClassicNetwork string
	Config         string
	Type           string
	CreationTime   string
	State          string
}

//NewCmdUHostList [ucloud uhost list]
func NewCmdUHostList() *cobra.Command {
	req := BizClient.NewDescribeUHostInstanceRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all UHost Instances",
		Long:  `List all UHost Instances`,
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := BizClient.DescribeUHostInstance(req)
			if err != nil {
				Cxt.Println(err)
				return
			}
			if resp.RetCode != 0 {
				HandleBizError(resp)
			} else {
				if global.json {
					PrintJSON(resp.UHostSet)
				} else {
					list := make([]UHostRow, 0)
					for _, host := range resp.UHostSet {
						row := UHostRow{}
						row.UHostName = host.Name
						row.ResourceID = host.UHostId
						row.UGroup = host.Tag
						for _, ip := range host.IPSet {
							if row.ClassicNetwork != "" {
								row.ClassicNetwork += " | "
							}
							if ip.Type == "Private" {
								row.ClassicNetwork += fmt.Sprintf("%s", ip.IP)
							} else {
								row.ClassicNetwork += fmt.Sprintf("%s %s", ip.IP, ip.Type)
							}
						}
						osName := strings.SplitN(host.OsName, " ", 2)
						cupCore := host.CPU
						memorySize := host.Memory / 1024
						diskSize := 0
						for _, disk := range host.DiskSet {
							if disk.Type == "Data" {
								diskSize += disk.Size
							}
						}
						row.Config = fmt.Sprintf("%s cpu:%d memory:%dG disk:%dG", osName[0], cupCore, memorySize, diskSize)
						row.CreationTime = FormatDate(host.CreateTime)
						row.State = host.State
						row.Type = host.UHostType + "/" + host.HostType
						list = append(list, row)
					}
					PrintTable(list, []string{"UHostName", "ResourceID", "UGroup", "ClassicNetwork", "Config", "Type", "CreationTime", "State"})
				}
			}
		},
	}
	cmd.Flags().SortFlags = false
	req.Region = cmd.Flags().String("region", ConfigInstance.Region, "Assign region(override default region of your config)")
	req.Zone = cmd.Flags().String("zone", "", "Zone")
	req.ProjectId = cmd.Flags().String("project-id", ConfigInstance.ProjectID, "Assign project-id(override default projec-id of your config)")
	cmd.Flags().StringSliceVar(&req.UHostIds, "resource-id", make([]string, 0), "UHost Instance ID, multiple values separated by comma(without space)")
	req.Tag = cmd.Flags().String("tag", "", "UGroup")
	req.Offset = cmd.Flags().Int("offset", 0, "offset default 0")
	req.Limit = cmd.Flags().Int("limit", 20, "limit default 20, max value 100")

	return cmd
}

//NewCmdUHostCreate [ucloud uhost create]
func NewCmdUHostCreate() *cobra.Command {
	req := BizClient.NewCreateUHostInstanceRequest()
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create UHost instance",
		Long:  "Create UHost instance",
		Run: func(cmd *cobra.Command, args []string) {
			*req.Memory *= 1024
			*req.Password = base64.StdEncoding.EncodeToString([]byte(*req.Password))
			req.LoginMode = sdk.String("Password")
			// req.UHostType = sdk.String("Normal")
			// req.HostType = sdk.String("N2")
			resp, err := BizClient.CreateUHostInstance(req)
			if err != nil {
				Cxt.PrintErr(err)
				return
			}
			if resp.RetCode != 0 {
				HandleBizError(resp)
			} else {
				Cxt.Printf("UHost:%v created successfully!\n", resp.UHostIds)
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
	if _, ok := n1Zone[ConfigInstance.Zone]; ok {
		defaultUhostType = "N1"
	}

	req.Disks = make([]uhost.UHostDisk, 2)
	req.Disks[0].IsBoot = sdk.Bool(true)
	req.Disks[0].Size = sdk.String("20")
	req.Disks[1].IsBoot = sdk.Bool(false)

	flags := cmd.Flags()
	cmd.Flags().SortFlags = false
	req.ProjectId = cmd.Flags().String("project-id", ConfigInstance.ProjectID, "Assign project-id(override default projec-id of your config)")
	req.Region = cmd.Flags().String("region", ConfigInstance.Region, "Assign region(override default region of your config)")
	req.Zone = cmd.Flags().String("zone", ConfigInstance.Zone, "Zone")
	req.UHostType = flags.String("type", defaultUhostType, "Uhost type. enumeration value. 'N1': series 1 standard type; 'N2': series 2 standard type; 'I1': series 1 high io type; 'I2', series 2 high IO type; 'D1': Series 1 big data model; 'G1': Series 1 GPU type, model For K80; 'G2': Series 2 GPU type, model is P40; 'G3': Series 2 GPU type, model is V100.")
	req.NetCapability = flags.String("net-capability", "Normal", "enumeration value. 'Normal' or 'Super'")
	req.CPU = flags.Int("cpu", 4, "The number of virtual CPU cores. Optional parameters: {1, 2, 4, 8, 12, 16, 24, 32}")
	req.Memory = flags.Int("memory", 8, "memory size. Unit: GB. Range: [1, 128], multiple of 2")
	req.ImageId = cmd.Flags().String("image-id", "", "The ID of image. Obtain by 'ucloud image list'")
	req.Password = cmd.Flags().String("password", "", "Password of the uhost user")
	req.Disks[0].Type = cmd.Flags().String("system-disk-type", "LOCAL_NORMAL", "Enumeration value. 'LOCAL_NORMAL', Ordinary local disk; 'CLOUD_NORMAL', Ordinary cloud disk; 'LOCAL_SSD SSD',local ssd disk; 'CLOUD_SSD SSD',cloud ssd disk; 'EXCLUSIVE_LOCAL_DISK',big data. The disk only supports a limited combination.")
	req.Disks[0].BackupType = cmd.Flags().String("system-disk-backup-type", "NONE", "Enumeration value, 'NONE' or 'DATAARK'. DataArk supports real-time backup, which can restore the disk back to any moment within the last 12 hours. (Normal Local Disk and Normal Cloud Disk Only)")
	req.Disks[1].Type = cmd.Flags().String("data-disk-type", "LOCAL_NORMAL", "Enumeration value. 'LOCAL_NORMAL', Ordinary local disk; 'CLOUD_NORMAL', Ordinary cloud disk; 'LOCAL_SSD SSD',local ssd disk; 'CLOUD_SSD SSD',cloud ssd disk; 'EXCLUSIVE_LOCAL_DISK',big data. The disk only supports a limited combination.")
	req.Disks[1].Size = cmd.Flags().String("data-disk-size", "20", "Disk size. Unit GB")
	req.Disks[1].BackupType = cmd.Flags().String("data-disk-backup-type", "NONE", "Enumeration value, 'NONE' or 'DATAARK'. DataArk supports real-time backup, which can restore the disk back to any moment within the last 12 hours. (Normal Local Disk and Normal Cloud Disk Only)")
	req.NetworkId = flags.String("network-id", "", "Network ID (no need to fill in the case of VPC2.0). In the case of VPC1.0, if not filled in, we will choose the basic network; if it is filled in, we will choose the subnet. See DescribeSubnet.")
	req.VPCId = flags.String("vpc-id", "", "VPC ID. This field is required under VPC2.0")
	req.SubnetId = flags.String("subnet-id", "", "Subnet ID. This field is required under VPC2.0")
	req.SecurityGroupId = flags.String("firewall-id", "", "Firewall Id, default: Web recommended firewall. see DescribeSecurityGroup.")
	req.Tag = flags.String("ugroup", "Default", "Business group")
	req.Name = flags.String("name", "UHost", "UHost instance name")
	req.ChargeType = flags.String("charge-type", "Month", "Enumeration value. 'Year', pay per year; 'Month', pay per month; 'Dynamic', pay per hour (requires access). The default is monthly payment")
	req.Quantity = flags.Int("quantity", 1, "The length of purchase. This parameter is not required when purchasing by the hour (Dynamic). When the monthly payment is made, the parameter 0 means the purchase until the end of the month.")
	req.CouponId = flags.String("coupon-id", "", "Coupon ID, The Coupon can deduct part of the payment,see DescribeCoupon or https://accountv2.ucloud.cn")

	cmd.MarkFlagRequired("image-id")
	cmd.MarkFlagRequired("password")

	return cmd
}

//NewCmdUHostDelete ucloud uhost delete
func NewCmdUHostDelete() *cobra.Command {
	isDestory := sdk.Bool(false)
	isEipReleased := sdk.Bool(false)

	req := BizClient.NewTerminateUHostInstanceRequest()
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete Uhost instance",
		Long:  "Delete Uhost instance",
		Run: func(cmd *cobra.Command, args []string) {
			if *isDestory {
				req.Destroy = sdk.Int(1)
			} else {
				req.Destroy = sdk.Int(0)
			}
			if *isEipReleased {
				req.EIPReleased = sdk.String("yes")
			} else {
				req.EIPReleased = sdk.String("no")
			}
			resp, err := BizClient.TerminateUHostInstance(req)
			if err != nil {
				Cxt.PrintErr(err)
			} else {
				if resp.RetCode != 0 {
					HandleBizError(resp)
				} else {
					Cxt.Printf("UHost:%v deleted successfully!\n", resp.UHostId)
				}
			}
		},
	}

	req.ProjectId = cmd.Flags().String("project-id", ConfigInstance.ProjectID, "Assign project-id(override default projec-id of your config)")
	req.Region = cmd.Flags().String("region", ConfigInstance.Region, "Assign region(override default region of your config)")
	req.Zone = cmd.Flags().String("zone", "", "Zone")
	req.UHostId = cmd.Flags().String("resource-id", "", "ResourceID of the uhost instance( or UHostId)")
	isDestory = cmd.Flags().Bool("destory", false, "false,the uhost instance will be thrown to UHost recycle If you have permission; true,the uhost instance will be deleted directly")
	isEipReleased = cmd.Flags().Bool("eip-released", false, "false,Unbind EIP only; true, Unbind EIP and release it")
	cmd.MarkFlagRequired("resource-id")

	return cmd
}

//NewCmdUHostStop ucloud uhost stop
func NewCmdUHostStop() *cobra.Command {
	req := BizClient.NewStopUHostInstanceRequest()
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Shut down uhost instance",
		Long:  "Shut down uhost instance",
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := BizClient.StopUHostInstance(req)
			if err != nil {
				Cxt.PrintErr(err)
			} else {
				if resp.RetCode != 0 {
					HandleBizError(resp)
				} else {
					fmt.Println(resp)
					Cxt.Printf("UHost:%v is shuting down. Wait a moment\n", resp.UhostId)
				}
			}
		},
	}
	req.ProjectId = cmd.Flags().String("project-id", ConfigInstance.ProjectID, "Assign project-id(override default projec-id of your config)")
	req.Region = cmd.Flags().String("region", ConfigInstance.Region, "Assign region(override default region of your config)")
	req.Zone = cmd.Flags().String("zone", ConfigInstance.Zone, "Zone")
	req.UHostId = cmd.Flags().String("resource-id", "", "ResourceID of the uhost instance( or UHostId)")
	cmd.MarkFlagRequired("resource-id")

	return cmd
}

//NewCmdUHostStart ucloud uhost start
func NewCmdUHostStart() *cobra.Command {
	req := BizClient.NewStartUHostInstanceRequest()
	cmd := &cobra.Command{
		Use:     "start",
		Short:   "Start Uhost instance",
		Long:    "Start Uhost instance",
		Example: "ucloud uhost start --resource-id uhost-xxx",
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := BizClient.StartUHostInstance(req)
			if err != nil {
				Cxt.PrintErr(err)
			} else {
				if resp.RetCode != 0 {
					HandleBizError(resp)
				} else {
					Cxt.Printf("UHost:%v is starting. Wait a moment\n", resp.UhostId)
				}
			}
		},
	}
	req.ProjectId = cmd.Flags().String("project-id", ConfigInstance.ProjectID, "Assign project-id(override default projec-id of your config)")
	req.Region = cmd.Flags().String("region", ConfigInstance.Region, "Assign region(override default region of your config)")
	req.Zone = cmd.Flags().String("zone", "", "Zone")
	req.UHostId = cmd.Flags().String("resource-id", "", "ResourceID of the uhost instance( or UHostId)")
	req.DiskPassword = cmd.Flags().String("disk-password", "", "Encrypted disk password")
	cmd.MarkFlagRequired("resource-id")
	return cmd
}

//NewCmdUHostReboot ucloud uhost restart
func NewCmdUHostReboot() *cobra.Command {
	req := BizClient.NewRebootUHostInstanceRequest()
	cmd := &cobra.Command{
		Use:     "restart",
		Short:   "Restart/reboot Uhost instance",
		Long:    "Restart/reboot Uhost instance",
		Example: "ucloud uhost restart --resource-id uhost-xxx",
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := BizClient.RebootUHostInstance(req)
			if err != nil {
				Cxt.PrintErr(err)
			} else {
				if resp.RetCode != 0 {
					HandleBizError(resp)
				} else {
					Cxt.Printf("UHost:%v is restarting. Wait a moment\n", resp.UhostId)
				}
			}
		},
	}
	req.ProjectId = cmd.Flags().String("project-id", ConfigInstance.ProjectID, "Assign project-id(override default projec-id of your config)")
	req.Region = cmd.Flags().String("region", ConfigInstance.Region, "Assign region(override default region of your config)")
	req.Zone = cmd.Flags().String("zone", "", "Zone")
	req.UHostId = cmd.Flags().String("resource-id", "", "ResourceID of the uhost instance( or UHostId)")
	req.DiskPassword = cmd.Flags().String("disk-password", "", "Encrypted disk password")
	cmd.MarkFlagRequired("resource-id")
	return cmd
}

//NewCmdUHostPoweroff ucloud uhost poweroff
func NewCmdUHostPoweroff() *cobra.Command {
	req := BizClient.NewPoweroffUHostInstanceRequest()
	cmd := &cobra.Command{
		Use:   "poweroff",
		Short: "Analog power off Uhost instnace. Danger! this operation may affect data integrity or cause file system corruption",
		Long:  "Analog power off Uhost instnace. Danger! this operation may affect data integrity or cause file system corruption",
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := BizClient.PoweroffUHostInstance(req)
			if err != nil {
				Cxt.PrintErr(err)
			} else {
				if resp.RetCode != 0 {
					HandleBizError(resp)
				} else {
					Cxt.Printf("UHost:%v is power off\n", resp.UhostId)
				}
			}
		},
	}
	req.ProjectId = cmd.Flags().String("project-id", ConfigInstance.ProjectID, "Assign project-id(override default projec-id of your config)")
	req.Region = cmd.Flags().String("region", ConfigInstance.Region, "Assign region(override default region of your config)")
	req.Zone = cmd.Flags().String("zone", "", "Zone")
	req.UHostId = cmd.Flags().String("resource-id", "", "ResourceID of the uhost instance( or UHostId)")
	return cmd
}

//NewCmdUHostScale ucloud uhost scale
func NewCmdUHostScale() *cobra.Command {
	req := BizClient.NewResizeUHostInstanceRequest()
	cmd := &cobra.Command{
		Use:   "scale",
		Short: "Scale uhost instance,such as cpu core count, memory size, system disk size and data disk size",
		Long:  "Scale uhost instance,such as cpu core count, memory size, system disk size and data disk size",
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
			resp, err := BizClient.ResizeUHostInstance(req)
			if err != nil {
				Cxt.PrintErr(err)
			} else {
				if resp.RetCode != 0 {
					HandleBizError(resp)
				} else {
					Cxt.Printf("UHost:%v scaled\n", resp.UhostId)
				}
			}

		},
	}
	req.ProjectId = cmd.Flags().String("project-id", ConfigInstance.ProjectID, "Assign project-id(override default projec-id of your config)")
	req.Region = cmd.Flags().String("region", ConfigInstance.Region, "Assign region(override default region of your config)")
	req.Zone = cmd.Flags().String("zone", "", "Zone")
	req.UHostId = cmd.Flags().String("resource-id", "", "ResourceID of the uhost instance( or UHostId)")
	req.CPU = cmd.Flags().Int("cpu", 0, "The number of virtual CPU cores. Series1 {1, 2, 4, 8, 12, 16, 24, 32}. Series2 {1,2,4,8,16}")
	req.Memory = cmd.Flags().Int("memory", 0, "memory size. Unit: GB. Range: [1, 128], multiple of 2")
	req.DiskSpace = cmd.Flags().Int("data-disk-size", 0, "Data disk size,unit GB. Range[10,1000], SSD disk range[100,500]. Step 10")
	req.BootDiskSpace = cmd.Flags().Int("system-disk-size", 0, "System disk size, unit GB. Range[20,100]. Step 10. System disk does not support shrinkage")
	req.NetCapValue = cmd.Flags().Int("net-cap", 0, "NIC scale. 1,upgrade; 2,downgrade; 0,unchanged")
	cmd.MarkFlagRequired("resource-id")
	return cmd
}
