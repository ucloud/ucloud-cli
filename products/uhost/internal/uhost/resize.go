package uhost

import (
	"fmt"

	"github.com/spf13/cobra"

	uhostsdk "github.com/ucloud/ucloud-sdk-go/services/uhost"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

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
					ctx.HandleError(err)
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
					if inst.State == HOST_RUNNING {
						stop, err := promptStopUhostIns(ctx, client, stopReq, *yes, confirmText)
						if err != nil {
							ctx.HandleError(err)
							return
						}
						if !stop.proceed {
							continue
						}
						if stop.stopped {
							inst.State = HOST_STOPPED
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
							ctx.PollerTo(w, describeUHostByID(ctx, *req.ProjectId, *req.Region, *req.Zone)).Spoll(resp.UHostId, text, []string{HOST_RUNNING, HOST_STOPPED, HOST_FAIL})
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
		return getUhostList(ctx, []string{HOST_RUNNING, HOST_STOPPED, HOST_FAIL}, *req.ProjectId, *req.Region, *req.Zone)
	})
	cmd.MarkFlagRequired("uhost-id")
	return cmd
}

// resizeAttachedDisk resizes a uhost's attached disk, stopping the uhost first
// if it is running. Mirrors cmd/uhost.go resizeAttachedDisk.
func resizeAttachedDisk(ctx *cli.Context, client *uhostsdk.UHostClient, req *uhostsdk.ResizeAttachedDiskRequest, host *uhostsdk.UHostInstanceSet, yes, async bool, promptText string) error {
	w := ctx.ProgressWriter()
	req.UHostId = &host.UHostId
	if host.State == HOST_RUNNING {
		proceed, err := tryStopUhost(ctx, client, req, host.UHostId, promptText, yes)
		if err != nil {
			return fmt.Errorf("try to stop uhost error :%w", err)
		}
		if !proceed {
			return nil
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
		ctx.PollerTo(w, describeUHostByID(ctx, *req.ProjectId, *req.Region, *req.Zone)).Spoll(host.UHostId, text, []string{HOST_RUNNING, HOST_STOPPED, HOST_FAIL})
	}
	return nil
}

func tryStopUhost(ctx *cli.Context, client *uhostsdk.UHostClient, req *uhostsdk.ResizeAttachedDiskRequest, uhostID, promptText string, yes bool) (bool, error) {
	req.DryRun = sdk.Bool(true)
	resp, err := client.ResizeAttachedDisk(req)
	if err != nil {
		return false, err
	}
	if resp.NeedRestart {
		stopReq := client.NewStopUHostInstanceRequest()
		stopReq.UHostId = &uhostID
		stopReq.ProjectId = req.ProjectId
		stopReq.Region = req.Region
		stopReq.Zone = req.Zone
		stop, err := promptStopUhostIns(ctx, client, stopReq, yes, promptText)
		if err != nil {
			return false, err
		}
		return stop.proceed, nil
	}
	return true, nil
}

type stopPromptResult struct {
	proceed bool
	stopped bool
}

// promptStopUhostIns prompts (unless yes) then stops the uhost. proceed is false
// only when the user declined or StopUHostInstance failed. Resize prerequisites
// always wait for the stop; --async only controls the resize operation itself.
func promptStopUhostIns(ctx *cli.Context, client *uhostsdk.UHostClient, req *uhostsdk.StopUHostInstanceRequest, yes bool, promptText string) (stopPromptResult, error) {
	ok, err := ctx.Confirm(yes, promptText)
	if err != nil {
		return stopPromptResult{}, err
	}
	if !ok {
		return stopPromptResult{}, nil
	}
	stop := stopUhostIns(ctx, client, req, false)
	return stopPromptResult{proceed: stop.requested, stopped: stop.stopped}, nil
}
