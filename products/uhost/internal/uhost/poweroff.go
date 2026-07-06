package uhost

import (
	"fmt"

	"github.com/spf13/cobra"

	uhostsdk "github.com/ucloud/ucloud-sdk-go/services/uhost"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newPoweroff ucloud uhost poweroff
func newPoweroff(ctx *cli.Context) *cobra.Command {
	var yes *bool
	var uhostIDs *[]string
	client := cli.NewServiceClient(ctx, uhostsdk.NewClient)
	req := client.NewPoweroffUHostInstanceRequest()
	cmd := &cobra.Command{
		Use:     "poweroff",
		Short:   "Analog power off Uhost instnace",
		Long:    "Analog power off Uhost instnace",
		Example: "ucloud uhost poweroff --uhost-id uhost-xxx1,uhost-xxx2",
		Run: func(cmd *cobra.Command, args []string) {
			w := ctx.ProgressWriter()
			confirmText := "Danger, it may affect data integrity. Are you sure you want to poweroff this uhost?"
			if len(*uhostIDs) > 1 {
				confirmText = "Danger, it may affect data integrity. Are you sure you want to poweroff those uhosts?"
			}
			ok, err := ctx.Confirm(*yes, confirmText)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if !ok {
				return
			}
			for _, id := range *uhostIDs {
				id = ctx.PickResourceID(id)
				req.UHostId = &id
				resp, err := client.PoweroffUHostInstance(req)
				if err != nil {
					ctx.HandleError(err)
				} else {
					fmt.Fprintf(w, "uhost[%v] is power off\n", resp.UHostId)
				}
			}
		},
	}
	cmd.Flags().SortFlags = false
	uhostIDs = cmd.Flags().StringSlice("uhost-id", nil, "ResourceIDs(UHostIds) of the uhost instance")
	req.ProjectId = cmd.Flags().String("project-id", ctx.DefaultProjectID(), "Assign project-id")
	req.Region = cmd.Flags().String("region", ctx.DefaultRegion(), "Assign region")
	req.Zone = cmd.Flags().String("zone", "", "Assign availability zone")
	yes = cmd.Flags().BoolP("yes", "y", false, "Optional. Do not prompt for confirmation.")

	command.SetCompletion(cmd, "uhost-id", func() []string {
		return getUhostList(ctx, []string{HOST_FAIL, HOST_RUNNING, HOST_STOPPED}, *req.ProjectId, *req.Region, *req.Zone)
	})
	cmd.MarkFlagRequired("uhost-id")

	return cmd
}
