package uhost

import (
	"fmt"

	"github.com/spf13/cobra"

	uhostsdk "github.com/ucloud/ucloud-sdk-go/services/uhost"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

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
					ctx.HandleError(err)
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
		return getUhostList(ctx, []string{HOST_RUNNING, HOST_STOPPED}, *req.ProjectId, *req.Region, *req.Zone)
	})
	cmd.MarkFlagRequired("uhost-id")
	cmd.MarkFlagRequired("password")
	return cmd
}
