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
	"strings"

	"github.com/spf13/cobra"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-sdk-go/private/services/uhost"
	"github.com/ucloud/ucloud-sdk-go/services/udisk"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/base"
	"github.com/ucloud/ucloud-cli/model/status"
	"github.com/ucloud/ucloud-cli/ux"
)

// NewCmdDisk ucloud disk
func NewCmdDisk() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "udisk",
		Short: "Read and manipulate udisk instances",
		Long:  "Read and manipulate udisk instances",
	}
	writer := base.Cxt.GetWriter()
	cmd.AddCommand(NewCmdDiskCreate(writer))
	cmd.AddCommand(NewCmdDiskList(writer))
	cmd.AddCommand(NewCmdDiskAttach(writer))
	cmd.AddCommand(NewCmdDiskDetach(writer))
	cmd.AddCommand(NewCmdDiskDelete())
	cmd.AddCommand(NewCmdDiskClone(writer))
	cmd.AddCommand(NewCmdDiskExpand())
	cmd.AddCommand(NewCmdDiskSnapshot(writer))
	cmd.AddCommand(NewCmdDiskRestore(writer))
	cmd.AddCommand(NewCmdSnapshotList(writer))
	cmd.AddCommand(NewCmdSnapshotDelete(writer))
	return cmd
}

// NewCmdDiskCreate ucloud udisk create
func NewCmdDiskCreate(out io.Writer) *cobra.Command {
	var async *bool
	var count *int
	var enableDataArk *string
	var snapshotID *string
	req := base.BizClient.NewCreateUDiskRequest()
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
			if *snapshotID != "" {
				cloneReq := base.BizClient.NewCloneUDiskSnapshotRequest()
				cloneReq.UDataArkMode = req.UDataArkMode
				cloneReq.SourceId = snapshotID
				cloneReq.ProjectId = req.ProjectId
				cloneReq.Region = req.Region
				cloneReq.Zone = req.Zone
				cloneReq.Name = req.Name
				cloneReq.Size = req.Size
				cloneReq.ChargeType = req.ChargeType
				cloneReq.Quantity = req.Quantity
				for i := 0; i < *count; i++ {
					resp, err := base.BizClient.CloneUDiskSnapshot(cloneReq)
					if err != nil {
						base.HandleError(err)
						return
					}
					if count := len(resp.UDiskId); count == 1 {
						text := fmt.Sprintf("udisk:%v is initializing", resp.UDiskId)
						if *async {
							fmt.Fprintln(out, text)
						} else {
							poller := base.NewSpoller(describeUdiskByID, out)
							poller.Spoll(resp.UDiskId[0], text, []string{status.DISK_AVAILABLE, status.DISK_FAILED})
						}
					} else if count > 1 {
						base.Cxt.Printf("udisk:%v created\n", resp.UDiskId)
					} else {
						base.Cxt.PrintErr(fmt.Errorf("none udisk created"))
					}
				}
			} else {
				for i := 0; i < *count; i++ {
					resp, err := base.BizClient.CreateUDisk(req)
					if err != nil {
						base.HandleError(err)
						return
					}
					if count := len(resp.UDiskId); count == 1 {
						text := fmt.Sprintf("udisk:%v is initializing", resp.UDiskId)
						if *async {
							fmt.Fprintln(out, text)
						} else {
							poller := base.NewSpoller(describeUdiskByID, out)
							poller.Spoll(resp.UDiskId[0], text, []string{status.DISK_AVAILABLE, status.DISK_FAILED})
						}
					} else if count > 1 {
						base.Cxt.Printf("udisk:%v created\n", resp.UDiskId)
					} else {
						base.Cxt.PrintErr(fmt.Errorf("none udisk created"))
					}
				}
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.Name = flags.String("name", "", "Required. Name of the udisk to create")
	req.Size = flags.Int("size-gb", 10, "Required. Size of the udisk to create. Unit:GB. Normal udisk [1,8000]; SSD udisk [1,4000] ")
	snapshotID = flags.String("snapshot-id", "", "Optional. Resource ID of a snapshot, which will apply to the udisk being created. If you set this option, 'udisk-type' will be omitted.")
	req.ProjectId = flags.String("project-id", base.ConfigIns.ProjectID, "Optional. Assign project-id")
	req.Region = flags.String("region", base.ConfigIns.Region, "Optional. Assign region")
	req.Zone = flags.String("zone", base.ConfigIns.Zone, "Optional. Assign availability zone")
	req.ChargeType = flags.String("charge-type", "Dynamic", "Optional.'Year',pay yearly;'Month',pay monthly;'Dynamic', pay hourly")
	req.Quantity = flags.Int("quantity", 1, "Optional. The duration of the instance. N years/months.")
	enableDataArk = flags.String("enable-data-ark", "false", "Optional. DataArk supports real-time backup, which can restore the udisk back to any moment within the last 12 hours.")
	req.Tag = flags.String("group", "Default", "Optional. Business group")
	req.DiskType = flags.String("udisk-type", "Oridinary", "Optional. 'Ordinary' or 'SSD'")
	async = flags.Bool("async", false, "Optional. Do not wait for the long-running operation to finish.")
	count = flags.Int("count", 1, "Optional. The count of udisk to create. Range [1,10]")

	flags.SetFlagValues("charge-type", "Month", "Year", "Dynamic", "Trial")
	flags.SetFlagValues("enable-data-ark", "true", "false")
	flags.SetFlagValues("udisk-type", "Oridinary", "SSD")

	cmd.MarkFlagRequired("size-gb")
	cmd.MarkFlagRequired("name")

	return cmd
}

// DiskRow TableRow
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

// NewCmdDiskList ucloud disk list
func NewCmdDiskList(out io.Writer) *cobra.Command {
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
			base.PrintList(list, out)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.ProjectId = flags.String("project-id", base.ConfigIns.ProjectID, "Optional. Assign project-id")
	req.Region = flags.String("region", base.ConfigIns.Region, "Optional. Assign region")
	req.Zone = flags.String("zone", base.ConfigIns.Zone, "Optional. Assign availability zone")
	req.UDiskId = flags.String("udisk-id", "", "Optional. Resource ID of the udisk to search")
	req.DiskType = flags.String("udisk-type", "", "Optional. Optional. Type of the udisk to search. 'Oridinary-Data-Disk','Oridinary-System-Disk' or 'SSD-Data-Disk'")
	req.Offset = cmd.Flags().Int("offset", 0, "Optional. Offset")
	req.Limit = cmd.Flags().Int("limit", 50, "Optional. Limit")
	flags.SetFlagValues("udisk-type", "Oridinary-Data-Disk", "Oridinary-System-Disk", "SSD-Data-Disk")
	return cmd
}

// NewCmdDiskAttach ucloud disk attach
func NewCmdDiskAttach(out io.Writer) *cobra.Command {
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
					fmt.Fprintln(out, text)
				} else {
					poller := base.NewSpoller(describeUdiskByID, out)
					poller.Spoll(resp.UDiskId, text, []string{status.DISK_INUSE, status.DISK_FAILED})
				}
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.UHostId = flags.String("uhost-id", "", "Required. Resource ID of the uhost instance which you want to attach the disk")
	udiskIDs = flags.StringSlice("udisk-id", nil, "Required. Resource ID of the udisk instances to attach")
	req.ProjectId = flags.String("project-id", base.ConfigIns.ProjectID, "Optional. Assign project-id")
	req.Region = flags.String("region", base.ConfigIns.Region, "Optional. Assign region")
	req.Zone = flags.String("zone", base.ConfigIns.Zone, "Optional. Assign availability zone")
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

// NewCmdDiskDetach ucloud udisk detach
func NewCmdDiskDetach(out io.Writer) *cobra.Command {
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
				err := detachUdisk(*async, id, out)
				if err != nil {
					base.Cxt.Println(err)
					continue
				}
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	udiskIDs = flags.StringSlice("udisk-id", nil, "Required. Resource ID of the udisk instances to detach")
	req.ProjectId = flags.String("project-id", base.ConfigIns.ProjectID, "Optional. Assign project-id")
	req.Region = flags.String("region", base.ConfigIns.Region, "Optional. Assign region")
	req.Zone = flags.String("zone", base.ConfigIns.Zone, "Optional. Assign availability zone")
	async = flags.BoolP("async", "a", false, "Optional. Do not wait for the long-running operation to finish.")
	yes = flags.BoolP("yes", "y", false, "Optional. Do not prompt for confirmation.")

	flags.SetFlagValuesFunc("udisk-id", func() []string {
		return getDiskList([]string{status.DISK_INUSE}, *req.ProjectId, *req.Region, *req.Zone)
	})

	cmd.MarkFlagRequired("udisk-id")
	return cmd
}

func detachUdisk(async bool, udiskID string, out io.Writer) error {
	any, err := describeUdiskByID(udiskID, nil)
	if err != nil {
		return err
	}
	if any == nil {
		return fmt.Errorf("udisk[%v] is not exist", any)
	}
	ins, ok := any.(*udisk.UDiskDataSet)
	if !ok {
		return fmt.Errorf("%#v convert to udisk failed", any)
	}
	req := base.BizClient.NewDetachUDiskRequest()
	req.UHostId = sdk.String(ins.UHostId)
	req.UDiskId = sdk.String(udiskID)
	resp, err := base.BizClient.DetachUDisk(req)
	if err != nil {
		return err
	}
	text := fmt.Sprintf("udisk[%s] is detaching from uhost[%s]", resp.UDiskId, resp.UHostId)
	if async {
		fmt.Fprintln(out, text)
	} else {
		poller := base.NewSpoller(describeUdiskByID, out)
		poller.Spoll(udiskID, text, []string{status.DISK_AVAILABLE, status.DISK_FAILED})
	}
	return nil
}

// NewCmdDiskDelete ucloud udisk delete
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
	req.ProjectId = flags.String("project-id", base.ConfigIns.ProjectID, "Optional. Assign project-id")
	req.Region = flags.String("region", base.ConfigIns.Region, "Optional. Assign region")
	req.Zone = flags.String("zone", base.ConfigIns.Zone, "Optional. Assign availability zone")
	yes = flags.BoolP("yes", "y", false, "Optional. Do not prompt for confirmation.")

	flags.SetFlagValuesFunc("udisk-id", func() []string {
		return getDiskList([]string{status.DISK_AVAILABLE, status.DISK_FAILED}, *req.ProjectId, *req.Region, *req.Zone)
	})

	cmd.MarkFlagRequired("udisk-id")

	return cmd
}

// NewCmdDiskClone ucloud disk clone
func NewCmdDiskClone(out io.Writer) *cobra.Command {
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
					fmt.Fprintln(out, text)
				} else {
					poller := base.NewSpoller(describeUdiskByID, out)
					poller.Spoll(resp.UDiskId[0], text, []string{status.DISK_AVAILABLE, status.DISK_FAILED})
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
	req.ProjectId = flags.String("project-id", base.ConfigIns.ProjectID, "Optional. Assign project-id")
	req.Region = flags.String("region", base.ConfigIns.Region, "Optional. Assign region")
	req.Zone = flags.String("zone", base.ConfigIns.Zone, "Optional. Assign availability zone")
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

// NewCmdDiskExpand ucloud udisk expand
func NewCmdDiskExpand() *cobra.Command {
	var udiskIDs *[]string
	req := base.BizClient.NewResizeUDiskRequest()
	cmd := &cobra.Command{
		Use:   "expand",
		Short: "Expand udisk size",
		Long:  "Expand udisk size",
		Run: func(cmd *cobra.Command, args []string) {
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
	req.ProjectId = flags.String("project-id", base.ConfigIns.ProjectID, "Optional. Assign project-id")
	req.Region = flags.String("region", base.ConfigIns.Region, "Optional. Assign region")
	req.Zone = flags.String("zone", base.ConfigIns.Zone, "Optional. Assign availability zone")

	flags.SetFlagValuesFunc("udisk-id", func() []string {
		return getDiskList([]string{status.DISK_AVAILABLE}, *req.ProjectId, *req.Region, *req.Zone)
	})

	cmd.MarkFlagRequired("udisk-id")
	cmd.MarkFlagRequired("size-gb")

	return cmd
}

// NewCmdDiskSnapshot ucloud udisk snapshot
func NewCmdDiskSnapshot(out io.Writer) *cobra.Command {
	var async *bool
	var udiskIDs *[]string
	req := base.BizClient.NewCreateUDiskSnapshotRequest()
	cmd := &cobra.Command{
		Use:   "snapshot",
		Short: "Create shapshots for udisks",
		Long:  "Create shapshots for udisks",
		Run: func(c *cobra.Command, args []string) {
			for _, id := range *udiskIDs {
				id = base.PickResourceID(id)
				req.UDiskId = &id
				resp, err := base.BizClient.CreateUDiskSnapshot(req)
				if err != nil {
					base.HandleError(err)
					return
				}
				if len(resp.SnapshotId) == 1 {
					text := fmt.Sprintf("snapshot[%s] is creating", resp.SnapshotId[0])
					if *async {
						fmt.Fprintln(out, text)
					} else {
						poller := base.NewSpoller(describeSnapshotByID, out)
						poller.Spoll(resp.SnapshotId[0], text, []string{status.SNAPSHOT_NORMAL})
					}
				} else {
					fmt.Fprintf(out, "snapshot%v is creating. expect snapshot count 1, accept %d\n", resp.SnapshotId, len(resp.SnapshotId))
				}
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	udiskIDs = flags.StringSlice("udisk-id", nil, "Required. Resource ID of udisks to snapshot")
	req.Name = flags.String("name", "", "Required. Name of snapshots")
	req.ProjectId = flags.String("project-id", base.ConfigIns.ProjectID, "Optional. Assign project-id")
	req.Region = flags.String("region", base.ConfigIns.Region, "Optional. Assign region")
	req.Zone = flags.String("zone", base.ConfigIns.Zone, "Optional. Assign availability zone")
	req.Comment = flags.String("comment", "", "Optional. Description of snapshots")
	async = flags.BoolP("async", "a", false, "Optional. Do not wait for the long-running operation to finish.")
	flags.SetFlagValuesFunc("udisk-id", func() []string {
		return getDiskList([]string{status.DISK_AVAILABLE, status.DISK_INUSE}, *req.ProjectId, *req.Region, *req.Zone)
	})
	cmd.MarkFlagRequired("udisk-id")
	cmd.MarkFlagRequired("name")
	return cmd
}

// NewCmdDiskRestore ucloud udisk restore
func NewCmdDiskRestore(out io.Writer) *cobra.Command {
	var snapshotIDs *[]string
	req := base.BizClient.NewRestoreUHostDiskRequest()
	cmd := &cobra.Command{
		Use:   "restore",
		Short: "Restore udisk from snapshot",
		Long:  "Restore udisk from snapshot",
		Run: func(cmd *cobra.Command, args []string) {
			for _, snapshotID := range *snapshotIDs {
				snapshotID = base.PickResourceID(snapshotID)
				any, err := describeSnapshotByID(snapshotID, nil)
				if err != nil {
					base.HandleError(err)
					continue
				}
				snapshot, ok := any.(*uhost.SnapshotSet)
				if !ok {
					fmt.Fprintf(out, "snapshot[%s] doesn't exist\n", snapshotID)
					continue
				}
				if snapshot.UHostId != "" {
					text := fmt.Sprintf("can we detach udisk[%s] from uhost[%s]?", snapshot.DiskId, snapshot.UHostId)
					sure, err := ux.Prompt(text)
					if err != nil {
						base.HandleError(err)
						continue
					}
					if !sure {
						continue
					}
					detachUdisk(false, snapshot.DiskId, out)
				}
				req.SnapshotIds = append(req.SnapshotIds, snapshotID)
				_, err = base.BizClient.RestoreUHostDisk(req)

				if err != nil {
					base.HandleError(err)
					return
				}

				text := fmt.Sprintf("udisk[%s] has been restored from snapshot[%s]", snapshot.DiskId, snapshot.SnapshotId)
				fmt.Fprintln(out, text)
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	snapshotIDs = flags.StringSlice("snapshot-id", nil, "Required. Resourece ID of the snapshots to restore from")
	req.ProjectId = flags.String("project-id", base.ConfigIns.ProjectID, "Optional. Assign project-id")
	req.Region = flags.String("region", base.ConfigIns.Region, "Optional. Assign region")
	req.Zone = flags.String("zone", base.ConfigIns.Zone, "Optional. Assign availability zone")
	flags.SetFlagValuesFunc("snapshot-id", func() []string {
		return getSnapshotList([]string{status.SNAPSHOT_NORMAL}, *req.ProjectId, *req.Region, *req.Zone)
	})
	cmd.MarkFlagRequired("snapshot-id")
	return cmd
}

// SnapshotRow 表格行
type SnapshotRow struct {
	Name             string
	ResourceID       string
	AvailabilityZone string
	BoundUDisk       string
	Size             string
	State            string
	UDiskType        string
	CreationTime     string
}

// NewCmdSnapshotList ucloud udisk list-snapshot
func NewCmdSnapshotList(out io.Writer) *cobra.Command {
	req := base.BizClient.NewDescribeSnapshotRequest()
	cmd := &cobra.Command{
		Use:   "list-snapshot",
		Short: "List snaphosts",
		Long:  "List snaphosts",
		Run: func(c *cobra.Command, args []string) {
			resp, err := base.BizClient.DescribeSnapshot(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			list := []SnapshotRow{}
			for _, snapshot := range resp.UHostSnapshotSet {
				row := SnapshotRow{
					Name:             snapshot.SnapshotName,
					ResourceID:       snapshot.SnapshotId,
					AvailabilityZone: snapshot.Zone,
					BoundUDisk:       snapshot.DiskId,
					Size:             fmt.Sprintf("%dGB", snapshot.Size),
					State:            snapshot.State,
					UDiskType:        snapshot.DiskType,
					CreationTime:     base.FormatDate(snapshot.CreateTime),
				}
				list = append(list, row)
			}
			base.PrintList(list, out)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.ProjectId = flags.String("project-id", base.ConfigIns.ProjectID, "Optional. Assign project-id")
	req.Region = flags.String("region", base.ConfigIns.Region, "Optional. Assign region")
	req.Zone = flags.String("zone", base.ConfigIns.Zone, "Optional. Assign availability zone")
	req.SnapshotIds = *flags.StringSlice("snaphost-id", nil, "Optional. Resource ID of snapshots to list")
	req.UHostId = flags.String("uhost-id", "", "Optional. Snapshots of the uhost")
	req.DiskId = flags.String("disk-id", "", "Optional. Snapshots of the udisk")
	req.Offset = cmd.Flags().Int("offset", 0, "Optional. Offset")
	req.Limit = cmd.Flags().Int("limit", 50, "Optional. Limit, length of snaphost list")

	return cmd
}

// NewCmdSnapshotDelete ucloud udisk delete-snapshot
func NewCmdSnapshotDelete(out io.Writer) *cobra.Command {
	var snapshotIds *[]string
	req := base.BizClient.NewDeleteSnapshotRequest()
	cmd := &cobra.Command{
		Use:   "delete-snapshot",
		Short: "Delete snapshots",
		Long:  "Delete snapshots",
		Run: func(c *cobra.Command, args []string) {
			for _, snapshotID := range *snapshotIds {
				req.SnapshotId = sdk.String(base.PickResourceID(snapshotID))
				resp, err := base.BizClient.DeleteSnapshot(req)
				if err != nil {
					base.HandleError(err)
					return
				}
				fmt.Fprintf(out, "snapshot[%s] deleted\n", resp.SnapshotId)
			}
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.ProjectId = flags.String("project-id", base.ConfigIns.ProjectID, "Optional. Assign project-id")
	req.Region = flags.String("region", base.ConfigIns.Region, "Optional. Assign region")
	req.Zone = flags.String("zone", base.ConfigIns.Zone, "Optional. Assign availability zone")
	snapshotIds = flags.StringSlice("snaphost-id", nil, "Optional. Resource ID of snapshots to delete")
	cmd.MarkFlagRequired("snapshot-id")
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

func describeUdiskByID(udiskID string, commonBase *request.CommonBase) (interface{}, error) {
	req := base.BizClient.NewDescribeUDiskRequest()
	if commonBase != nil {
		req.CommonBase = *commonBase
	}
	req.UDiskId = sdk.String(udiskID)
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

func getSnapshotList(states []string, project, region, zone string) []string {
	req := base.BizClient.NewDescribeUDiskSnapshotRequest()
	req.Limit = sdk.Int(50)
	req.ProjectId = &project
	req.Region = &region
	req.Zone = &zone
	resp, err := base.BizClient.DescribeUDiskSnapshot(req)
	if err != nil {
		return nil
	}
	list := []string{}
	for _, snapshot := range resp.DataSet {
		for _, s := range states {
			if snapshot.Status == s {
				list = append(list, snapshot.SnapshotId+"/"+strings.Replace(snapshot.Name, " ", "-", -1))
			}
		}
	}
	return list
}

func describeSnapshotByID(snapshotID string, commonBase *request.CommonBase) (interface{}, error) {
	req := base.BizClient.NewDescribeSnapshotRequest()
	if commonBase != nil {
		req.CommonBase = *commonBase
	}
	req.SnapshotIds = append(req.SnapshotIds, snapshotID)
	req.Limit = sdk.Int(50)
	resp, err := base.BizClient.DescribeSnapshot(req)
	if err != nil {
		return nil, err
	}
	if len(resp.UHostSnapshotSet) != 1 {
		return nil, nil
	}
	return &resp.UHostSnapshotSet[0], nil
}
