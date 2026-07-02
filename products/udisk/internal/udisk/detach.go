package udisk

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	udisksdk "github.com/ucloud/ucloud-sdk-go/services/udisk"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

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
					ctx.HandleError(err)
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
		return getDiskList(ctx, []string{DISK_INUSE}, *req.ProjectId, *req.Region, *req.Zone)
	})

	cmd.MarkFlagRequired("udisk-id")
	return cmd
}

// DetachUdisk detaches a udisk from its uhost, narrating progress to out.
// Ported from cmd/disk.go's detachUdisk (base.BizClient → cli.NewServiceClient);
// exported so the restore command and (legacy) callers share one copy.
func DetachUdisk(ctx *cli.Context, async bool, udiskID string, out io.Writer) error {
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
