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

	"github.com/ucloud/ucloud-sdk-go/services/udisk"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/base"
	"github.com/ucloud/ucloud-cli/model/status"
	"github.com/ucloud/ucloud-cli/ux"
)

//NewCmdDisk ucloud disk
func NewCmdDisk() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "udisk",
		Short: "Read and manipulate udisk instances",
		Long:  "Read and manipulate udisk instances",
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

//NewCmdDiskCreate ucloud udisk create
func NewCmdDiskCreate() *cobra.Command {
	var async *bool
	var count *int
	req := base.BizClient.NewCreateUDiskRequest()
	enableDataArk := sdk.String("false")
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create udisk instance",
		Long:  "Create udisk instance",
		Run: func(cmd *cobra.Command, args []string) {
			if *count > 10 || *count < 1 {
				base.Cxt.Printf("Error, count should be between 1 and 10\n")
				return
			}
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
			for i := 0; i < *count; i++ {
				resp, err := base.BizClient.CreateUDisk(req)
				if err != nil {
					base.HandleError(err)
					return
				}
				if count := len(resp.UDiskId); count == 1 {
					text := fmt.Sprintf("udisk:%v is initializing", resp.UDiskId)
					if *async {
						base.Cxt.Println(text)
					} else {
						pollDisk(resp.UDiskId[0], *req.ProjectId, *req.Region, *req.Zone, text, []string{status.DISK_AVAILABLE, status.DISK_FAILED})
					}
				} else if count > 1 {
					base.Cxt.Printf("udisk:%v created\n", resp.UDiskId)
				} else {
					base.Cxt.PrintErr(fmt.Errorf("none udisk created"))
				}
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.Name = flags.String("name", "", "Required. Name of the udisk to create")
	req.Size = flags.Int("size-gb", 10, "Required. Size of the udisk to create. Unit:GB. Normal udisk [1,8000]; SSD udisk [1,4000] ")
	req.ProjectId = flags.String("project-id", base.ConfigInstance.ProjectID, "Optional. Assign project-id")
	req.Region = flags.String("region", base.ConfigInstance.Region, "Optional. Assign region")
	req.Zone = flags.String("zone", base.ConfigInstance.Zone, "Optional. Assign availability zone")
	req.ChargeType = flags.String("charge-type", "Dynamic", "Optional.'Year',pay yearly;'Month',pay monthly;'Dynamic', pay hourly")
	req.Quantity = flags.Int("quantity", 1, "Optional. The duration of the instance. N years/months.")
	enableDataArk = flags.String("enable-data-ark", "false", "Optional. DataArk supports real-time backup, which can restore the udisk back to any moment within the last 12 hours.")
	req.Tag = flags.String("group", "Default", "Optional. Business group")
	req.DiskType = flags.String("udisk-type", "Oridinary", "Optional. 'Ordinary' or 'SSD'")
	req.CouponId = flags.String("coupon-id", "", "Optional. Coupon ID, The Coupon can deduct part of the payment.See https://accountv2.ucloud.cn")
	async = flags.Bool("async", false, "Optional. Do not wait for the long-running operation to finish.")
	count = flags.Int("count", 1, "Optional. The count of udisk to create. Range [1,10]")

	flags.SetFlagValues("charge-type", "Month", "Year", "Dynamic", "Trial")
	flags.SetFlagValues("enable-data-ark", "true", "false")
	flags.SetFlagValues("udisk-type", "Oridinary", "SSD")

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
		Short: "List udisk instance",
		Long:  "List udisk instance",
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
					MountUHost:     fmt.Sprintf("%s/%s", disk.UHostName, disk.UHostIP),
					MountPoint:     disk.DeviceName,
					State:          disk.Status,
					CreationTime:   base.FormatDate(disk.CreateTime),
					ExpirationTime: base.FormatDate(disk.ExpiredTime),
				}
				if disk.UHostIP == "" {
					row.MountUHost = ""
				}
				list = append(list, row)
			}
			if global.json {
				base.PrintJSON(list)
			} else {
				base.PrintTableS(list)
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.ProjectId = flags.String("project-id", base.ConfigInstance.ProjectID, "Optional. Assign project-id")
	req.Region = flags.String("region", base.ConfigInstance.Region, "Optional. Assign region")
	req.Zone = flags.String("zone", base.ConfigInstance.Zone, "Optional. Assign availability zone")
	req.UDiskId = flags.String("resource-id", "", "Optional. Resource ID of the udisk to search")
	req.DiskType = flags.String("udisk-type", "", "Optional. Optional. Type of the udisk to search. 'Oridinary-Data-Disk','Oridinary-System-Disk' or 'SSD-Data-Disk'")
	req.Offset = cmd.Flags().Int("offset", 0, "Optional. Offset")
	req.Limit = cmd.Flags().Int("limit", 50, "Optional. Limit")
	flags.SetFlagValues("udisk-type", "Oridinary-Data-Disk", "Oridinary-System-Disk", "SSD-Data-Disk")
	return cmd
}

//NewCmdDiskAttach ucloud disk attach
func NewCmdDiskAttach() *cobra.Command {
	var async *bool
	var udiskIDs *[]string

	req := base.BizClient.NewAttachUDiskRequest()
	cmd := &cobra.Command{
		Use:     "attach",
		Short:   "Attach udisk instances to an uhost",
		Long:    "Attach udisk instances to an uhost",
		Example: "ucloud udisk attach --uhost-id uhost-xxxx --udisk-id bs-xxx1,bs-xxx2",
		Run: func(cmd *cobra.Command, args []string) {
			for _, id := range *udiskIDs {
				id = base.PickResourceID(id)
				req.UDiskId = &id
				*req.UHostId = base.PickResourceID(*req.UHostId)
				resp, err := base.BizClient.AttachUDisk(req)
				if err != nil {
					base.HandleError(err)
					return
				}
				text := fmt.Sprintf("udisk[%s] is attaching to uhost uhost[%s]", *req.UDiskId, *req.UHostId)
				if *async {
					base.Cxt.Println(text)
				} else {
					pollDisk(resp.UDiskId, *req.ProjectId, *req.Region, *req.Zone, text, []string{status.DISK_INUSE, status.DISK_FAILED})
				}
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.UHostId = flags.String("uhost-id", "", "Required. Resource ID of the uhost instance which you want to attach the disk")
	udiskIDs = flags.StringSlice("udisk-id", nil, "Required. Resource ID of the udisk instances to attach")
	req.ProjectId = flags.String("project-id", base.ConfigInstance.ProjectID, "Optional. Assign project-id")
	req.Region = flags.String("region", base.ConfigInstance.Region, "Optional. Assign region")
	req.Zone = flags.String("zone", base.ConfigInstance.Zone, "Optional. Assign availability zone")
	async = flags.Bool("async", false, "Optional. Do not wait for the long-running operation to finish.")

	flags.SetFlagValuesFunc("udisk-id", func() []string {
		return getDiskList([]string{status.DISK_AVAILABLE}, *req.ProjectId, *req.Region, *req.Zone)
	})
	flags.SetFlagValuesFunc("uhost-id", func() []string {
		return getUhostList([]string{status.HOST_RUNNING, status.HOST_STOPPED}, *req.ProjectId, *req.Region, *req.Zone)
	})

	cmd.MarkFlagRequired("uhost-id")
	cmd.MarkFlagRequired("udisk-id")

	return cmd
}

//NewCmdDiskDetach ucloud udisk detach
func NewCmdDiskDetach() *cobra.Command {
	var async, yes *bool
	var udiskIDs *[]string
	req := base.BizClient.NewDetachUDiskRequest()
	cmd := &cobra.Command{
		Use:   "detach",
		Short: "Detach udisk instances from an uhost",
		Long:  "Detach udisk instances from an uhost",
		Run: func(cmd *cobra.Command, args []string) {
			text := `Please confirm that you have already unmounted file system corresponding to this hard drive,(See "https://docs.ucloud.cn/storage_cdn/udisk/userguide/umount" for help), otherwise it will cause file system damage and UHost cannot be normally shut down. Sure to detach?`
			if !*yes {
				sure, err := ux.Prompt(text)
				if err != nil {
					base.Cxt.PrintErr(err)
					return
				}
				if !sure {
					return
				}
			}
			for _, id := range *udiskIDs {
				id = base.PickResourceID(id)
				any, err := describeUdiskByID(id, *req.ProjectId, *req.Region, *req.Zone)
				if err != nil {
					base.HandleError(err)
					continue
				}
				if any == nil {
					base.Cxt.PrintErr(fmt.Errorf("udisk[%v] is not exist", any))
					continue
				}
				ins, ok := any.(*udisk.UDiskDataSet)
				if !ok {
					base.Cxt.PrintErr(fmt.Errorf("%#v convert to udisk failed", any))
					continue
				}
				req.UHostId = &ins.UHostId
				req.UDiskId = &id
				*req.UHostId = base.PickResourceID(*req.UHostId)
				resp, err := base.BizClient.DetachUDisk(req)
				if err != nil {
					base.HandleError(err)
					continue
				}
				text := fmt.Sprintf("udisk[%s] is detaching from uhost[%s]", resp.UDiskId, resp.UHostId)
				if *async {
					base.Cxt.Println(text)
				} else {
					pollDisk(resp.UDiskId, *req.ProjectId, *req.Region, *req.Zone, text, []string{status.DISK_AVAILABLE, status.DISK_FAILED})
				}
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	udiskIDs = flags.StringSlice("udisk-id", nil, "Required. Resource ID of the udisk instances to detach")
	req.ProjectId = flags.String("project-id", base.ConfigInstance.ProjectID, "Optional. Assign project-id")
	req.Region = flags.String("region", base.ConfigInstance.Region, "Optional. Assign region")
	req.Zone = flags.String("zone", base.ConfigInstance.Zone, "Optional. Assign availability zone")
	async = flags.Bool("async", false, "Optional. Do not wait for the long-running operation to finish.")
	yes = flags.BoolP("yes", "y", false, "Optional. Do not prompt for confirmation.")

	flags.SetFlagValuesFunc("udisk-id", func() []string {
		return getDiskList([]string{status.DISK_INUSE}, *req.ProjectId, *req.Region, *req.Zone)
	})

	cmd.MarkFlagRequired("udisk-id")
	return cmd
}

//NewCmdDiskDelete ucloud udisk delete
func NewCmdDiskDelete() *cobra.Command {
	var yes *bool
	var udiskIDs *[]string
	req := base.BizClient.NewDeleteUDiskRequest()
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete udisk instances",
		Long:  "Delete udisk instances",
		Run: func(cmd *cobra.Command, args []string) {
			if !*yes {
				sure, err := ux.Prompt(fmt.Sprintf("Are you sure to delete udisk(s)?"))
				if err != nil {
					base.Cxt.PrintErr(err)
					return
				}
				if !sure {
					return
				}
			}
			for _, id := range *udiskIDs {
				id := base.PickResourceID(id)
				req.UDiskId = &id
				_, err := base.BizClient.DeleteUDisk(req)
				if err != nil {
					base.HandleError(err)
					continue
				} else {
					base.Cxt.Printf("udisk[%s] deleted\n", *req.UDiskId)
				}
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	udiskIDs = flags.StringSlice("udisk-id", nil, "Required. The Resource ID of udisks to delete")
	req.ProjectId = flags.String("project-id", base.ConfigInstance.ProjectID, "Optional. Assign project-id")
	req.Region = flags.String("region", base.ConfigInstance.Region, "Optional. Assign region")
	req.Zone = flags.String("zone", base.ConfigInstance.Zone, "Optional. Assign availability zone")
	yes = flags.BoolP("yes", "y", false, "Optional. Do not prompt for confirmation.")

	flags.SetFlagValuesFunc("udisk-id", func() []string {
		return getDiskList([]string{status.DISK_AVAILABLE, status.DISK_FAILED}, *req.ProjectId, *req.Region, *req.Zone)
	})

	cmd.MarkFlagRequired("udisk-id")

	return cmd
}

//NewCmdDiskClone ucloud disk clone
func NewCmdDiskClone() *cobra.Command {
	var async *bool
	req := base.BizClient.NewCloneUDiskRequest()
	enableDataArk := sdk.String("false")
	cmd := &cobra.Command{
		Use:   "clone",
		Short: "Clone an udisk",
		Long:  "Clone an udisk",
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
				text := fmt.Sprintf("cloned udisk:[%s] is initializing", resp.UDiskId[0])
				if *async {
					base.Cxt.Println(text)
				} else {
					pollDisk(resp.UDiskId[0], *req.ProjectId, *req.Region, *req.Zone, text, []string{status.DISK_AVAILABLE, status.DISK_FAILED})
				}
			} else {
				base.Cxt.Printf("udisk[%v] cloned", resp.UDiskId)
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.SourceId = flags.String("source-id", "", "Required. Resource ID of parent udisk")
	req.Name = flags.String("name", "", "Required. Name of new udisk")
	req.ProjectId = flags.String("project-id", base.ConfigInstance.ProjectID, "Optional. Assign project-id")
	req.Region = flags.String("region", base.ConfigInstance.Region, "Optional. Assign region")
	req.Zone = flags.String("zone", base.ConfigInstance.Zone, "Optional. Assign availability zone")
	req.ChargeType = flags.String("charge-type", "Month", "Optional.'Year',pay yearly;'Month',pay monthly;'Dynamic', pay hourly")
	req.Quantity = flags.Int("quantity", 1, "Optional. The duration of the instance. N years/months.")
	enableDataArk = flags.String("enable-data-ark", "false", "Optional. DataArk supports real-time backup, which can restore the udisk back to any moment within the last 12 hours.")
	req.CouponId = flags.String("coupon-id", "", "Optional. Coupon ID, The Coupon can deduct part of the payment,see https://accountv2.ucloud.cn")
	async = flags.Bool("async", false, "Optional. Do not wait for the long-running operation to finish.")

	flags.SetFlagValues("charge-type", "Month", "Year", "Dynamic", "Trial")
	flags.SetFlagValues("enable-data-ark", "true", "false")

	flags.SetFlagValuesFunc("source-id", func() []string {
		return getDiskList([]string{status.DISK_AVAILABLE}, *req.ProjectId, *req.Region, *req.Zone)
	})

	cmd.MarkFlagRequired("source-id")
	cmd.MarkFlagRequired("name")

	return cmd
}

//NewCmdDiskExpand ucloud udisk expand
func NewCmdDiskExpand() *cobra.Command {
	var udiskIDs *[]string
	req := base.BizClient.NewResizeUDiskRequest()
	cmd := &cobra.Command{
		Use:   "expand",
		Short: "Expand udisk size",
		Long:  "Expand udisk size",
		Run: func(cmd *cobra.Command, args []string) {
			if *req.Size > 8000 || *req.Size < 1 {
				base.Cxt.Println("size-gb should be between 1 and 8000")
				return
			}
			for _, id := range *udiskIDs {
				id = base.PickResourceID(id)
				req.UDiskId = &id
				_, err := base.BizClient.ResizeUDisk(req)
				if err != nil {
					base.HandleError(err)
					return
				}
				base.Cxt.Printf("udisk:[%s] expanded to %d GB\n", *req.UDiskId, *req.Size)
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	udiskIDs = flags.StringSlice("udisk-id", nil, "Required. Resource ID of the udisks to expand")
	req.Size = flags.Int("size-gb", 0, "Required. Size of the udisk after expanded. Unit: GB. Range [1,8000]")
	req.ProjectId = flags.String("project-id", base.ConfigInstance.ProjectID, "Optional. Assign project-id")
	req.Region = flags.String("region", base.ConfigInstance.Region, "Optional. Assign region")
	req.Zone = flags.String("zone", base.ConfigInstance.Zone, "Optional. Assign availability zone")
	req.CouponId = flags.String("coupon-id", "", "Optional. Coupon ID, The Coupon can deduct part of the payment,see https://accountv2.ucloud.cn")

	flags.SetFlagValuesFunc("udisk-id", func() []string {
		return getDiskList([]string{status.DISK_AVAILABLE}, *req.ProjectId, *req.Region, *req.Zone)
	})

	cmd.MarkFlagRequired("udisk-id")
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
