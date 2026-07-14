package ulhost

import (
	"encoding/base64"
	"fmt"

	"github.com/spf13/cobra"

	ucompsharesdk "github.com/ucloud/ucloud-sdk-go/services/ucompshare"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newReinstallOS ucloud ulhost reinstall-os
func newReinstallOS(ctx *cli.Context) *cobra.Command {
	var async *bool
	client := cli.NewServiceClient(ctx, ucompsharesdk.NewClient)
	req := client.NewReinstallULHostInstanceRequest()
	cmd := &cobra.Command{
		Use:   "reinstall-os",
		Short: "Reinstall the operating system of the ULHost instance",
		Long:  "Reinstall the operating system of the ULHost instance.",
		Run: func(cmd *cobra.Command, args []string) {
			w := ctx.ProgressWriter()
			req.ULHostId = sdk.String(ctx.PickResourceID(*req.ULHostId))
			req.ImageId = sdk.String(ctx.PickResourceID(*req.ImageId))
			// Encode password to base64
			req.Password = sdk.String(base64.StdEncoding.EncodeToString([]byte(*req.Password)))
			resp, err := client.ReinstallULHostInstance(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			text := fmt.Sprintf("ulhost[%s] is reinstalling OS", resp.ULHostId)
			if *async {
				fmt.Fprintln(w, text)
			} else {
				ctx.PollerTo(w, describeULHostByID(ctx, *req.ProjectId, *req.Region)).Spoll(resp.ULHostId, text, []string{HOST_RUNNING, HOST_FAIL})
			}
			ctx.EmitResult(cli.OpResultRow{ResourceID: resp.ULHostId, Action: "reinstall-os", Status: "Reinstalling"})
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.ULHostId = flags.String("ulhost-id", "", "Required. Resource ID of the ulhost to reinstall operating system")
	req.ImageId = flags.String("image-id", "", "Required. Resource ID of the image to install. See 'ucloud ulhost image list'")
	req.Password = flags.String("password", "", "Required. Password of the ulhost instance")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	async = flags.BoolP("async", "a", false, "Optional. Do not wait for the long-running operation to finish.")
	command.SetCompletion(cmd, "ulhost-id", func() []string {
		return getULHostList(ctx, []string{HOST_RUNNING, HOST_STOPPED}, *req.ProjectId, *req.Region)
	})
	cmd.MarkFlagRequired("ulhost-id")
	cmd.MarkFlagRequired("image-id")
	cmd.MarkFlagRequired("password")
	return cmd
}
