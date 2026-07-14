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

// newResetPassword ucloud ulhost reset-password
func newResetPassword(ctx *cli.Context) *cobra.Command {
	var ulhostIDs *[]string
	client := cli.NewServiceClient(ctx, ucompsharesdk.NewClient)
	req := client.NewResetULHostInstancePasswordRequest()
	cmd := &cobra.Command{
		Use:   "reset-password",
		Short: "Reset the administrator password for the ULHost instances.",
		Long:  "Reset the administrator password for the ULHost instances.",
		Run: func(cmd *cobra.Command, args []string) {
			w := ctx.ProgressWriter()
			// Encode password to base64
			req.Password = sdk.String(base64.StdEncoding.EncodeToString([]byte(*req.Password)))
			results := []cli.OpResultRow{}
			for _, id := range *ulhostIDs {
				id = ctx.PickResourceID(id)
				req.ULHostId = &id
				resp, err := client.ResetULHostInstancePassword(req)
				if err != nil {
					ctx.HandleError(err)
					continue
				}
				fmt.Fprintf(w, "ulhost[%s] reset password\n", resp.ULHostId)
				results = append(results, cli.OpResultRow{ResourceID: resp.ULHostId, Action: "reset-password", Status: "OK"})
			}
			ctx.EmitResult(results...)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	ulhostIDs = flags.StringSlice("ulhost-id", nil, "Required. Resource IDs of the ulhosts to reset the administrator's password")
	req.Password = flags.String("password", "", "Required. New Password")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	command.SetCompletion(cmd, "ulhost-id", func() []string {
		return getULHostList(ctx, []string{HOST_RUNNING, HOST_STOPPED}, *req.ProjectId, *req.Region)
	})
	cmd.MarkFlagRequired("ulhost-id")
	cmd.MarkFlagRequired("password")
	return cmd
}
