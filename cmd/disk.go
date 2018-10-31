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
	"strings"

	"github.com/spf13/cobra"

	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/base"
	"github.com/ucloud/ucloud-cli/model/status"
	"github.com/ucloud/ucloud-cli/ux"
)

//NewCmdDisk ucloud disk
func NewCmdDisk() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disk",
		Short: "Read and manipulate ucloud disks",
		Long:  "Read and manipulate ucloud disks",
	}
	cmd.AddCommand(NewCmdDiskCreate())
	cmd.AddCommand(NewCmdDiskList())
	cmd.AddCommand(NewCmdDiskAttach())
	cmd.AddCommand(NewCmdDiskDetach())
	cmd.AddCommand(NewCmdDiskDelete())
	cmd.AddCommand(NewCmdDiskClone())
	cmd.AddCommand(NewCmdDiskExpand())
	return cmd
}

//NewCmdDiskCreate ucloud disk create
func NewCmdDiskCreate() *cobra.Command {
	req := base.BizClient.NewCreateUDiskRequest()
	enableDataArk := sdk.String("false")
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create ucloud disk",
		Long:  "Create ucloud disk",
		Run: func(cmd *cobra.Command, args []string) {
			if *enableDataArk == "true" {
				req.UDataArkMode = sdk.String("Yes")
			} else {
				req.UDataArkMode = sdk.String("No")
			}

			if *req.DiskType == "Oridinary" {
				*req.DiskType = "DataDisk"
			} else if *req.DiskType == "SSD" {
				*req.DiskType = "SSDDataDisk"
			}

			resp, err := base.BizClient.CreateUDisk(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			if count := len(resp.UDiskId); count == 1 {
				text := fmt.Sprintf("udisk:%v is initializating", resp.UDiskId)
				pollDisk(resp.UDiskId[0], *req.ProjectId, *req.Region, *req.Zone, text, []string{status.DISK_AVAILABLE, status.DISK_FAILED})
			} else if count > 1 {
				base.Cxt.Printf("udisk:%v created\n", resp.UDiskId)
			} else {
				base.Cxt.PrintErr(fmt.Errorf("none disk created"))
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.ProjectId = flags.String("project-id", base.ConfigInstance.ProjectID, "Optional. Assign project-id")
	req.Region = flags.String("region", base.ConfigInstance.Region, "Optional. Assign region")
	req.Zone = flags.String("zone", base.ConfigInstance.Zone, "Optional. Assign availability zone")
	req.Name = flags.String("name", "", "Required. Name of the disk to create")
	req.Size = flags.Int("size-gb", 10, "Required. Size of the disk to create. Unit:GB. Normal disk [1,8000]; SSD disk [1,4000] ")
	req.ChargeType = flags.String("charge-type", "Dynamic", "Optional.'Year',pay yearly;'Month',pay monthly;'Dynamic', pay hourly")
	req.Quantity = flags.Int("quantity", 1, "Optional. The duration of the instance. N years/months.")
	enableDataArk = flags.String("enable-data-ark", "false", "Optional. DataArk supports real-time backup, which can restore the disk back to any moment within the last 12 hours.")
	req.Tag = flags.String("group", "Default", "Optional. Business group")
	req.DiskType = flags.String("disk-type", "Oridinary", "Optional. 'Ordinary' or 'SSD'")
	req.CouponId = flags.String("coupon-id", "", "Optional. Coupon ID, The Coupon can deduct part of the payment.See https://accountv2.ucloud.cn")

	flags.SetFlagValues("charge-type", "Month", "Year", "Dynamic", "Trial")
	flags.SetFlagValues("enable-data-ark", "true", "false")
	flags.SetFlagValues("disk-type", "Oridinary", "SSD")

	cmd.MarkFlagRequired("size-gb")
	cmd.MarkFlagRequired("name")

	return cmd
}

//DiskRow TableRow
type DiskRow struct {
	ResourceID     string
	Name           string
	Group          string
	Size           string
	Type           string
	MountUHost     string
	MountPoint     string
	EnableDataArk  string
	State          string
	CreationTime   string
	ExpirationTime string
}

//NewCmdDiskList ucloud disk list
func NewCmdDiskList() *cobra.Command {
	req := base.BizClient.NewDescribeUDiskRequest()
	typeMap := map[string]string{
		"DataDisk":    "Oridinary-Data-Disk",
		"SystemDisk":  "Oridinary-System-Disk",
		"SSDDataDisk": "SSD-Data-Disk",
	}
	arkModeMap := map[string]string{
		"Yes": "true",
		"No":  "false",
	}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List ucloud disk",
		Long:  "List ucloud disk",
		Run: func(cmd *cobra.Command, args []string) {
			for key, val := range typeMap {
				if *req.DiskType == val {
					*req.DiskType = key
				}
			}
			resp, err := base.BizClient.DescribeUDisk(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			list := []DiskRow{}
			for _, disk := range resp.DataSet {
				row := DiskRow{
					ResourceID:     disk.UDiskId,
					Name:           disk.Name,
					Group:          disk.Tag,
					Size:           fmt.Sprintf("%dGB", disk.Size),
					Type:           typeMap[disk.DiskType],
					EnableDataArk:  arkModeMap[disk.UDataArkMode],
					MountUHost:     disk.UHostIP,
					MountPoint:     fmt.Sprintf("%s", disk.DeviceName),
					State:          disk.Status,
					CreationTime:   base.FormatDate(disk.CreateTime),
					ExpirationTime: base.FormatDate(disk.ExpiredTime),
				}
				list = append(list, row)
			}
			base.PrintTableS(list)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.ProjectId = flags.String("project-id", base.ConfigInstance.ProjectID, "Optional. Assign project-id")
	req.Region = flags.String("region", base.ConfigInstance.Region, "Optional. Assign region")
	req.Zone = flags.String("zone", base.ConfigInstance.Zone, "Optional. Assign availability zone")
	req.UDiskId = flags.String("resource-id", "", "Optional. Resource ID of the disk to search")
	req.DiskType = flags.String("disk-type", "", "Optional. Optional. Type of the disk to search. 'Oridinary-Data-Disk','Oridinary-System-Disk' or 'SSD-Data-Disk'")
	req.Offset = cmd.Flags().Int("offset", 0, "Optional. Offset")
	req.Limit = cmd.Flags().Int("limit", 50, "Optional. Limit")
	flags.SetFlagValues("disk-type", "Oridinary-Data-Disk", "Oridinary-System-Disk", "SSD-Data-Disk")
	return cmd
}

//NewCmdDiskAttach ucloud disk attach
func NewCmdDiskAttach() *cobra.Command {
	req := base.BizClient.NewAttachUDiskRequest()
	cmd := &cobra.Command{
		Use:   "attach",
		Short: "Attach a disk to an uhost instance",
		Long:  "Attach a disk to an uhost instance",
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := base.BizClient.AttachUDisk(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			text := fmt.Sprintf("udisk[%s] is attaching to uhost uhost[%s]", *req.UDiskId, *req.UHostId)
			pollDisk(resp.UDiskId, *req.ProjectId, *req.Region, *req.Zone, text, []string{status.DISK_INUSE, status.DISK_FAILED})
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.ProjectId = flags.String("project-id", base.ConfigInstance.ProjectID, "Optional. Assign project-id")
	req.Region = flags.String("region", base.ConfigInstance.Region, "Optional. Assign region")
	req.Zone = flags.String("zone", base.ConfigInstance.Zone, "Optional. Assign availability zone")
	req.UHostId = flags.String("uhost-id", "", "Resource ID of the uhost instance which you want to attach the disk")
	req.UDiskId = flags.String("disk-id", "", "Resource ID of the udisk instance to attach")

	cmd.MarkFlagRequired("uhost-id")
	cmd.MarkFlagRequired("disk-id")

	return cmd
}

//NewCmdDiskDetach ucloud disk detach
func NewCmdDiskDetach() *cobra.Command {
	req := base.BizClient.NewDetachUDiskRequest()
	cmd := &cobra.Command{
		Use:   "detach",
		Short: "Detach an ucloud disk from uhost",
		Long:  "Detach an ucloud disk from uhost",
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := base.BizClient.DetachUDisk(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			text := fmt.Sprintf("udisk[%s] is detaching from uhost[%s]", resp.UDiskId, resp.UHostId)
			pollDisk(resp.UDiskId, *req.ProjectId, *req.Region, *req.Zone, text, []string{status.DISK_AVAILABLE, status.DISK_FAILED})
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.ProjectId = flags.String("project-id", base.ConfigInstance.ProjectID, "Optional. Assign project-id")
	req.Region = flags.String("region", base.ConfigInstance.Region, "Optional. Assign region")
	req.Zone = flags.String("zone", base.ConfigInstance.Zone, "Optional. Assign availability zone")
	req.UHostId = flags.String("uhost-id", "", "Resource ID of the uhost instance, from which you want to detach the disk")
	req.UDiskId = flags.String("disk-id", "", "Resource ID of the udisk instance to detach")

	cmd.MarkFlagRequired("uhost-id")
	cmd.MarkFlagRequired("disk-id")
	return cmd
}

//NewCmdDiskDelete ucloud disk delete
func NewCmdDiskDelete() *cobra.Command {
	req := base.BizClient.NewDeleteUDiskRequest()
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete an ucloud disk",
		Long:  "Delete an ucloud disk",
		Run: func(cmd *cobra.Command, args []string) {
			if strings.Index(*req.UDiskId, "/") > -1 {
				*req.UDiskId = strings.SplitN(*req.UDiskId, "/", 2)[0]
			}
			sure, err := ux.Prompt(fmt.Sprintf("Are you sure to delete disk[%s]?", *req.UDiskId))
			if err != nil {
				base.Cxt.PrintErr(err)
				return
			}
			if !sure {
				return
			}
			_, err = base.BizClient.DeleteUDisk(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			base.Cxt.Printf("udisk[%s] deleted\n", *req.UDiskId)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.ProjectId = flags.String("project-id", base.ConfigInstance.ProjectID, "Optional. Assign project-id")
	req.Region = flags.String("region", base.ConfigInstance.Region, "Optional. Assign region")
	req.Zone = flags.String("zone", base.ConfigInstance.Zone, "Optional. Assign availability zone")
	req.UDiskId = flags.String("resource-id", "", "Required. The Resource ID of a disk to delete")

	flags.SetFlagValuesFunc("resource-id", func() []string {
		return getDiskList([]string{status.DISK_AVAILABLE, status.DISK_FAILED}, *req.ProjectId, *req.Region, *req.Zone)
	})

	cmd.MarkFlagRequired("resource-id")

	return cmd
}

//NewCmdDiskClone ucloud disk clone
func NewCmdDiskClone() *cobra.Command {
	req := base.BizClient.NewCloneUDiskRequest()
	enableDataArk := sdk.String("false")
	cmd := &cobra.Command{
		Use:   "clone",
		Short: "Clone an ucloud disk",
		Long:  "Clone an ucloud disk",
		Run: func(cmd *cobra.Command, args []string) {
			if *enableDataArk == "true" {
				req.UDataArkMode = sdk.String("Yes")
			} else {
				req.UDataArkMode = sdk.String("No")
			}
			if strings.Index(*req.SourceId, "/") > -1 {
				*req.SourceId = strings.SplitN(*req.SourceId, "/", 2)[0]
			}
			resp, err := base.BizClient.CloneUDisk(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			if len(resp.UDiskId) == 1 {
				pollText := fmt.Sprintf("cloned disk:[%s] is initializating", resp.UDiskId[0])
				pollDisk(resp.UDiskId[0], *req.ProjectId, *req.Region, *req.Zone, pollText, []string{status.DISK_AVAILABLE, status.DISK_FAILED})
			} else {
				base.Cxt.Printf("disk[%v] cloned", resp.UDiskId)
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.SourceId = flags.String("source-id", "", "Required. Resource ID of parent disk")
	req.Name = flags.String("name", "", "Required. Name of new disk")
	req.ProjectId = flags.String("project-id", base.ConfigInstance.ProjectID, "Optional. Assign project-id")
	req.Region = flags.String("region", base.ConfigInstance.Region, "Optional. Assign region")
	req.Zone = flags.String("zone", base.ConfigInstance.Zone, "Optional. Assign availability zone")
	req.ChargeType = flags.String("charge-type", "Month", "Optional.'Year',pay yearly;'Month',pay monthly;'Dynamic', pay hourly")
	req.Quantity = flags.Int("quantity", 1, "Optional. The duration of the instance. N years/months.")
	enableDataArk = flags.String("enable-data-ark", "false", "Optional. DataArk supports real-time backup, which can restore the disk back to any moment within the last 12 hours.")
	req.CouponId = flags.String("coupon-id", "", "Optional. Coupon ID, The Coupon can deduct part of the payment,see https://accountv2.ucloud.cn")

	flags.SetFlagValues("charge-type", "Month", "Year", "Dynamic", "Trial")
	flags.SetFlagValues("enable-data-ark", "true", "false")

	flags.SetFlagValuesFunc("source-id", func() []string {
		return getDiskList([]string{status.DISK_AVAILABLE}, *req.ProjectId, *req.Region, *req.Zone)
	})

	cmd.MarkFlagRequired("source-id")
	cmd.MarkFlagRequired("name")

	return cmd
}

//NewCmdDiskExpand ucloud disk expand
func NewCmdDiskExpand() *cobra.Command {
	req := base.BizClient.NewResizeUDiskRequest()
	cmd := &cobra.Command{
		Use:   "expand",
		Short: "Expand disk size",
		Long:  "Expand disk size",
		Run: func(cmd *cobra.Command, args []string) {
			if strings.Index(*req.UDiskId, "/") > -1 {
				*req.UDiskId = strings.SplitN(*req.UDiskId, "/", 2)[0]
			}
			if *req.Size > 8000 || *req.Size < 1 {
				base.Cxt.Println("size-gb should be between 1 and 8000")
			}
			_, err := base.BizClient.ResizeUDisk(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			base.Cxt.Printf("disk:[%s] expanded to %d GB\n", *req.UDiskId, *req.Size)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.UDiskId = flags.String("disk-id", "", "Required. Resource ID of the disk to expand")
	req.Size = flags.Int("size-gb", 0, "Required. Size of the disk after expanded. Unit: GB. Range [1,8000]")
	req.ProjectId = flags.String("project-id", base.ConfigInstance.ProjectID, "Optional. Assign project-id")
	req.Region = flags.String("region", base.ConfigInstance.Region, "Optional. Assign region")
	req.Zone = flags.String("zone", base.ConfigInstance.Zone, "Optional. Assign availability zone")
	req.CouponId = flags.String("coupon-id", "", "Optional. Coupon ID, The Coupon can deduct part of the payment,see https://accountv2.ucloud.cn")

	flags.SetFlagValuesFunc("disk-id", func() []string {
		return getDiskList([]string{status.DISK_AVAILABLE}, *req.ProjectId, *req.Region, *req.Zone)
	})

	cmd.MarkFlagRequired("disk-id")
	cmd.MarkFlagRequired("size-gb")

	return cmd
}

func getDiskList(states []string, project, region, zone string) []string {
	req := base.BizClient.NewDescribeUDiskRequest()
	req.ProjectId = sdk.String(project)
	req.Region = sdk.String(region)
	req.Zone = sdk.String(zone)
	req.Limit = sdk.Int(50)
	resp, err := base.BizClient.DescribeUDisk(req)
	if err != nil {
		//todo runtime log
		return nil
	}
	list := []string{}
	for _, disk := range resp.DataSet {
		for _, s := range states {
			if disk.Status == s {
				list = append(list, disk.UDiskId+"/"+strings.Replace(disk.Name, " ", "-", -1))
			}
		}
	}
	return list
}

func pollDisk(resourceID, projectID, region, zone, pollText string, targetState []string) {
	pollFunc := base.Poll(describeUdiskByID)
	done := pollFunc(resourceID, projectID, region, zone, targetState)
	ux.DotSpinner.Start(pollText)
	<-done
	ux.DotSpinner.Stop()
}

func describeUdiskByID(udiskID, project, region, zone string) (interface{}, error) {
	req := base.BizClient.NewDescribeUDiskRequest()
	req.UDiskId = sdk.String(udiskID)
	req.ProjectId = sdk.String(project)
	req.Region = sdk.String(region)
	req.Zone = sdk.String(zone)
	req.Limit = sdk.Int(50)
	resp, err := base.BizClient.DescribeUDisk(req)
	if err != nil {
		return nil, err
	}
	if len(resp.DataSet) < 1 {
		return nil, nil
	}
	return &resp.DataSet[0], nil
}
