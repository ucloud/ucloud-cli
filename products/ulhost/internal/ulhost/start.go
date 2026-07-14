package ulhost

import (
	"fmt"

	"github.com/spf13/cobra"

	ucompsharesdk "github.com/ucloud/ucloud-sdk-go/services/ucompshare"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newStart ucloud ulhost start
func newStart(ctx *cli.Context) *cobra.Command {
	var async *bool
	var ulhostIDs *[]string
	client := cli.NewServiceClient(ctx, ucompsharesdk.NewClient)
	req := client.NewStartULHostInstanceRequest()
	cmd := &cobra.Command{
		Use:     "start",
		Short:   "Start ULHost instance",
		Long:    "Start ULHost instance",
		Example: "ucloud ulhost start --ulhost-id ulhost-xxx1,ulhost-xxx2",
		Run: func(cmd *cobra.Command, args []string) {
			w := ctx.ProgressWriter()
			results := []cli.OpResultRow{}
			for _, id := range *ulhostIDs {
				id := ctx.PickResourceID(id)
				req.ULHostId = &id
				resp, err := client.StartULHostInstance(req)
				if err != nil {
					ctx.HandleError(err)
					continue
				}
				text := fmt.Sprintf("ulhost[%v] is starting", resp.ULHostId)
				if *async {
					fmt.Fprintln(w, text)
				} else {
					ctx.PollerTo(w, describeULHostByID(ctx, *req.ProjectId, *req.Region)).Spoll(resp.ULHostId, text, []string{HOST_RUNNING, HOST_FAIL})
				}
				results = append(results, cli.OpResultRow{ResourceID: resp.ULHostId, Action: "start", Status: "Starting"})
			}
			ctx.EmitResult(results...)
		},
	}
	cmd.Flags().SortFlags = false
	ulhostIDs = cmd.Flags().StringSlice("ulhost-id", nil, "Required. ResourceIDs(ULHostIds) of the ulhost instance")
	req.ProjectId = cmd.Flags().String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = cmd.Flags().String("region", ctx.DefaultRegion(), "Optional. Assign region")
	async = cmd.Flags().Bool("async", false, "Optional. Do not wait for the long-running operation to finish.")
	command.SetCompletion(cmd, "ulhost-id", func() []string {
		return getULHostList(ctx, []string{HOST_STOPPED}, *req.ProjectId, *req.Region)
	})
	cmd.MarkFlagRequired("ulhost-id")
	return cmd
}
