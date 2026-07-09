package ulb

import (
	"fmt"

	"github.com/spf13/cobra"

	ulbsdk "github.com/ucloud/ucloud-sdk-go/services/ulb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newBackendAdd returns ucloud ulb vserver backend add.
func newBackendAdd(ctx *cli.Context) *cobra.Command {
	var enable *string
	var weight *int
	var ids []string
	client := cli.NewServiceClient(ctx, ulbsdk.NewClient)
	req := client.NewAllocateBackendRequest()
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add backend nodes for ULB Vserver instance",
		Long:  "Add backend nodes for ULB Vserver instance",
		Run: func(c *cobra.Command, args []string) {
			if *enable == "enable" {
				req.Enabled = sdk.Int(1)
			} else if *enable == "disable" {
				req.Enabled = sdk.Int(0)
			} else {
				fmt.Fprintln(ctx.ProgressWriter(), "Error, backend-mode must be enable or disable")
				return
			}
			if *weight < 0 || *weight > 100 {
				fmt.Fprintln(ctx.ProgressWriter(), "Error, weight must be between 0 and 100")
				return
			}
			req.ProjectId = sdk.String(ctx.PickResourceID(*req.ProjectId))
			req.ULBId = sdk.String(ctx.PickResourceID(*req.ULBId))
			req.VServerId = sdk.String(ctx.PickResourceID(*req.VServerId))
			results := []cli.OpResultRow{}
			for _, id := range ids {
				req.ResourceId = sdk.String(id)
				resp, err := client.AllocateBackend(req)
				if err != nil {
					ctx.HandleError(err)
					continue
				}
				fmt.Fprintf(ctx.ProgressWriter(), "backend node[%s] added, backend-id:%s\n", *req.ResourceId, resp.BackendId)
				results = append(results, cli.OpResultRow{ResourceID: resp.BackendId, Action: "add-backend", Status: "Added"})
			}
			ctx.EmitResult(results...)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	req.ULBId = flags.String("ulb-id", "", "Required. Resource ID of ULB which the backend nodes belong to")
	req.VServerId = flags.String("vserver-id", "", "Required. Resource ID of VServer which the backend nodes belong to")
	flags.StringSliceVar(&ids, "resource-id", nil, "Required. Resource ID of the backend nodes to add")
	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)
	req.ResourceType = flags.String("resource-type", "UHost", "Optional. Resource type of the backend node to add. Accept values: UHost,UPM,UDHost,UDocker")
	req.Port = flags.Int("port", 80, "Optional. The port of your real server on the backend node listening on")
	enable = flags.String("backend-mode", "enable", "Optional. Enable backend node or not. Accept values: enable, disable")
	weight = flags.Int("weight", 1, "Optional. effective for lb-method WeightRoundrobin. Rnage [0,100]")

	command.SetFlagValues(cmd, "resource-type", "Uhost", "UPM", "UDHost", "UDocker")
	command.SetFlagValues(cmd, "backend-mode", "enable", "disable")
	command.SetCompletion(cmd, "ulb-id", func() []string {
		return getAllULBIDNames(ctx, *req.ProjectId, *req.Region)
	})
	command.SetCompletion(cmd, "vserver-id", func() []string {
		ulbID := ctx.PickResourceID(*req.ULBId)
		return getAllVServerIDNames(ctx, ulbID, *req.ProjectId, *req.Region)
	})

	cmd.MarkFlagRequired("ulb-id")
	cmd.MarkFlagRequired("vserver-id")
	cmd.MarkFlagRequired("resource-id")
	return cmd
}
