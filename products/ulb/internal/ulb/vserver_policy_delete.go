package ulb

import (
	"fmt"

	"github.com/spf13/cobra"

	ulbsdk "github.com/ucloud/ucloud-sdk-go/services/ulb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newPolicyDelete returns ucloud ulb vserver policy delete.
func newPolicyDelete(ctx *cli.Context) *cobra.Command {
	policyIDs := []string{}
	client := cli.NewServiceClient(ctx, ulbsdk.NewClient)
	req := client.NewDeletePolicyRequest()
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete content forward policies of ULB VServer",
		Long:  "Delete content forward policies of ULB VServer",
		Run: func(c *cobra.Command, args []string) {
			req.ProjectId = sdk.String(ctx.PickResourceID(*req.ProjectId))
			results := []cli.OpResultRow{}
			for _, p := range policyIDs {
				req.PolicyId = sdk.String(p)
				_, err := client.DeletePolicy(req)
				if err != nil {
					ctx.HandleError(err)
					continue
				}
				fmt.Fprintf(ctx.ProgressWriter(), "policy[%s] deleted\n", p)
				results = append(results, cli.OpResultRow{ResourceID: p, Action: "delete-policy", Status: "Deleted"})
			}
			ctx.EmitResult(results...)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)

	flags.StringSliceVar(&policyIDs, "policy-id", nil, "Required. PolicyID of policies to delete")
	req.VServerId = flags.String("vserver-id", "", "Optional. Resource ID of VServer")

	cmd.MarkFlagRequired("policy-id")

	return cmd
}
