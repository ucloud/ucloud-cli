package ulhost

import (
	"fmt"

	"github.com/spf13/cobra"

	ucompsharesdk "github.com/ucloud/ucloud-sdk-go/services/ucompshare"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newPoweroff ucloud ulhost poweroff
func newPoweroff(ctx *cli.Context) *cobra.Command {
	var yes *bool
	var ulhostIDs *[]string
	client := cli.NewServiceClient(ctx, ucompsharesdk.NewClient)
	req := client.NewPoweroffULHostInstanceRequest()
	cmd := &cobra.Command{
		Use:     "poweroff",
		Short:   "Analog power off ULHost instance",
		Long:    "Analog power off ULHost instance. Danger, it may affect data integrity.",
		Example: "ucloud ulhost poweroff --ulhost-id ulhost-xxx1,ulhost-xxx2",
		Run: func(cmd *cobra.Command, args []string) {
			w := ctx.ProgressWriter()
			confirmText := "Danger, it may affect data integrity. Are you sure you want to poweroff this ulhost?"
			if len(*ulhostIDs) > 1 {
				confirmText = "Danger, it may affect data integrity. Are you sure you want to poweroff those ulhosts?"
			}
			ok, err := ctx.Confirm(*yes, confirmText)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if !ok {
				return
			}
			for _, id := range *ulhostIDs {
				id = ctx.PickResourceID(id)
				req.ULHostId = &id
				resp, err := client.PoweroffULHostInstance(req)
				if err != nil {
					ctx.HandleError(err)
				} else {
					fmt.Fprintf(w, "ulhost[%v] is power off\n", resp.ULHostId)
				}
			}
		},
	}
	cmd.Flags().SortFlags = false
	ulhostIDs = cmd.Flags().StringSlice("ulhost-id", nil, "ResourceIDs(ULHostIds) of the ulhost instance")
	req.ProjectId = cmd.Flags().String("project-id", ctx.DefaultProjectID(), "Assign project-id")
	req.Region = cmd.Flags().String("region", ctx.DefaultRegion(), "Assign region")
	yes = cmd.Flags().BoolP("yes", "y", false, "Optional. Do not prompt for confirmation.")

	command.SetCompletion(cmd, "ulhost-id", func() []string {
		return getULHostList(ctx, []string{HOST_FAIL, HOST_RUNNING, HOST_STOPPED}, *req.ProjectId, *req.Region)
	})
	cmd.MarkFlagRequired("ulhost-id")

	return cmd
}
