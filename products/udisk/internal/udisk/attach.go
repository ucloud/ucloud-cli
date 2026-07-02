package udisk

import (
	"fmt"

	"github.com/spf13/cobra"

	udisksdk "github.com/ucloud/ucloud-sdk-go/services/udisk"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

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
					ctx.PollerTo(w, describeUdiskByID(ctx)).Spoll(resp.UDiskId, text, []string{DISK_INUSE, DISK_FAILED})
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
		return getDiskList(ctx, []string{DISK_AVAILABLE}, *req.ProjectId, *req.Region, *req.Zone)
	})
	command.SetCompletion(cmd, "uhost-id", func() []string {
		return getUhostList(ctx, []string{HOST_RUNNING, HOST_STOPPED}, *req.ProjectId, *req.Region, *req.Zone)
	})

	cmd.MarkFlagRequired("uhost-id")
	cmd.MarkFlagRequired("udisk-id")

	return cmd
}
