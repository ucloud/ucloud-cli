package ulhost

import (
	"fmt"

	"github.com/spf13/cobra"

	ucompsharesdk "github.com/ucloud/ucloud-sdk-go/services/ucompshare"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newStop ucloud ulhost stop
func newStop(ctx *cli.Context) *cobra.Command {
	var ulhostIDs *[]string
	var async *bool
	client := cli.NewServiceClient(ctx, ucompsharesdk.NewClient)
	req := client.NewStopULHostInstanceRequest()
	cmd := &cobra.Command{
		Use:     "stop",
		Short:   "Shut down ULHost instance",
		Long:    "Shut down ULHost instance",
		Example: "ucloud ulhost stop --ulhost-id ulhost-xxx1,ulhost-xxx2",
		Run: func(cmd *cobra.Command, args []string) {
			w := ctx.ProgressWriter()
			results := []cli.OpResultRow{}
			for _, id := range *ulhostIDs {
				id = ctx.PickResourceID(id)
				req.ULHostId = &id
				resp, err := client.StopULHostInstance(req)
				if err != nil {
					ctx.HandleError(err)
					continue
				}
				text := fmt.Sprintf("ulhost[%v] is shutting down", resp.ULHostId)
				if *async {
					fmt.Fprintln(w, text)
				} else {
					ctx.PollerTo(w, describeULHostByID(ctx, *req.ProjectId, *req.Region)).Spoll(resp.ULHostId, text, []string{HOST_STOPPED, HOST_FAIL})
				}
				results = append(results, cli.OpResultRow{ResourceID: resp.ULHostId, Action: "stop", Status: "Stopping"})
			}
			ctx.EmitResult(results...)
		},
	}
	cmd.Flags().SortFlags = false
	ulhostIDs = cmd.Flags().StringSlice("ulhost-id", nil, "Required. ResourceIDs(ULHostIds) of the ulhost instances")
	req.ProjectId = cmd.Flags().String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = cmd.Flags().String("region", ctx.DefaultRegion(), "Optional. Assign region")
	async = cmd.Flags().Bool("async", false, "Optional. Do not wait for the long-running operation to finish.")
	command.SetCompletion(cmd, "ulhost-id", func() []string {
		return getULHostList(ctx, []string{HOST_RUNNING}, *req.ProjectId, *req.Region)
	})
	cmd.MarkFlagRequired("ulhost-id")

	return cmd
}
