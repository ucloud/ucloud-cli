package udisk

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	puhost "github.com/ucloud/ucloud-sdk-go/private/services/uhost"
	udisksdk "github.com/ucloud/ucloud-sdk-go/services/udisk"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/internal/common"
	"github.com/ucloud/ucloud-cli/model/status"
	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// NewCommand builds the `udisk` root command and mounts the 11 subcommands.
// Mirrors cmd/disk.go NewCmdDisk (same AddCommand order).
func NewCommand(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "udisk",
		Short: "Read and manipulate udisk instances",
		Long:  "Read and manipulate udisk instances",
	}
	cmd.AddCommand(newCreate(ctx))
	cmd.AddCommand(newList(ctx))
	cmd.AddCommand(newAttach(ctx))
	cmd.AddCommand(newDetach(ctx))
	cmd.AddCommand(newDelete(ctx))
	cmd.AddCommand(newClone(ctx))
	cmd.AddCommand(newExpand(ctx))
	cmd.AddCommand(newSnapshot(ctx))
	cmd.AddCommand(newRestore(ctx))
	cmd.AddCommand(newSnapshotList(ctx))
	cmd.AddCommand(newSnapshotDelete(ctx))
	return cmd
}

// newCreate ucloud udisk create
func newCreate(ctx *cli.Context) *cobra.Command {
	var async *bool
	var count *int
	var enableDataArk *string
	var snapshotID *string
	client := cli.NewServiceClient(ctx, udisksdk.NewClient)
	req := client.NewCreateUDiskRequest()
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create udisk instance",
		Long:  "Create udisk instance",
		Run: func(cmd *cobra.Command, args []string) {
			w := ctx.ProgressWriter()
			results := []cli.OpResultRow{}
			if *count > 10 || *count < 1 {
				fmt.Fprintf(w, "Error, count should be between 1 and 10\n")
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
				cloneReq := client.NewCloneUDiskSnapshotRequest()
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
					resp, err := client.CloneUDiskSnapshot(cloneReq)
					if err != nil {
						ctx.HandleError(err)
						return
					}
					if count := len(resp.UDiskId); count == 1 {
						text := fmt.Sprintf("udisk:%v is initializing", resp.UDiskId)
						if *async {
							fmt.Fprintln(w, text)
						} else {
							ctx.PollerTo(w, describeUdiskByID(ctx)).Spoll(resp.UDiskId[0], text, []string{status.DISK_AVAILABLE, status.DISK_FAILED})
						}
						results = append(results, cli.OpResultRow{ResourceID: resp.UDiskId[0], Action: "create", Status: "Initializing"})
					} else if count > 1 {
						fmt.Fprintf(w, "udisk:%v created\n", resp.UDiskId)
						for _, id := range resp.UDiskId {
							results = append(results, cli.OpResultRow{ResourceID: id, Action: "create", Status: "Created"})
						}
					} else {
						fmt.Fprintln(ctx.Err(), "Error: none udisk created")
					}
				}
			} else {
				for i := 0; i < *count; i++ {
					resp, err := client.CreateUDisk(req)
					if err != nil {
						ctx.HandleError(err)
						return
					}
					if count := len(resp.UDiskId); count == 1 {
						text := fmt.Sprintf("udisk:%v is initializing", resp.UDiskId)
						if *async {
							fmt.Fprintln(w, text)
						} else {
							ctx.PollerTo(w, describeUdiskByID(ctx)).Spoll(resp.UDiskId[0], text, []string{status.DISK_AVAILABLE, status.DISK_FAILED})
						}
						results = append(results, cli.OpResultRow{ResourceID: resp.UDiskId[0], Action: "create", Status: "Initializing"})
					} else if count > 1 {
						fmt.Fprintf(w, "udisk:%v created\n", resp.UDiskId)
						for _, id := range resp.UDiskId {
							results = append(results, cli.OpResultRow{ResourceID: id, Action: "create", Status: "Created"})
						}
					} else {
						fmt.Fprintln(ctx.Err(), "Error: none udisk created")
					}
				}
			}
			ctx.EmitResult(results...)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.Name = flags.String("name", "", "Required. Name of the udisk to create")
	req.Size = flags.Int("size-gb", 10, "Required. Size of the udisk to create. Unit:GB. Normal udisk [1,8000]; SSD udisk [1,4000] ")
	snapshotID = flags.String("snapshot-id", "", "Optional. Resource ID of a snapshot, which will apply to the udisk being created. If you set this option, 'udisk-type' will be omitted.")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	req.Zone = flags.String("zone", ctx.DefaultZone(), "Optional. Assign availability zone")
	req.ChargeType = flags.String("charge-type", "Dynamic", "Optional.'Year',pay yearly;'Month',pay monthly;'Dynamic', pay hourly")
	req.Quantity = flags.Int("quantity", 1, "Optional. The duration of the instance. N years/months.")
	enableDataArk = flags.String("enable-data-ark", "false", "Optional. DataArk supports real-time backup, which can restore the udisk back to any moment within the last 12 hours.")
	req.Tag = flags.String("group", "Default", "Optional. Business group")
	req.DiskType = flags.String("udisk-type", "Oridinary", "Optional. 'Ordinary' or 'SSD'")
	async = flags.Bool("async", false, "Optional. Do not wait for the long-running operation to finish.")
	count = flags.Int("count", 1, "Optional. The count of udisk to create. Range [1,10]")

	command.SetFlagValues(cmd, "charge-type", "Month", "Year", "Dynamic", "Trial")
	command.SetFlagValues(cmd, "enable-data-ark", "true", "false")
	command.SetFlagValues(cmd, "udisk-type", "Oridinary", "SSD")

	cmd.MarkFlagRequired("size-gb")
	cmd.MarkFlagRequired("name")

	return cmd
}

// newList ucloud udisk list
func newList(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, udisksdk.NewClient)
	req := client.NewDescribeUDiskRequest()
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
			resp, err := client.DescribeUDisk(req)
			if err != nil {
				ctx.HandleError(err)
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
					CreationTime:   common.FormatDate(disk.CreateTime),
					ExpirationTime: common.FormatDate(disk.ExpiredTime),
				}
				if disk.UHostIP == "" {
					row.MountUHost = ""
				}
				list = append(list, row)
			}
			ctx.PrintList(list)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	req.Zone = flags.String("zone", ctx.DefaultZone(), "Optional. Assign availability zone")
	req.UDiskId = flags.String("udisk-id", "", "Optional. Resource ID of the udisk to search")
	req.DiskType = flags.String("udisk-type", "", "Optional. Optional. Type of the udisk to search. 'Oridinary-Data-Disk','Oridinary-System-Disk' or 'SSD-Data-Disk'")
	req.Offset = cmd.Flags().Int("offset", 0, "Optional. Offset")
	req.Limit = cmd.Flags().Int("limit", 50, "Optional. Limit")
	command.SetFlagValues(cmd, "udisk-type", "Oridinary-Data-Disk", "Oridinary-System-Disk", "SSD-Data-Disk")
	return cmd
}

// newAttach ucloud udisk attach
func newAttach(ctx *cli.Context) *cobra.Command {
	var async *bool
	var udiskIDs *[]string

	client := cli.NewServiceClient(ctx, udisksdk.NewClient)
	req := client.NewAttachUDiskRequest()
	cmd := &cobra.Command{
		Use:     "attach",
		Short:   "Attach udisk instances to an uhost",
		Long:    "Attach udisk instances to an uhost",
		Example: "ucloud udisk attach --uhost-id uhost-xxxx --udisk-id bs-xxx1,bs-xxx2",
		Run: func(cmd *cobra.Command, args []string) {
			w := ctx.ProgressWriter()
			results := []cli.OpResultRow{}
			for _, id := range *udiskIDs {
				id = ctx.PickResourceID(id)
				req.UDiskId = &id
				*req.UHostId = ctx.PickResourceID(*req.UHostId)
				resp, err := client.AttachUDisk(req)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				text := fmt.Sprintf("udisk[%s] is attaching to uhost uhost[%s]", *req.UDiskId, *req.UHostId)
				if *async {
					fmt.Fprintln(w, text)
				} else {
					ctx.PollerTo(w, describeUdiskByID(ctx)).Spoll(resp.UDiskId, text, []string{status.DISK_INUSE, status.DISK_FAILED})
				}
				results = append(results, cli.OpResultRow{ResourceID: resp.UDiskId, Action: "attach", Status: "Attaching"})
			}
			ctx.EmitResult(results...)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.UHostId = flags.String("uhost-id", "", "Required. Resource ID of the uhost instance which you want to attach the disk")
	udiskIDs = flags.StringSlice("udisk-id", nil, "Required. Resource ID of the udisk instances to attach")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	req.Zone = flags.String("zone", ctx.DefaultZone(), "Optional. Assign availability zone")
	async = flags.Bool("async", false, "Optional. Do not wait for the long-running operation to finish.")

	command.SetCompletion(cmd, "udisk-id", func() []string {
		return getDiskList(ctx, []string{status.DISK_AVAILABLE}, *req.ProjectId, *req.Region, *req.Zone)
	})
	command.SetCompletion(cmd, "uhost-id", func() []string {
		return getUhostList(ctx, []string{status.HOST_RUNNING, status.HOST_STOPPED}, *req.ProjectId, *req.Region, *req.Zone)
	})

	cmd.MarkFlagRequired("uhost-id")
	cmd.MarkFlagRequired("udisk-id")

	return cmd
}

// newDetach ucloud udisk detach
func newDetach(ctx *cli.Context) *cobra.Command {
	var async, yes *bool
	var udiskIDs *[]string
	client := cli.NewServiceClient(ctx, udisksdk.NewClient)
	req := client.NewDetachUDiskRequest()
	cmd := &cobra.Command{
		Use:   "detach",
		Short: "Detach udisk instances from an uhost",
		Long:  "Detach udisk instances from an uhost",
		Run: func(cmd *cobra.Command, args []string) {
			text := `Please confirm that you have already unmounted file system corresponding to this hard drive,(See "https://docs.ucloud.cn/storage_cdn/udisk/userguide/umount" for help), otherwise it will cause file system damage and UHost cannot be normally shut down. Sure to detach?`
			if !ctx.Confirm(*yes, text) {
				return
			}
			w := ctx.ProgressWriter()
			results := []cli.OpResultRow{}
			for _, id := range *udiskIDs {
				id = ctx.PickResourceID(id)
				err := DetachUdisk(ctx, *async, id, w)
				if err != nil {
					fmt.Fprintln(ctx.Err(), err)
					continue
				}
				results = append(results, cli.OpResultRow{ResourceID: id, Action: "detach", Status: "Detaching"})
			}
			ctx.EmitResult(results...)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	udiskIDs = flags.StringSlice("udisk-id", nil, "Required. Resource ID of the udisk instances to detach")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	req.Zone = flags.String("zone", ctx.DefaultZone(), "Optional. Assign availability zone")
	async = flags.BoolP("async", "a", false, "Optional. Do not wait for the long-running operation to finish.")
	yes = flags.BoolP("yes", "y", false, "Optional. Do not prompt for confirmation.")

	command.SetCompletion(cmd, "udisk-id", func() []string {
		return getDiskList(ctx, []string{status.DISK_INUSE}, *req.ProjectId, *req.Region, *req.Zone)
	})

	cmd.MarkFlagRequired("udisk-id")
	return cmd
}

// newDelete ucloud udisk delete
func newDelete(ctx *cli.Context) *cobra.Command {
	var yes *bool
	var udiskIDs *[]string
	client := cli.NewServiceClient(ctx, udisksdk.NewClient)
	req := client.NewDeleteUDiskRequest()
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete udisk instances",
		Long:  "Delete udisk instances",
		Run: func(cmd *cobra.Command, args []string) {
			if !ctx.Confirm(*yes, fmt.Sprintf("Are you sure to delete udisk(s)?")) {
				return
			}
			w := ctx.ProgressWriter()
			results := []cli.OpResultRow{}
			for _, id := range *udiskIDs {
				id := ctx.PickResourceID(id)
				req.UDiskId = &id
				_, err := client.DeleteUDisk(req)
				if err != nil {
					ctx.HandleError(err)
					continue
				} else {
					fmt.Fprintf(w, "udisk[%s] deleted\n", *req.UDiskId)
					results = append(results, cli.OpResultRow{ResourceID: *req.UDiskId, Action: "delete", Status: "Deleted"})
				}
			}
			ctx.EmitResult(results...)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	udiskIDs = flags.StringSlice("udisk-id", nil, "Required. The Resource ID of udisks to delete")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	req.Zone = flags.String("zone", ctx.DefaultZone(), "Optional. Assign availability zone")
	yes = flags.BoolP("yes", "y", false, "Optional. Do not prompt for confirmation.")

	command.SetCompletion(cmd, "udisk-id", func() []string {
		return getDiskList(ctx, []string{status.DISK_AVAILABLE, status.DISK_FAILED}, *req.ProjectId, *req.Region, *req.Zone)
	})

	cmd.MarkFlagRequired("udisk-id")

	return cmd
}

// newClone ucloud udisk clone
func newClone(ctx *cli.Context) *cobra.Command {
	var async *bool
	client := cli.NewServiceClient(ctx, udisksdk.NewClient)
	req := client.NewCloneUDiskRequest()
	enableDataArk := sdk.String("false")
	cmd := &cobra.Command{
		Use:   "clone",
		Short: "Clone an udisk",
		Long:  "Clone an udisk",
		Run: func(cmd *cobra.Command, args []string) {
			w := ctx.ProgressWriter()
			if *enableDataArk == "true" {
				req.UDataArkMode = sdk.String("Yes")
			} else {
				req.UDataArkMode = sdk.String("No")
			}
			if strings.Index(*req.SourceId, "/") > -1 {
				*req.SourceId = strings.SplitN(*req.SourceId, "/", 2)[0]
			}
			resp, err := client.CloneUDisk(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if len(resp.UDiskId) == 1 {
				text := fmt.Sprintf("cloned udisk:[%s] is initializing", resp.UDiskId[0])
				if *async {
					fmt.Fprintln(w, text)
				} else {
					ctx.PollerTo(w, describeUdiskByID(ctx)).Spoll(resp.UDiskId[0], text, []string{status.DISK_AVAILABLE, status.DISK_FAILED})
				}
				ctx.EmitResult(cli.OpResultRow{ResourceID: resp.UDiskId[0], Action: "clone", Status: "Initializing"})
			} else {
				fmt.Fprintf(w, "udisk[%v] cloned", resp.UDiskId)
				results := []cli.OpResultRow{}
				for _, id := range resp.UDiskId {
					results = append(results, cli.OpResultRow{ResourceID: id, Action: "clone", Status: "Cloned"})
				}
				ctx.EmitResult(results...)
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.SourceId = flags.String("source-id", "", "Required. Resource ID of parent udisk")
	req.Name = flags.String("name", "", "Required. Name of new udisk")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	req.Zone = flags.String("zone", ctx.DefaultZone(), "Optional. Assign availability zone")
	req.ChargeType = flags.String("charge-type", "Month", "Optional.'Year',pay yearly;'Month',pay monthly;'Dynamic', pay hourly")
	req.Quantity = flags.Int("quantity", 1, "Optional. The duration of the instance. N years/months.")
	enableDataArk = flags.String("enable-data-ark", "false", "Optional. DataArk supports real-time backup, which can restore the udisk back to any moment within the last 12 hours.")
	req.CouponId = flags.String("coupon-id", "", "Optional. Coupon ID, The Coupon can deduct part of the payment,see https://accountv2.ucloud.cn")
	async = flags.Bool("async", false, "Optional. Do not wait for the long-running operation to finish.")

	command.SetFlagValues(cmd, "charge-type", "Month", "Year", "Dynamic", "Trial")
	command.SetFlagValues(cmd, "enable-data-ark", "true", "false")

	command.SetCompletion(cmd, "source-id", func() []string {
		return getDiskList(ctx, []string{status.DISK_AVAILABLE}, *req.ProjectId, *req.Region, *req.Zone)
	})

	cmd.MarkFlagRequired("source-id")
	cmd.MarkFlagRequired("name")

	return cmd
}

// newExpand ucloud udisk expand
func newExpand(ctx *cli.Context) *cobra.Command {
	var udiskIDs *[]string
	client := cli.NewServiceClient(ctx, udisksdk.NewClient)
	req := client.NewResizeUDiskRequest()
	cmd := &cobra.Command{
		Use:   "expand",
		Short: "Expand udisk size",
		Long:  "Expand udisk size",
		Run: func(cmd *cobra.Command, args []string) {
			w := ctx.ProgressWriter()
			results := []cli.OpResultRow{}
			for _, id := range *udiskIDs {
				id = ctx.PickResourceID(id)
				req.UDiskId = &id
				_, err := client.ResizeUDisk(req)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				fmt.Fprintf(w, "udisk:[%s] expanded to %d GB\n", *req.UDiskId, *req.Size)
				results = append(results, cli.OpResultRow{ResourceID: *req.UDiskId, Action: "expand", Status: "Expanded"})
			}
			ctx.EmitResult(results...)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	udiskIDs = flags.StringSlice("udisk-id", nil, "Required. Resource ID of the udisks to expand")
	req.Size = flags.Int("size-gb", 0, "Required. Size of the udisk after expanded. Unit: GB. Range [1,8000]")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	req.Zone = flags.String("zone", ctx.DefaultZone(), "Optional. Assign availability zone")

	command.SetCompletion(cmd, "udisk-id", func() []string {
		return getDiskList(ctx, []string{status.DISK_AVAILABLE}, *req.ProjectId, *req.Region, *req.Zone)
	})

	cmd.MarkFlagRequired("udisk-id")
	cmd.MarkFlagRequired("size-gb")

	return cmd
}

// newSnapshot ucloud udisk snapshot
func newSnapshot(ctx *cli.Context) *cobra.Command {
	var async *bool
	var udiskIDs *[]string
	client := cli.NewServiceClient(ctx, udisksdk.NewClient)
	req := client.NewCreateUDiskSnapshotRequest()
	cmd := &cobra.Command{
		Use:   "snapshot",
		Short: "Create shapshots for udisks",
		Long:  "Create shapshots for udisks",
		Run: func(c *cobra.Command, args []string) {
			w := ctx.ProgressWriter()
			results := []cli.OpResultRow{}
			for _, id := range *udiskIDs {
				id = ctx.PickResourceID(id)
				req.UDiskId = &id
				resp, err := client.CreateUDiskSnapshot(req)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				if len(resp.SnapshotId) == 1 {
					text := fmt.Sprintf("snapshot[%s] is creating", resp.SnapshotId[0])
					if *async {
						fmt.Fprintln(w, text)
					} else {
						ctx.PollerTo(w, describeSnapshotByID(ctx)).Spoll(resp.SnapshotId[0], text, []string{status.SNAPSHOT_NORMAL})
					}
					results = append(results, cli.OpResultRow{ResourceID: resp.SnapshotId[0], Action: "snapshot", Status: "Creating"})
				} else {
					fmt.Fprintf(w, "snapshot%v is creating. expect snapshot count 1, accept %d\n", resp.SnapshotId, len(resp.SnapshotId))
					for _, sid := range resp.SnapshotId {
						results = append(results, cli.OpResultRow{ResourceID: sid, Action: "snapshot", Status: "Creating"})
					}
				}
			}
			ctx.EmitResult(results...)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	udiskIDs = flags.StringSlice("udisk-id", nil, "Required. Resource ID of udisks to snapshot")
	req.Name = flags.String("name", "", "Required. Name of snapshots")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	req.Zone = flags.String("zone", ctx.DefaultZone(), "Optional. Assign availability zone")
	req.Comment = flags.String("comment", "", "Optional. Description of snapshots")
	async = flags.BoolP("async", "a", false, "Optional. Do not wait for the long-running operation to finish.")
	command.SetCompletion(cmd, "udisk-id", func() []string {
		return getDiskList(ctx, []string{status.DISK_AVAILABLE, status.DISK_INUSE}, *req.ProjectId, *req.Region, *req.Zone)
	})
	cmd.MarkFlagRequired("udisk-id")
	cmd.MarkFlagRequired("name")
	return cmd
}

// newRestore ucloud udisk restore
func newRestore(ctx *cli.Context) *cobra.Command {
	var snapshotIDs *[]string
	client := cli.NewServiceClient(ctx, puhost.NewClient)
	req := client.NewRestoreUHostDiskRequest()
	cmd := &cobra.Command{
		Use:   "restore",
		Short: "Restore udisk from snapshot",
		Long:  "Restore udisk from snapshot",
		Run: func(cmd *cobra.Command, args []string) {
			w := ctx.ProgressWriter()
			results := []cli.OpResultRow{}
			for _, snapshotID := range *snapshotIDs {
				snapshotID = ctx.PickResourceID(snapshotID)
				any, err := describeSnapshotByID(ctx)(snapshotID, nil)
				if err != nil {
					ctx.HandleError(err)
					continue
				}
				snapshot, ok := any.(*puhost.SnapshotSet)
				if !ok {
					fmt.Fprintf(w, "snapshot[%s] doesn't exist\n", snapshotID)
					continue
				}
				if snapshot.UHostId != "" {
					text := fmt.Sprintf("can we detach udisk[%s] from uhost[%s]?", snapshot.DiskId, snapshot.UHostId)
					if !ctx.Confirm(false, text) {
						continue
					}
					DetachUdisk(ctx, false, snapshot.DiskId, w)
				}
				req.SnapshotIds = append(req.SnapshotIds, snapshotID)
				_, err = client.RestoreUHostDisk(req)

				if err != nil {
					ctx.HandleError(err)
					return
				}

				text := fmt.Sprintf("udisk[%s] has been restored from snapshot[%s]", snapshot.DiskId, snapshot.SnapshotId)
				fmt.Fprintln(w, text)
				results = append(results, cli.OpResultRow{ResourceID: snapshot.DiskId, Action: "restore", Status: "Restored"})
			}
			ctx.EmitResult(results...)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	snapshotIDs = flags.StringSlice("snapshot-id", nil, "Required. Resourece ID of the snapshots to restore from")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	req.Zone = flags.String("zone", ctx.DefaultZone(), "Optional. Assign availability zone")
	command.SetCompletion(cmd, "snapshot-id", func() []string {
		return getSnapshotList(ctx, []string{status.SNAPSHOT_NORMAL}, *req.ProjectId, *req.Region, *req.Zone)
	})
	cmd.MarkFlagRequired("snapshot-id")
	return cmd
}

// newSnapshotList ucloud udisk list-snapshot
func newSnapshotList(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, puhost.NewClient)
	req := client.NewDescribeSnapshotRequest()
	cmd := &cobra.Command{
		Use:   "list-snapshot",
		Short: "List snapshots",
		Long:  "List snapshots",
		Run: func(c *cobra.Command, args []string) {
			resp, err := client.DescribeSnapshot(req)
			if err != nil {
				ctx.HandleError(err)
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
					CreationTime:     common.FormatDate(snapshot.CreateTime),
				}
				list = append(list, row)
			}
			ctx.PrintList(list)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	req.Zone = flags.String("zone", ctx.DefaultZone(), "Optional. Assign availability zone")
	// StringSliceVar binds the flag to req.SnapshotIds so Cobra fills it during
	// parse; dereferencing StringSlice() here would freeze it to the initial nil
	// slice and drop the --snapshot-id filter.
	flags.StringSliceVar(&req.SnapshotIds, "snapshot-id", nil, "Optional. Resource ID of snapshots to list")
	req.UHostId = flags.String("uhost-id", "", "Optional. Snapshots of the uhost")
	req.DiskId = flags.String("disk-id", "", "Optional. Snapshots of the udisk")
	req.Offset = cmd.Flags().Int("offset", 0, "Optional. Offset")
	req.Limit = cmd.Flags().Int("limit", 50, "Optional. Limit, length of snapshot list")

	return cmd
}

// newSnapshotDelete ucloud udisk delete-snapshot
func newSnapshotDelete(ctx *cli.Context) *cobra.Command {
	var snapshotIds *[]string
	client := cli.NewServiceClient(ctx, puhost.NewClient)
	req := client.NewDeleteSnapshotRequest()
	cmd := &cobra.Command{
		Use:   "delete-snapshot",
		Short: "Delete snapshots",
		Long:  "Delete snapshots",
		Run: func(c *cobra.Command, args []string) {
			w := ctx.ProgressWriter()
			results := []cli.OpResultRow{}
			for _, snapshotID := range *snapshotIds {
				req.SnapshotId = sdk.String(ctx.PickResourceID(snapshotID))
				resp, err := client.DeleteSnapshot(req)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				fmt.Fprintf(w, "snapshot[%s] deleted\n", resp.SnapshotId)
				results = append(results, cli.OpResultRow{ResourceID: resp.SnapshotId, Action: "delete-snapshot", Status: "Deleted"})
			}
			ctx.EmitResult(results...)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	req.Zone = flags.String("zone", ctx.DefaultZone(), "Optional. Assign availability zone")
	snapshotIds = flags.StringSlice("snapshot-id", nil, "Required. Resource ID of snapshots to delete")
	cmd.MarkFlagRequired("snapshot-id")
	return cmd
}
