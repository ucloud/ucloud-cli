package uhost

import (
	"fmt"

	"github.com/spf13/cobra"

	uhostsdk "github.com/ucloud/ucloud-sdk-go/services/uhost"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newReboot ucloud uhost restart
func newReboot(ctx *cli.Context) *cobra.Command {
	var uhostIDs *[]string
	var async *bool
	client := cli.NewServiceClient(ctx, uhostsdk.NewClient)
	req := client.NewRebootUHostInstanceRequest()
	cmd := &cobra.Command{
		Use:     "restart",
		Short:   "Restart uhost instance",
		Long:    "Restart uhost instance",
		Example: "ucloud uhost restart --uhost-id uhost-xxx1,uhost-xxx2",
		Run: func(cmd *cobra.Command, args []string) {
			w := ctx.ProgressWriter()
			for _, id := range *uhostIDs {
				id = ctx.PickResourceID(id)
				req.UHostId = &id
				resp, err := client.RebootUHostInstance(req)
				if err != nil {
					ctx.HandleError(err)
				} else {
					text := fmt.Sprintf("uhost[%v] is restarting", resp.UHostId)
					if *async {
						fmt.Fprintln(w, text)
					} else {
						ctx.PollerTo(w, describeUHostByID(ctx, *req.ProjectId, *req.Region, *req.Zone)).Spoll(resp.UHostId, text, []string{HOST_RUNNING, HOST_FAIL})
					}
				}
			}
		},
	}
	cmd.Flags().SortFlags = false
	uhostIDs = cmd.Flags().StringSlice("uhost-id", nil, "Required. ResourceIDs(UHostIds) of the uhost instance")
	req.ProjectId = cmd.Flags().String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = cmd.Flags().String("region", ctx.DefaultRegion(), "Optional. Assign region")
	req.Zone = cmd.Flags().String("zone", "", "Optional. Assign availability zone")
	req.DiskPassword = cmd.Flags().String("disk-password", "", "Optional. Encrypted disk password")
	async = cmd.Flags().Bool("async", false, "Optional. Do not wait for the long-running operation to finish.")
	command.SetCompletion(cmd, "uhost-id", func() []string {
		return getUhostList(ctx, []string{HOST_FAIL, HOST_RUNNING, HOST_STOPPED}, *req.ProjectId, *req.Region, *req.Zone)
	})
	cmd.MarkFlagRequired("uhost-id")
	return cmd
}
