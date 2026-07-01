package uhost

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	uhostsdk "github.com/ucloud/ucloud-sdk-go/services/uhost"
	"github.com/ucloud/ucloud-sdk-go/services/unet"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"
	sdkerror "github.com/ucloud/ucloud-sdk-go/ucloud/error"

	"github.com/ucloud/ucloud-cli/internal/common"
	"github.com/ucloud/ucloud-cli/model/status"
	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// _RetCodeRegionNoPermission is the SDK RetCode returned when the account has no
// permission for UHost in a region; the --all-region path skips such regions.
// Verbatim from cmd/uhost.go.
const _RetCodeRegionNoPermission = 230

// NewCommand builds the `uhost` root command and mounts the 14 subcommands in
// the same AddCommand order as cmd/uhost.go NewCmdUHost: list, create, delete,
// stop, start, restart, poweroff, resize, clone, reset-password, reinstall-os,
// create-image, isolation-group (subtree), leave-isolation-group.
func NewCommand(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uhost",
		Short: "List,create,delete,stop,restart,poweroff or resize UHost instance",
		Long:  `List,create,delete,stop,restart,poweroff or resize UHost instance`,
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(newList(ctx))
	cmd.AddCommand(newCreate(ctx))
	cmd.AddCommand(newDelete(ctx))
	cmd.AddCommand(newStop(ctx))
	cmd.AddCommand(newStart(ctx))
	cmd.AddCommand(newReboot(ctx))
	cmd.AddCommand(newPoweroff(ctx))
	cmd.AddCommand(newResize(ctx))
	cmd.AddCommand(newClone(ctx))
	cmd.AddCommand(newResetPassword(ctx))
	cmd.AddCommand(newReinstallOS(ctx))
	cmd.AddCommand(newCreateImage(ctx))
	cmd.AddCommand(newIsolationGroup(ctx))
	cmd.AddCommand(newLeaveIsolationGroup(ctx))

	return cmd
}

// listUhost renders the uhost slice via ctx.PrintList, selecting columns per
// output mode using the per-mode row structs (rows.go). AWS-style: --output
// selects only the format — table shows curated columns (uhostRowDefault, or
// uhostRowAllRegion with a trailing Zone under --all-region); json/yaml always
// emit the full uhostRow so no field is lost (e.g. DiskSet/VPC/Subnet).
func listUhost(ctx *cli.Context, uhosts []uhostsdk.UHostInstanceSet, listAllRegion bool) {
	list := make([]uhostRow, 0)
	for _, host := range uhosts {
		row := uhostRow{}
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
		row.CreationTime = common.FormatDate(host.CreateTime)
		row.State = host.State
		row.Type = host.MachineType + "/" + host.HostType
		if host.HotplugFeature {
			row.Type += "/HotPlug"
		}
		list = append(list, row)
	}

	// JSON/YAML mode: print the full row set (matches cmd/uhost.go, which
	// marshalled the full UHostRow slice in --json mode). ctx.PrintList routes
	// json/yaml by format; for table mode we narrow to the per-mode struct.
	if ctx.Format() != cli.OutputTable {
		ctx.PrintList(list)
		return
	}

	if listAllRegion {
		rows := make([]uhostRowAllRegion, 0, len(list))
		for _, r := range list {
			rows = append(rows, uhostRowAllRegion{
				UHostName: r.UHostName, ResourceID: r.ResourceID, Group: r.Group,
				PrivateIP: r.PrivateIP, PublicIP: r.PublicIP, Config: r.Config,
				Image: r.Image, Type: r.Type, State: r.State,
				CreationTime: r.CreationTime, Zone: r.Zone,
			})
		}
		ctx.PrintList(rows)
		return
	}
	rows := make([]uhostRowDefault, 0, len(list))
	for _, r := range list {
		rows = append(rows, uhostRowDefault{
			UHostName: r.UHostName, ResourceID: r.ResourceID, Group: r.Group,
			PrivateIP: r.PrivateIP, PublicIP: r.PublicIP, Config: r.Config,
			Image: r.Image, Type: r.Type, State: r.State, CreationTime: r.CreationTime,
		})
	}
	ctx.PrintList(rows)
}

func listUhostID(ctx *cli.Context, uhosts []uhostsdk.UHostInstanceSet) {
	ids := make([]string, 0)
	for _, u := range uhosts {
		ids = append(ids, u.UHostId)
	}
	// The id list IS the result of --uhost-id-only, not narration: write it to
	// stdout (ctx.Out), never ProgressWriter — otherwise in non-TTY/json mode the
	// ids go to stderr and `ids=$(ucloud uhost list --uhost-id-only)` captures
	// nothing.
	fmt.Fprintln(ctx.Out(), strings.Join(ids, ","))
}

func fetchUHosts(client *uhostsdk.UHostClient, req *uhostsdk.DescribeUHostInstanceRequest) ([]uhostsdk.UHostInstanceSet, int, error) {
	resp, err := client.DescribeUHostInstance(req)
	if err != nil {
		return nil, 0, err
	}
	return resp.UHostSet, resp.TotalCount, nil
}

func fetchUHostsPageOff(client *uhostsdk.UHostClient, req *uhostsdk.DescribeUHostInstanceRequest) ([]uhostsdk.UHostInstanceSet, error) {
	_req := *req
	result := make([]uhostsdk.UHostInstanceSet, 0)
	for limit, offset := 50, 0; ; offset += limit {
		_req.Offset = sdk.Int(offset)
		_req.Limit = sdk.Int(limit)
		uhosts, total, err := fetchUHosts(client, &_req)
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

func getAllUHosts(ctx *cli.Context, client *uhostsdk.UHostClient, req *uhostsdk.DescribeUHostInstanceRequest, pageOff bool, allRegion bool) ([]uhostsdk.UHostInstanceSet, error) {
	if allRegion {
		result := make([]uhostsdk.UHostInstanceSet, 0)
		regions, err := ctx.AllRegions()
		if err != nil {
			return nil, err
		}
		for _, region := range regions {
			_req := *req
			_req.Region = sdk.String(region)
			//如果要获取所有region的主机，则不分页
			uhosts, err := fetchUHostsPageOff(client, &_req)
			// Has no permission in current region for UHost
			if e, ok := err.(sdkerror.Error); ok && e.Code() == _RetCodeRegionNoPermission {
				continue
			}
			if err != nil {
				return nil, err
			}
			result = append(result, uhosts...)
		}
		return result, nil
	}

	if pageOff {
		_req := *req
		uhosts, err := fetchUHostsPageOff(client, &_req)
		if err != nil {
			return nil, err
		}
		return uhosts, nil
	}

	uhosts, _, err := fetchUHosts(client, req)
	if err != nil {
		return nil, err
	}
	return uhosts, nil
}

// newList ucloud uhost list
func newList(ctx *cli.Context) *cobra.Command {
	var allRegion, pageOff, idOnly bool
	var uhostIds []string
	client := cli.NewServiceClient(ctx, uhostsdk.NewClient)
	req := client.NewDescribeUHostInstanceRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all UHost Instances",
		Long:  `List all UHost Instances`,
		Run: func(cmd *cobra.Command, args []string) {
			*req.VPCId = ctx.PickResourceID(*req.VPCId)
			*req.SubnetId = ctx.PickResourceID(*req.SubnetId)
			*req.IsolationGroup = ctx.PickResourceID(*req.IsolationGroup)
			for _, uhost := range uhostIds {
				req.UHostIds = append(req.UHostIds, ctx.PickResourceID(uhost))
			}

			uhosts, err := getAllUHosts(ctx, client, req, pageOff, allRegion)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if idOnly {
				listUhostID(ctx, uhosts)
			} else {
				listUhost(ctx, uhosts, allRegion)
			}
		},
	}
	cmd.Flags().SortFlags = false
	req.ProjectId = cmd.Flags().String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = cmd.Flags().String("region", ctx.DefaultRegion(), "Optional. Assign region.")
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
	ctx.BindGroup(cmd, req)

	command.SetFlagValues(cmd, "page-off", "true", "false")
	command.SetFlagValues(cmd, "uhost-id-only", "true", "false")
	command.SetCompletion(cmd, "project-id", ctx.ProjectList)
	command.SetCompletion(cmd, "region", ctx.RegionList)
	command.SetCompletion(cmd, "zone", func() []string {
		return ctx.ZoneList(req.GetRegion())
	})

	command.SetCompletion(cmd, "vpc-id", func() []string {
		return getAllVPCIdNames(ctx, *req.ProjectId, *req.Region)
	})
	command.SetCompletion(cmd, "subnet-id", func() []string {
		return getAllSubnetIDNames(ctx, *req.VPCId, *req.ProjectId, *req.Region)
	})
	command.SetCompletion(cmd, "isolation-group", func() []string {
		return getIsolationGroupList(ctx, *req.ProjectId, *req.Region)
	})
	command.SetCompletion(cmd, "uhost-id", func() []string {
		return getUhostList(ctx, nil, *req.ProjectId, *req.Region, *req.Zone)
	})

	return cmd
}

// newStop ucloud uhost stop
func newStop(ctx *cli.Context) *cobra.Command {
	var uhostIDs *[]string
	var async *bool
	client := cli.NewServiceClient(ctx, uhostsdk.NewClient)
	req := client.NewStopUHostInstanceRequest()
	cmd := &cobra.Command{
		Use:     "stop",
		Short:   "Shut down uhost instance",
		Long:    "Shut down uhost instance",
		Example: "ucloud uhost stop --uhost-id uhost-xxx1,uhost-xxx2",
		Run: func(cmd *cobra.Command, args []string) {
			for _, id := range *uhostIDs {
				id = ctx.PickResourceID(id)
				req.UHostId = &id
				stopUhostIns(ctx, client, req, *async)
			}
		},
	}
	cmd.Flags().SortFlags = false
	uhostIDs = cmd.Flags().StringSlice("uhost-id", nil, "Required. ResourceIDs(UHostIds) of the uhost instances")
	req.ProjectId = cmd.Flags().String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = cmd.Flags().String("region", ctx.DefaultRegion(), "Optional. Assign region")
	req.Zone = cmd.Flags().String("zone", "", "Optional. Assign availability zone")
	async = cmd.Flags().Bool("async", false, "Optional. Do not wait for the long-running operation to finish.")
	command.SetCompletion(cmd, "uhost-id", func() []string {
		return getUhostList(ctx, []string{status.HOST_RUNNING}, *req.ProjectId, *req.Region, *req.Zone)
	})
	cmd.MarkFlagRequired("uhost-id")

	return cmd
}

// stopUhostIns stops a uhost and (unless async) polls it to Stopped. Mirrors
// cmd/uhost.go stopUhostIns (sequential base.NewPoller → ctx.PollerTo.Spoll).
func stopUhostIns(ctx *cli.Context, client *uhostsdk.UHostClient, req *uhostsdk.StopUHostInstanceRequest, async bool) bool {
	w := ctx.ProgressWriter()
	resp, err := client.StopUHostInstance(req)
	if err != nil {
		ctx.HandleError(err)
		return false
	}

	text := fmt.Sprintf("uhost[%v] is shutting down", resp.UHostId)
	if async {
		fmt.Fprintln(w, text)
		return false
	}
	// base.Poller.Poll returned a bool (reached target state) that cmd/uhost.go
	// fed back into resize (inst.State = Stopped). The platform Spoll narrates to
	// the writer but returns nothing, so we return true here: a successful
	// (non-async) stop request that we then polled is treated as "stopped" for
	// the resize state-transition, which matches the original intent.
	ctx.PollerTo(w, describeUHostByID(ctx, *req.ProjectId, *req.Region, *req.Zone)).Spoll(resp.UHostId, text, []string{status.HOST_STOPPED, status.HOST_FAIL})
	return true
}

// promptStopUhostIns prompts (unless yes) then stops the uhost. Mirrors
// cmd/uhost.go promptStopUhostIns.
func promptStopUhostIns(ctx *cli.Context, client *uhostsdk.UHostClient, req *uhostsdk.StopUHostInstanceRequest, yes, async bool, promptText string) bool {
	if !ctx.Confirm(yes, promptText) {
		return false
	}
	return stopUhostIns(ctx, client, req, false)
}

// newStart ucloud uhost start
func newStart(ctx *cli.Context) *cobra.Command {
	var async *bool
	var uhostIDs *[]string
	client := cli.NewServiceClient(ctx, uhostsdk.NewClient)
	req := client.NewStartUHostInstanceRequest()
	cmd := &cobra.Command{
		Use:     "start",
		Short:   "Start Uhost instance",
		Long:    "Start Uhost instance",
		Example: "ucloud uhost start --uhost-id uhost-xxx1,uhost-xxx2",
		Run: func(cmd *cobra.Command, args []string) {
			w := ctx.ProgressWriter()
			for _, id := range *uhostIDs {
				id := ctx.PickResourceID(id)
				req.UHostId = &id
				resp, err := client.StartUHostInstance(req)
				if err != nil {
					ctx.HandleError(err)
				} else {
					text := fmt.Sprintf("uhost[%v] is starting", resp.UHostId)
					if *async {
						fmt.Fprintln(w, text)
					} else {
						ctx.PollerTo(w, describeUHostByID(ctx, *req.ProjectId, *req.Region, *req.Zone)).Spoll(resp.UHostId, text, []string{status.HOST_RUNNING, status.HOST_FAIL})
					}
				}
			}
		},
	}
	cmd.Flags().SortFlags = false
	uhostIDs = cmd.Flags().StringSlice("uhost-id", nil, "Requried. ResourceIDs(UHostIds) of the uhost instance")
	req.ProjectId = cmd.Flags().String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = cmd.Flags().String("region", ctx.DefaultRegion(), "Optional. Assign region")
	req.Zone = cmd.Flags().String("zone", "", "Optional. Assign availability zone")
	async = cmd.Flags().Bool("async", false, "Optional. Do not wait for the long-running operation to finish.")
	command.SetCompletion(cmd, "uhost-id", func() []string {
		return getUhostList(ctx, []string{status.HOST_STOPPED}, *req.ProjectId, *req.Region, *req.Zone)
	})
	cmd.MarkFlagRequired("uhost-id")
	return cmd
}

// newReboot ucloud uhost restart
func newReboot(ctx *cli.Context) *cobra.Command {
	var uhostIDs *[]string
	var async *bool
	client := cli.NewServiceClient(ctx, uhostsdk.NewClient)
	req := client.NewRebootUHostInstanceRequest()
	cmd := &cobra.Command{
		Use:     "restart",
		Short:   "Restart uhost instance",
		Long:    "Restart uhost instance",
		Example: "ucloud uhost restart --uhost-id uhost-xxx1,uhost-xxx2",
		Run: func(cmd *cobra.Command, args []string) {
			w := ctx.ProgressWriter()
			for _, id := range *uhostIDs {
				id = ctx.PickResourceID(id)
				req.UHostId = &id
				resp, err := client.RebootUHostInstance(req)
				if err != nil {
					ctx.HandleError(err)
				} else {
					text := fmt.Sprintf("uhost[%v] is restarting", resp.UHostId)
					if *async {
						fmt.Fprintln(w, text)
					} else {
						ctx.PollerTo(w, describeUHostByID(ctx, *req.ProjectId, *req.Region, *req.Zone)).Spoll(resp.UHostId, text, []string{status.HOST_RUNNING, status.HOST_FAIL})
					}
				}
			}
		},
	}
	cmd.Flags().SortFlags = false
	uhostIDs = cmd.Flags().StringSlice("uhost-id", nil, "Required. ResourceIDs(UHostIds) of the uhost instance")
	req.ProjectId = cmd.Flags().String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = cmd.Flags().String("region", ctx.DefaultRegion(), "Optional. Assign region")
	req.Zone = cmd.Flags().String("zone", "", "Optional. Assign availability zone")
	req.DiskPassword = cmd.Flags().String("disk-password", "", "Optional. Encrypted disk password")
	async = cmd.Flags().Bool("async", false, "Optional. Do not wait for the long-running operation to finish.")
	command.SetCompletion(cmd, "uhost-id", func() []string {
		return getUhostList(ctx, []string{status.HOST_FAIL, status.HOST_RUNNING, status.HOST_STOPPED}, *req.ProjectId, *req.Region, *req.Zone)
	})
	cmd.MarkFlagRequired("uhost-id")
	return cmd
}

// newPoweroff ucloud uhost poweroff
func newPoweroff(ctx *cli.Context) *cobra.Command {
	var yes *bool
	var uhostIDs *[]string
	client := cli.NewServiceClient(ctx, uhostsdk.NewClient)
	req := client.NewPoweroffUHostInstanceRequest()
	cmd := &cobra.Command{
		Use:     "poweroff",
		Short:   "Analog power off Uhost instnace",
		Long:    "Analog power off Uhost instnace",
		Example: "ucloud uhost poweroff --uhost-id uhost-xxx1,uhost-xxx2",
		Run: func(cmd *cobra.Command, args []string) {
			w := ctx.ProgressWriter()
			confirmText := "Danger, it may affect data integrity. Are you sure you want to poweroff this uhost?"
			if len(*uhostIDs) > 1 {
				confirmText = "Danger, it may affect data integrity. Are you sure you want to poweroff those uhosts?"
			}
			if !ctx.Confirm(*yes, confirmText) {
				return
			}
			for _, id := range *uhostIDs {
				id = ctx.PickResourceID(id)
				req.UHostId = &id
				resp, err := client.PoweroffUHostInstance(req)
				if err != nil {
					ctx.HandleError(err)
				} else {
					fmt.Fprintf(w, "uhost[%v] is power off\n", resp.UHostId)
				}
			}
		},
	}
	cmd.Flags().SortFlags = false
	uhostIDs = cmd.Flags().StringSlice("uhost-id", nil, "ResourceIDs(UHostIds) of the uhost instance")
	req.ProjectId = cmd.Flags().String("project-id", ctx.DefaultProjectID(), "Assign project-id")
	req.Region = cmd.Flags().String("region", ctx.DefaultRegion(), "Assign region")
	req.Zone = cmd.Flags().String("zone", "", "Assign availability zone")
	yes = cmd.Flags().BoolP("yes", "y", false, "Optional. Do not prompt for confirmation.")

	command.SetCompletion(cmd, "uhost-id", func() []string {
		return getUhostList(ctx, []string{status.HOST_FAIL, status.HOST_RUNNING, status.HOST_STOPPED}, *req.ProjectId, *req.Region, *req.Zone)
	})
	cmd.MarkFlagRequired("uhost-id")

	return cmd
}

// resizeAttachedDisk resizes a uhost's attached disk, stopping the uhost first
// if it is running. Mirrors cmd/uhost.go resizeAttachedDisk.
func resizeAttachedDisk(ctx *cli.Context, client *uhostsdk.UHostClient, req *uhostsdk.ResizeAttachedDiskRequest, host *uhostsdk.UHostInstanceSet, yes, async bool, promptText string) error {
	w := ctx.ProgressWriter()
	req.UHostId = &host.UHostId
	if host.State == status.HOST_RUNNING {
		err := tryStopUhost(ctx, client, req, host.UHostId, promptText, yes, async)
		if err != nil {
			return fmt.Errorf("try to stop uhost error :%w", err)
		}
	}
	req.DryRun = sdk.Bool(false)
	_, err := client.ResizeAttachedDisk(req)
	if err != nil {
		return err
	}
	text := fmt.Sprintf("uhost [%s] disk [%s] resize", host.UHostId, *req.DiskId)
	if async {
		fmt.Fprintln(w, text)
	} else {
		ctx.PollerTo(w, describeUHostByID(ctx, *req.ProjectId, *req.Region, *req.Zone)).Spoll(host.UHostId, text, []string{status.HOST_RUNNING, status.HOST_STOPPED, status.HOST_FAIL})
	}
	return nil
}

func tryStopUhost(ctx *cli.Context, client *uhostsdk.UHostClient, req *uhostsdk.ResizeAttachedDiskRequest, uhostID, promptText string, yes, async bool) error {
	req.DryRun = sdk.Bool(true)
	resp, err := client.ResizeAttachedDisk(req)
	if err != nil {
		return err
	}
	if resp.NeedRestart {
		stopReq := client.NewStopUHostInstanceRequest()
		stopReq.UHostId = &uhostID
		stopReq.ProjectId = req.ProjectId
		stopReq.Region = req.Region
		stopReq.Zone = req.Zone
		promptStopUhostIns(ctx, client, stopReq, yes, async, promptText)
	}
	return nil
}

// newResize ucloud uhost resize
func newResize(ctx *cli.Context) *cobra.Command {
	var yes, async *bool
	var bootDiskSize, dataDiskSize int
	var dataDiskID string
	var uhostIDs *[]string
	client := cli.NewServiceClient(ctx, uhostsdk.NewClient)
	req := client.NewResizeUHostInstanceRequest()
	cmd := &cobra.Command{
		Use:     "resize",
		Short:   "Resize uhost instance,such as cpu core count, memory size and disk size",
		Long:    "Resize uhost instance,such as cpu core count, memory size and disk size",
		Example: "ucloud uhost resize --uhost-id uhost-xxx1,uhost-xxx2 --cpu 4 --memory-gb 8",
		Run: func(cmd *cobra.Command, args []string) {
			w := ctx.ProgressWriter()
			if *req.CPU == 0 {
				req.CPU = nil
			}
			if *req.Memory == 0 {
				req.Memory = nil
			} else {
				*req.Memory *= 1024
			}
			for _, id := range *uhostIDs {
				id = ctx.PickResourceID(id)
				req.UHostId = &id
				host, err := describeUHostByID(ctx, *req.ProjectId, *req.Region, *req.Zone)(id, nil)
				if err != nil {
					fmt.Fprintln(ctx.Err(), err)
					return
				}
				inst := host.(*uhostsdk.UHostInstanceSet)
				stopReq := client.NewStopUHostInstanceRequest()
				stopReq.ProjectId = req.ProjectId
				stopReq.Region = req.Region
				stopReq.Zone = req.Zone
				stopReq.UHostId = &id
				confirmText := "Resize uhost must be done after the uhost is stopped. Do you want to stop this uhost?"
				if req.CPU != nil || req.Memory != nil || *req.NetCapValue != 0 {
					if inst.State == status.HOST_RUNNING {
						ret := promptStopUhostIns(ctx, client, stopReq, *yes, *async, confirmText)
						if ret {
							inst.State = status.HOST_STOPPED
						}
					}
					resp, err := client.ResizeUHostInstance(req)
					if err != nil {
						ctx.HandleError(err)
					} else {
						text := fmt.Sprintf("uhost [%v] cpu, memory resize", resp.UHostId)
						if *async {
							fmt.Fprintln(w, text)
						} else {
							ctx.PollerTo(w, describeUHostByID(ctx, *req.ProjectId, *req.Region, *req.Zone)).Spoll(resp.UHostId, text, []string{status.HOST_RUNNING, status.HOST_STOPPED, status.HOST_FAIL})
						}
					}
				}

				if dataDiskSize != 0 || bootDiskSize != 0 {
					_req := client.NewResizeAttachedDiskRequest()
					var bootDisk uhostsdk.UHostDiskSet
					var dataDisks = map[string]uhostsdk.UHostDiskSet{}
					for _, disk := range inst.DiskSet {
						if disk.IsBoot == "True" {
							bootDisk = disk
						} else if disk.IsBoot == "False" {
							dataDisks[disk.DiskId] = disk
						}
					}
					if bootDiskSize != 0 {
						if bootDiskSize <= bootDisk.Size {
							ctx.LogError(fmt.Sprintf("Error, disk does not support shrinkage. current system-disk-size %dg", bootDisk.Size))
							continue
						} else {
							_req.DiskSpace = &bootDiskSize
							_req.DiskId = &bootDisk.DiskId
						}
						err := resizeAttachedDisk(ctx, client, _req, inst, *yes, *async, confirmText)
						if err != nil {
							ctx.HandleError(err)
						}
					}

					if dataDiskSize != 0 {
						var dataDisk uhostsdk.UHostDiskSet
						if len(dataDisks) > 1 {
							if dataDiskID == "" {
								ctx.LogError(fmt.Sprintf("Error, the uhost %s have %d data disks. data-disk-id should be assigned", id, len(dataDisks)))
								continue
							}
							var ok bool
							dataDisk, ok = dataDisks[dataDiskID]
							if !ok {
								ctx.LogError(fmt.Sprintf("Error, the disk %s does not exist", dataDiskID))
								continue
							}
						} else if len(dataDisks) == 1 {
							for _, disk := range dataDisks {
								dataDisk = disk
							}
						} else if len(dataDisks) == 0 {
							ctx.LogError(fmt.Sprintf("Error, the uhost %s have no data disk. data-disk-id should be assigned", id))
							continue
						}
						if dataDiskSize <= dataDisk.Size {
							ctx.LogError(fmt.Sprintf("Error, disk does not support shrinkage. current data-disk-size %dg", dataDisk.Size))
							continue
						}
						_req.DiskSpace = &dataDiskSize
						_req.DiskId = &dataDisk.DiskId
						err := resizeAttachedDisk(ctx, client, _req, inst, *yes, *async, confirmText)
						if err != nil {
							ctx.HandleError(err)
						}
					}
				}
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	uhostIDs = cmd.Flags().StringSlice("uhost-id", nil, "Required. ResourceIDs(or UhostIDs) of the uhost instances")
	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	req.CPU = cmd.Flags().Int("cpu", 0, "Optional. The number of virtual CPU cores. Series1 {1, 2, 4, 8, 12, 16, 24, 32}. Series2 {1,2,4,8,16}")
	req.Memory = cmd.Flags().Int("memory-gb", 0, "Optional. memory size. Unit: GB. Range: [1, 128], multiple of 2")
	cmd.Flags().IntVar(&bootDiskSize, "system-disk-size-gb", 0, "Optional. System disk size, unit GB. Range[20,100]. Step 10. System disk does not support shrinkage")
	cmd.Flags().IntVar(&dataDiskSize, "data-disk-size-gb", 0, "Optional. Data disk size,unit GB. Step 10. disk does not support shrinkage")
	cmd.Flags().StringVar(&dataDiskID, "data-disk-id", "", "Optional. If the uhost specified has two or more data disks, this parameter should be assigned")
	req.NetCapValue = cmd.Flags().Int("net-cap", 0, "Optional. NIC scale. 1,upgrade; 2,downgrade; 0,unchanged")
	yes = cmd.Flags().BoolP("yes", "y", false, "Optional. Do not prompt for confirmation.")
	async = cmd.Flags().BoolP("async", "a", false, "Optional. Do not wait for the long-running operation to finish.")
	command.SetCompletion(cmd, "uhost-id", func() []string {
		return getUhostList(ctx, []string{status.HOST_RUNNING, status.HOST_STOPPED, status.HOST_FAIL}, *req.ProjectId, *req.Region, *req.Zone)
	})
	cmd.MarkFlagRequired("uhost-id")
	return cmd
}

// newClone ucloud uhost clone
func newClone(ctx *cli.Context) *cobra.Command {
	var uhostID *string
	var async *bool

	var password string
	var keyPairId string

	client := cli.NewServiceClient(ctx, uhostsdk.NewClient)
	unetClient := cli.NewServiceClient(ctx, unet.NewClient)
	req := client.NewCreateUHostInstanceRequest()
	cmd := &cobra.Command{
		Use:   "clone",
		Short: "Create an uhost with the same configuration as another uhost, excluding bound eip and udisk",
		Long:  "Create an uhost with the same configuration as another uhost, excluding bound eip and udisk",
		Run: func(com *cobra.Command, args []string) {
			w := ctx.ProgressWriter()
			if len(password) > 0 {
				req.LoginMode = sdk.String("Password")
				req.KeyPairId = nil
				req.Password = sdk.String(password)
			} else if len(keyPairId) > 0 {
				req.LoginMode = sdk.String("KeyPair")
				req.KeyPairId = sdk.String(keyPairId)
				req.Password = nil
			} else {
				fmt.Fprintln(ctx.Err(), errors.New("password or key-pair-id is required"))
				return
			}
			*uhostID = ctx.PickResourceID(*uhostID)
			queryReq := client.NewDescribeUHostInstanceRequest()
			queryReq.ProjectId = req.ProjectId
			queryReq.Region = req.Region
			queryReq.Zone = req.Zone
			queryReq.UHostIds = []string{*uhostID}
			queryResp, err := client.DescribeUHostInstance(queryReq)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if len(queryResp.UHostSet) < 1 {
				fmt.Fprintln(ctx.Err(), fmt.Errorf("uhost[%s] not exist", *uhostID))
				return
			}
			if queryResp.UHostSet[0].SecGroupInstance == true {
				fmt.Fprintln(ctx.Err(), fmt.Errorf("uhost[%s] is in security groups, it is not allowed to clone", *uhostID))
				return
			}
			queryFirewallReq := unetClient.NewDescribeFirewallRequest()
			queryFirewallReq.ProjectId = req.ProjectId
			queryFirewallReq.Region = req.Region
			queryFirewallReq.ResourceId = uhostID
			queryFirewallReq.ResourceType = sdk.String("uhost")

			firewallResp, err := unetClient.DescribeFirewall(queryFirewallReq)
			if err != nil {
				ctx.HandleError(err)
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
				item := uhostsdk.UHostDisk{
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
			resp, err := client.CreateUHostInstance(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if len(resp.UHostIds) == 1 {
				text := fmt.Sprintf("cloned uhost:[%s] is initializing", resp.UHostIds[0])
				if *async {
					fmt.Fprintln(w, text)
				} else {
					ctx.PollerTo(w, describeUHostByID(ctx, *req.ProjectId, *req.Region, *req.Zone)).Spoll(resp.UHostIds[0], text, []string{status.HOST_RUNNING, status.HOST_FAIL})
				}
			} else {
				ctx.HandleError(fmt.Errorf("expect uhost count 1, accept %d", len(resp.UHostIds)))
				return
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	uhostID = flags.String("uhost-id", "", "Required. Resource ID of the uhost to clone from")
	flags.StringVar(&password, "password", "", "Optional. Password of the uhost user(root/ubuntu)")
	flags.StringVar(&keyPairId, "key-pair-id", "", "Optional. Resource ID of ssh key pair. See 'ucloud api --Action DescribeUHostKeyPairs' Where both password and key-pair-id are set, the key-pair-id is ignored")

	req.Name = flags.String("name", "", "Optional. Name of the uhost to clone")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	req.Zone = flags.String("zone", ctx.DefaultZone(), "Optional. Assign availability zone")
	async = flags.Bool("async", false, "Optional. Do not wait for the long-running operation to finish.")
	command.SetCompletion(cmd, "uhost-id", func() []string {
		return getUhostList(ctx, []string{status.HOST_RUNNING, status.HOST_STOPPED}, *req.ProjectId, *req.Region, *req.Zone)
	})
	cmd.MarkFlagRequired("uhost-id")
	return cmd
}

// newCreateImage ucloud uhost create-image
func newCreateImage(ctx *cli.Context) *cobra.Command {
	var async *bool
	client := cli.NewServiceClient(ctx, uhostsdk.NewClient)
	req := client.NewCreateCustomImageRequest()
	cmd := &cobra.Command{
		Use:   "create-image",
		Short: "Create image from an uhost instance",
		Long:  "Create image from an uhost instance",
		Run: func(cmd *cobra.Command, args []string) {
			w := ctx.ProgressWriter()
			req.UHostId = sdk.String(ctx.PickResourceID(*req.UHostId))
			resp, err := client.CreateCustomImage(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			// "iamge[%s] is making" typo preserved verbatim from cmd/uhost.go.
			text := fmt.Sprintf("iamge[%s] is making", resp.ImageId)
			if *async {
				fmt.Fprintln(w, text)
			} else {
				ctx.PollerTo(w, describeImageByID(ctx, *req.ProjectId, *req.Region, *req.Zone)).Spoll(resp.ImageId, text, []string{status.IMAGE_AVAILABLE, status.IMAGE_UNAVAILABLE})
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	req.UHostId = flags.String("uhost-id", "", "Resource ID of uhost to create image from")
	req.ImageName = flags.String("image-name", "", "Required. Name of the image to create")
	req.ImageDescription = flags.String("image-desc", "", "Optional. Description of the image to create")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	req.Zone = flags.String("zone", ctx.DefaultZone(), "Optional. Assign availability zone")
	async = flags.BoolP("async", "a", false, "Optional. Do not wait for the long-running operation to finish.")

	command.SetCompletion(cmd, "uhost-id", func() []string {
		return getUhostList(ctx, []string{status.HOST_RUNNING, status.HOST_STOPPED}, *req.ProjectId, *req.Region, *req.Zone)
	})

	cmd.MarkFlagRequired("uhost-id")
	cmd.MarkFlagRequired("image-name")
	return cmd
}

// newResetPassword ucloud uhost reset-password
func newResetPassword(ctx *cli.Context) *cobra.Command {
	var yes *bool
	var uhostIDs *[]string
	client := cli.NewServiceClient(ctx, uhostsdk.NewClient)
	req := client.NewResetUHostInstancePasswordRequest()
	cmd := &cobra.Command{
		Use:   "reset-password",
		Short: "Reset the administrator password for the UHost instances.",
		Long:  "Reset the administrator password for the UHost instances.",
		Run: func(cmd *cobra.Command, args []string) {
			w := ctx.ProgressWriter()
			for _, id := range *uhostIDs {
				id = ctx.PickResourceID(id)
				req.UHostId = &id
				err := checkAndCloseUhost(ctx, client, *yes, false, id, *req.ProjectId, *req.Region, *req.Zone)
				if err != nil {
					fmt.Fprintln(ctx.Err(), err)
					continue
				}
				host, err := describeUHostByID(ctx, *req.ProjectId, *req.Region, *req.Zone)(id, nil)
				inst, ok := host.(*uhostsdk.UHostInstanceSet)
				if !ok {
					return
				}
				if inst.BootDiskState == "Initializing" {
					fmt.Fprintf(w, "uhost[%s] boot disk in initializing, wait 10 minutes\n", id)
					return
				}
				resp, err := client.ResetUHostInstancePassword(req)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				fmt.Fprintf(w, "uhost[%s] reset password\n", resp.UHostId)
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	uhostIDs = flags.StringSlice("uhost-id", nil, "Required. Resource IDs of the uhosts to reset the administrator's password")
	req.Password = flags.String("password", "", "Required. New Password")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	req.Zone = flags.String("zone", ctx.DefaultZone(), "Optional. Assign availability zone")
	yes = cmd.Flags().BoolP("yes", "y", false, "Optional. Do not prompt for confirmation.")
	command.SetCompletion(cmd, "uhost-id", func() []string {
		return getUhostList(ctx, []string{status.HOST_RUNNING, status.HOST_STOPPED}, *req.ProjectId, *req.Region, *req.Zone)
	})
	cmd.MarkFlagRequired("uhost-id")
	cmd.MarkFlagRequired("password")
	return cmd
}

// checkAndCloseUhost stops the uhost (with optional prompt) if it is running.
// Mirrors cmd/uhost.go checkAndCloseUhost.
func checkAndCloseUhost(ctx *cli.Context, client *uhostsdk.UHostClient, yes, async bool, uhostID, project, region, zone string) error {
	host, err := describeUHostByID(ctx, project, region, zone)(uhostID, nil)
	if err != nil {
		return err
	}
	inst, ok := host.(*uhostsdk.UHostInstanceSet)
	if ok {
		if inst.State == "Running" {
			if !ctx.Confirm(yes, fmt.Sprintf("uhost[%s] will be stopped, can we do this?", uhostID)) {
				return fmt.Errorf("skip, you do not agree to stop uhost")
			}
			_req := client.NewStopUHostInstanceRequest()
			_req.ProjectId = &project
			_req.Region = &region
			_req.Zone = &zone
			_req.UHostId = &uhostID
			stopUhostIns(ctx, client, _req, async)
		}
	} else {
		return fmt.Errorf("Something wrong, uhost[%s] may not exist", uhostID)
	}
	return nil
}

// newReinstallOS ucloud uhost reinstall-os
func newReinstallOS(ctx *cli.Context) *cobra.Command {
	var isReserveDataDisk, yes, async *bool
	var password, keyPairId string
	client := cli.NewServiceClient(ctx, uhostsdk.NewClient)
	req := client.NewReinstallUHostInstanceRequest()
	cmd := &cobra.Command{
		Use:   "reinstall-os",
		Short: "Reinstall the operating system of the UHost instance",
		Long:  "Reinstall the operating system of the UHost instance. we will detach all udisk disks if the uhost attached some, and then stop the uhost if it's running",
		Run: func(cmd *cobra.Command, args []string) {
			w := ctx.ProgressWriter()
			if *isReserveDataDisk {
				req.ReserveDisk = sdk.String("Yes")
			} else {
				req.ReserveDisk = sdk.String("No")
			}
			req.UHostId = sdk.String(ctx.PickResourceID(*req.UHostId))
			if len(password) > 0 {
				req.LoginMode = sdk.String("Password")
				req.KeyPairId = nil
				req.Password = sdk.String(password)
			} else if len(keyPairId) > 0 {
				req.LoginMode = sdk.String("KeyPair")
				req.KeyPairId = sdk.String(keyPairId)
				req.Password = nil
			} else {
				fmt.Fprintln(ctx.Err(), fmt.Errorf("password or key-pair-id is required"))
				return
			}

			any, err := describeUHostByID(ctx, *req.ProjectId, *req.Region, *req.Zone)(*req.UHostId, nil)
			if err != nil {
				fmt.Fprintln(ctx.Err(), err)
				return
			}
			uhostIns, ok := any.(*uhostsdk.UHostInstanceSet)
			if ok {
				for _, disk := range uhostIns.DiskSet {
					if disk.Type == "Udisk" {
						sure := false
						if !*yes {
							text := fmt.Sprintf("udisk[%s/%s] will be detached, can we do this?", disk.DiskId, disk.Name)
							sure = ctx.Confirm(false, text)
							if !sure {
								fmt.Fprintf(w, "you don't agree to detach udisk\n")
								return
							}
						}
						if *yes || sure {
							err := detachUdisk(ctx, false, disk.DiskId, w)
							if err != nil {
								fmt.Fprintln(ctx.Err(), err)
								return
							}
						}
					}
				}
			} else {
				fmt.Fprintf(w, "Something wrong, uhost[%s] may not exist\n", *req.UHostId)
				return
			}

			err = checkAndCloseUhost(ctx, client, *yes, *async, *req.UHostId, *req.ProjectId, *req.Region, *req.Zone)
			if err != nil {
				fmt.Fprintln(ctx.Err(), err)
				return
			}
			resp, err := client.ReinstallUHostInstance(req)
			if err != nil {
				fmt.Fprintln(ctx.Err(), err)
				return
			}
			text := fmt.Sprintf("uhost[%s] is reinstalling OS", *req.UHostId)
			if *async {
				fmt.Fprintln(w, text)
			} else {
				ctx.PollerTo(w, describeUHostByID(ctx, *req.ProjectId, *req.Region, *req.Zone)).Spoll(resp.UHostId, text, []string{status.HOST_RUNNING, status.HOST_FAIL})
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.UHostId = flags.String("uhost-id", "", "Required. Resource ID of the uhost to reinstall operating system")
	flags.StringVar(&password, "password", "", "Optional. Password of the uhost user(root/ubuntu)")
	flags.StringVar(&keyPairId, "key-pair-id", "", "Optional. Resource ID of ssh key pair. See 'ucloud api --Action DescribeUHostKeyPairs' Where both password and key-pair-id are set, the key-pair-id is ignored")
	req.ImageId = flags.String("image-id", "", "Optional. Resource ID the image to install. See 'ucloud image list'. Default is original image of the uhost")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	req.Zone = flags.String("zone", ctx.DefaultZone(), "Optional. Assign availability zone")
	isReserveDataDisk = flags.Bool("keep-data-disk", false, "Keep data disk or not. If you keep data disk, you can't change OS type(Linux->Window,e.g.)")
	yes = cmd.Flags().BoolP("yes", "y", false, "Optional. Do not prompt for confirmation.")
	async = flags.BoolP("async", "a", false, "Optional. Do not wait for the long-running operation to finish.")
	command.SetCompletion(cmd, "uhost-id", func() []string {
		return getUhostList(ctx, []string{status.HOST_RUNNING, status.HOST_STOPPED}, *req.ProjectId, *req.Region, *req.Zone)
	})
	cmd.MarkFlagRequired("uhost-id")
	return cmd
}
