package ulb

import (
	"fmt"

	"github.com/spf13/cobra"

	ulbsdk "github.com/ucloud/ucloud-sdk-go/services/ulb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newPolicyAdd returns ucloud ulb vserver policy add.
func newPolicyAdd(ctx *cli.Context) *cobra.Command {
	backendIDs := []string{}
	client := cli.NewServiceClient(ctx, ulbsdk.NewClient)
	req := client.NewCreatePolicyRequest()
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add content forward policy for VServer",
		Long:  "Add content forward policy for VServer",
		Run: func(c *cobra.Command, args []string) {
			if *req.Type != "Domain" && *req.Type != "Path" {
				fmt.Fprintln(ctx.ProgressWriter(), "Error, forward method must be Domain or Path")
				return
			}
			req.ProjectId = sdk.String(ctx.PickResourceID(*req.ProjectId))
			req.ULBId = sdk.String(ctx.PickResourceID(*req.ULBId))
			req.VServerId = sdk.String(ctx.PickResourceID(*req.VServerId))
			for _, idname := range backendIDs {
				req.BackendId = append(req.BackendId, ctx.PickResourceID(idname))
			}
			resp, err := client.CreatePolicy(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "policy[%s] created\n", resp.PolicyId)
			ctx.EmitResult(cli.OpResultRow{ResourceID: resp.PolicyId, Action: "create-policy", Status: "Created"})
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.ULBId = flags.String("ulb-id", "", "Required. Resource ID of ULB")
	req.VServerId = flags.String("vserver-id", "", "Required. Resource ID of VServer")
	flags.StringSliceVar(&backendIDs, "backend-id", nil, "Required. BackendID of the VServer's backend nodes")
	req.Type = flags.String("forward-method", "", "Required. Forward method, accept values:Domain and Path; Both forwarding methods can be described by using regular expressions or wildcards")
	req.Match = flags.String("expression", "", "Required. Expression of domain or path, such as \"www.[123].demo.com\" or \"/path/img/*.jpg\"")
	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)

	command.SetFlagValues(cmd, "forward-method", "Domain", "Path")
	command.SetCompletion(cmd, "ulb-id", func() []string {
		return getAllULBIDNames(ctx, *req.ProjectId, *req.Region)
	})
	command.SetCompletion(cmd, "vserver-id", func() []string {
		return getAllVServerIDNames(ctx, *req.ULBId, *req.ProjectId, *req.Region)
	})
	command.SetCompletion(cmd, "backend-id", func() []string {
		return getAllBackendNodeIDNames(ctx, *req.ULBId, *req.VServerId, *req.ProjectId, *req.Region)
	})

	cmd.MarkFlagRequired("ulb-id")
	cmd.MarkFlagRequired("vserver-id")
	cmd.MarkFlagRequired("backend-id")
	cmd.MarkFlagRequired("forward-method")
	cmd.MarkFlagRequired("expression")

	return cmd
}
