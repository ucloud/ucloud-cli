package ulb

import (
	"fmt"

	"github.com/spf13/cobra"

	ulbsdk "github.com/ucloud/ucloud-sdk-go/services/ulb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newBackendUpdate returns ucloud ulb vserver backend update.
func newBackendUpdate(ctx *cli.Context) *cobra.Command {
	var mode *string
	var weight *int
	backendIDs := []string{}
	client := cli.NewServiceClient(ctx, ulbsdk.NewClient)
	req := client.NewUpdateBackendAttributeRequest()
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update attributes of ULB backend nodes",
		Long:  "Update attributes of ULB backend nodes",
		Run: func(c *cobra.Command, args []string) {
			if *mode == "enable" {
				req.Enabled = sdk.Int(1)
			} else if *mode == "disable" {
				req.Enabled = sdk.Int(0)
			} else if *mode == "" {
				req.Enabled = nil
			} else {
				fmt.Fprintln(ctx.ProgressWriter(), "Error, backend-mode must be enable or disable")
				return
			}
			if *weight != -1 && (*weight < 0 || *weight > 100) {
				fmt.Fprintln(ctx.ProgressWriter(), "Error, weight must be between 0 and 100")
				return
			}
			if *weight != -1 {
				req.Weight = weight
			}

			if *req.Port == 0 {
				req.Port = nil
			}
			req.ULBId = sdk.String(ctx.PickResourceID(*req.ULBId))
			req.ProjectId = sdk.String(ctx.PickResourceID(*req.ProjectId))
			results := []cli.OpResultRow{}
			for _, bid := range backendIDs {
				id := ctx.PickResourceID(bid)
				req.BackendId = sdk.String(id)
				_, err := client.UpdateBackendAttribute(req)
				if err != nil {
					ctx.HandleError(err)
					continue
				}
				fmt.Fprintf(ctx.ProgressWriter(), "backend node[%s] updated\n", bid)
				results = append(results, cli.OpResultRow{ResourceID: id, Action: "update-backend", Status: "Updated"})
			}
			ctx.EmitResult(results...)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	req.ULBId = flags.String("ulb-id", "", "Required. Resource ID of ULB which the backend nodes belong to")
	flags.StringSliceVar(&backendIDs, "backend-id", nil, "Required. BackendID of backend nodes to update")
	req.Port = flags.Int("port", 0, "Optional. Port of your real server listening on backend nodes to update. Rnage [1,65535]")
	mode = flags.String("backend-mode", "", "Optional. Enable backend node or not. Accept values: enable, disable")
	weight = flags.Int("weight", -1, "Optional. effective for lb-method WeightRoundrobin. Rnage [0,100], -1 meaning no update")

	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)

	command.SetFlagValues(cmd, "backend-mode", "enable", "disable")
	command.SetCompletion(cmd, "ulb-id", func() []string {
		return getAllULBIDNames(ctx, *req.ProjectId, *req.Region)
	})
	command.SetCompletion(cmd, "backend-id", func() []string {
		return getAllBackendNodeIDNames(ctx, *req.ULBId, "", *req.ProjectId, *req.Region)
	})

	cmd.MarkFlagRequired("ulb-id")
	cmd.MarkFlagRequired("backend-id")

	return cmd
}
