package ufs

import (
	"fmt"

	"github.com/spf13/cobra"

	ufssdk "github.com/ucloud/ucloud-sdk-go/services/ufs"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newCreate ucloud ufs create
func newCreate(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, ufssdk.NewClient)
	req := client.NewCreateUFSVolumeRequest()
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a UFS volume",
		Long:  "Create a UFS volume",
		Run: func(cmd *cobra.Command, args []string) {
			w := ctx.ProgressWriter()
			resp, err := client.CreateUFSVolume(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}

			text := fmt.Sprintf("ufs:%v created", resp.VolumeId)
			fmt.Fprintln(w, text)
			ctx.EmitResult(cli.OpResultRow{ResourceID: resp.VolumeId, Action: "create", Status: "Created"})
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.VolumeName = flags.String("name", "", "Required. Name of the UFS volume to create")
	req.Size = flags.Int("size-gb", 100, "Required. Size of the UFS volume. Unit: GB")
	req.StorageType = flags.String("storage-type", "Basic", "Optional. Storage type: 'Basic' (capacity) or 'Advanced' (performance)")
	req.ProtocolType = flags.String("protocol-type", "NFS", "Optional. Protocol type: 'NFS' or 'SMB'")
	req.ChargeType = flags.String("charge-type", "Dynamic", "Optional. 'Year', pay yearly; 'Month', pay monthly; 'Dynamic', pay hourly")
	req.Quantity = flags.Int("quantity", 1, "Optional. The duration of the instance. N years/months")
	req.Tag = flags.String("group", "Default", "Optional. Business group")
	req.Remark = flags.String("remark", "", "Optional. Remark")

	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)

	command.SetFlagValues(cmd, "charge-type", "Month", "Year", "Dynamic", "Trial")
	command.SetFlagValues(cmd, "storage-type", "Basic", "Advanced")
	command.SetFlagValues(cmd, "protocol-type", "NFS", "SMB")

	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("size-gb")

	return cmd
}
