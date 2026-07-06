package uhost

import (
	"github.com/spf13/cobra"

	uhostsdk "github.com/ucloud/ucloud-sdk-go/services/uhost"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newStop ucloud uhost stop
func newStop(ctx *cli.Context) *cobra.Command {
	var uhostIDs *[]string
	var async *bool
	client := cli.NewServiceClient(ctx, uhostsdk.NewClient)
	req := client.NewStopUHostInstanceRequest()
	cmd := &cobra.Command{
		Use:     "stop",
		Short:   "Shut down uhost instance",
		Long:    "Shut down uhost instance",
		Example: "ucloud uhost stop --uhost-id uhost-xxx1,uhost-xxx2",
		Run: func(cmd *cobra.Command, args []string) {
			for _, id := range *uhostIDs {
				id = ctx.PickResourceID(id)
				req.UHostId = &id
				stopUhostIns(ctx, client, req, *async)
			}
		},
	}
	cmd.Flags().SortFlags = false
	uhostIDs = cmd.Flags().StringSlice("uhost-id", nil, "Required. ResourceIDs(UHostIds) of the uhost instances")
	req.ProjectId = cmd.Flags().String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = cmd.Flags().String("region", ctx.DefaultRegion(), "Optional. Assign region")
	req.Zone = cmd.Flags().String("zone", "", "Optional. Assign availability zone")
	async = cmd.Flags().Bool("async", false, "Optional. Do not wait for the long-running operation to finish.")
	command.SetCompletion(cmd, "uhost-id", func() []string {
		return getUhostList(ctx, []string{HOST_RUNNING}, *req.ProjectId, *req.Region, *req.Zone)
	})
	cmd.MarkFlagRequired("uhost-id")

	return cmd
}
