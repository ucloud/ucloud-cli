package upfs

import (
	"fmt"

	"github.com/spf13/cobra"

	upfssdk "github.com/ucloud/ucloud-sdk-go/services/upfs"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newCreate ucloud upfs create
func newCreate(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, upfssdk.NewClient)
	req := client.NewCreateUPFSVolumeRequest()
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a UPFS volume",
		Long:  "Create a UPFS volume",
		Run: func(cmd *cobra.Command, args []string) {
			w := ctx.ProgressWriter()
			resp, err := client.CreateUPFSVolume(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}

			text := fmt.Sprintf("upfs:%v created", resp.VolumeId)
			fmt.Fprintln(w, text)
			ctx.EmitResult(cli.OpResultRow{ResourceID: resp.VolumeId, Action: "create", Status: "Created"})
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.VolumeName = flags.String("name", "", "Required. Name of the UPFS volume to create")
	req.Size = flags.Int("size-gb", 500, "Required. Size of the UPFS volume. Unit: GB, must be a multiple of 100, minimum 500")
	req.ProtocolType = flags.String("protocol-type", "POSIX", "Optional. Protocol type, currently only supports POSIX")
	req.ChargeType = flags.String("charge-type", "Dynamic", "Optional. 'Year', pay yearly; 'Month', pay monthly; 'Dynamic', pay hourly")
	req.Quantity = flags.Int("quantity", 1, "Optional. The duration of the instance. N years/months")
	req.Tag = flags.String("group", "Default", "Optional. Business group")
	req.Remark = flags.String("remark", "", "Optional. Remark")

	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)

	command.SetFlagValues(cmd, "charge-type", "Month", "Year", "Dynamic", "Trial")
	command.SetFlagValues(cmd, "protocol-type", "POSIX")

	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("size-gb")

	return cmd
}
