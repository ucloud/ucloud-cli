package uhost

import (
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	udisksdk "github.com/ucloud/ucloud-sdk-go/services/udisk"
	uhostsdk "github.com/ucloud/ucloud-sdk-go/services/uhost"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

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
				ctx.HandleError(fmt.Errorf("password or key-pair-id is required"))
				return
			}

			any, err := describeUHostByID(ctx, *req.ProjectId, *req.Region, *req.Zone)(*req.UHostId, nil)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			uhostIns, ok := any.(*uhostsdk.UHostInstanceSet)
			if ok {
				for _, disk := range uhostIns.DiskSet {
					if disk.Type == "Udisk" {
						sure := false
						if !*yes {
							text := fmt.Sprintf("udisk[%s/%s] will be detached, can we do this?", disk.DiskId, disk.Name)
							var cErr error
							sure, cErr = ctx.Confirm(false, text)
							if cErr != nil {
								ctx.HandleError(cErr)
								return
							}
							if !sure {
								fmt.Fprintf(w, "you don't agree to detach udisk\n")
								return
							}
						}
						if *yes || sure {
							err := detachUdisk(ctx, false, disk.DiskId, w)
							if err != nil {
								ctx.HandleError(err)
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
				if errors.Is(err, errStopDeclined) {
					return
				}
				ctx.HandleError(err)
				return
			}
			resp, err := client.ReinstallUHostInstance(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			text := fmt.Sprintf("uhost[%s] is reinstalling OS", *req.UHostId)
			if *async {
				fmt.Fprintln(w, text)
			} else {
				ctx.PollerTo(w, describeUHostByID(ctx, *req.ProjectId, *req.Region, *req.Zone)).Spoll(resp.UHostId, text, []string{HOST_RUNNING, HOST_FAIL})
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
		return getUhostList(ctx, []string{HOST_RUNNING, HOST_STOPPED}, *req.ProjectId, *req.Region, *req.Zone)
	})
	cmd.MarkFlagRequired("uhost-id")
	return cmd
}

// detachUdisk detaches a udisk from its uhost, narrating progress to out.
// Copied self-contained from cmd/disk_compat.go (base.BizClient →
// cli.NewServiceClient), used by reinstall-os before reinstalling the OS.
func detachUdisk(ctx *cli.Context, async bool, udiskID string, out io.Writer) error {
	any, err := describeUdiskByID(ctx)(udiskID, nil)
	if err != nil {
		return err
	}
	if any == nil {
		return fmt.Errorf("udisk[%v] is not exist", any)
	}
	ins, ok := any.(*udisksdk.UDiskDataSet)
	if !ok {
		return fmt.Errorf("%#v convert to udisk failed", any)
	}
	client := cli.NewServiceClient(ctx, udisksdk.NewClient)
	req := client.NewDetachUDiskRequest()
	req.UHostId = sdk.String(ins.UHostId)
	req.UDiskId = sdk.String(udiskID)
	resp, err := client.DetachUDisk(req)
	if err != nil {
		return err
	}
	text := fmt.Sprintf("udisk[%s] is detaching from uhost[%s]", resp.UDiskId, resp.UHostId)
	if async {
		fmt.Fprintln(out, text)
	} else {
		ctx.PollerTo(out, describeUdiskByID(ctx)).Spoll(udiskID, text, []string{DISK_AVAILABLE, DISK_FAILED})
	}
	return nil
}
