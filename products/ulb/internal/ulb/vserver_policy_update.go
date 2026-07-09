package ulb

import (
	"fmt"

	"github.com/spf13/cobra"

	ulbsdk "github.com/ucloud/ucloud-sdk-go/services/ulb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newPolicyUpdate returns ucloud ulb vserver policy update.
func newPolicyUpdate(ctx *cli.Context) *cobra.Command {
	policyIDs := []string{}
	backendIDs := []string{}
	addBackendIDs := []string{}
	removeBackendIDs := []string{}
	client := cli.NewServiceClient(ctx, ulbsdk.NewClient)
	req := client.NewUpdatePolicyRequest()
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update content forward policies of ULB VServer",
		Long:  "Update content forward policies ULB VServer",
		Run: func(c *cobra.Command, args []string) {
			req.ProjectId = sdk.String(ctx.PickResourceID(*req.ProjectId))
			req.ULBId = sdk.String(ctx.PickResourceID(*req.ULBId))
			req.VServerId = sdk.String(ctx.PickResourceID(*req.VServerId))

			vsList, err := getAllVServers(ctx, *req.ULBId, *req.VServerId, *req.ProjectId, *req.Region)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			vs := vsList[0]

			results := []cli.OpResultRow{}
			for _, policyID := range policyIDs {
				var policy *ulbsdk.ULBPolicySet
				for _, p := range vs.PolicySet {
					if p.PolicyId == policyID {
						policy = &p
						break
					}
				}
				if policy == nil {
					fmt.Fprintf(ctx.ProgressWriter(), "policy[%s] not found\n", policyID)
					continue
				}
				req.PolicyId = sdk.String(policyID)
				if *req.Type == "" {
					req.Type = sdk.String(policy.Type)
				} else if *req.Type != "Domain" && *req.Type != "Path" {
					fmt.Fprintln(ctx.ProgressWriter(), "Error, forward-method must be Domain or Path")
					continue
				}
				if *req.Match == "" {
					req.Match = sdk.String(policy.Match)
				}
				backendIDMap := map[string]bool{}
				if backendIDs == nil {
					for _, b := range policy.BackendSet {
						backendIDMap[b.BackendId] = true
					}
				} else {
					for _, bid := range backendIDs {
						backendIDMap[ctx.PickResourceID(bid)] = true
					}
				}
				for _, bid := range addBackendIDs {
					backendIDMap[ctx.PickResourceID(bid)] = true
				}
				for _, bid := range removeBackendIDs {
					backendIDMap[ctx.PickResourceID(bid)] = false
				}
				req.BackendId = nil
				for bid, ok := range backendIDMap {
					if ok {
						req.BackendId = append(req.BackendId, bid)
					}
				}
				resp, err := client.UpdatePolicy(req)
				if err != nil {
					ctx.HandleError(err)
					continue
				}
				fmt.Fprintf(ctx.ProgressWriter(), "policy[%s] updated\n", resp.PolicyId)
				results = append(results, cli.OpResultRow{ResourceID: resp.PolicyId, Action: "update-policy", Status: "Updated"})
			}
			ctx.EmitResult(results...)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)
	req.ULBId = flags.String("ulb-id", "", "Required. Resource ID of ULB")
	req.VServerId = flags.String("vserver-id", "", "Required. Resource ID of VServer")
	flags.StringSliceVar(&policyIDs, "policy-id", nil, "Required. PolicyID of policies to update")
	flags.StringSliceVar(&backendIDs, "backend-id", nil, "Optional. BackendID of backend nodes. If assign this flag, it will rewrite all backend nodes of the policy")
	flags.StringSliceVar(&addBackendIDs, "add-backend-id", nil, "Optional. BackendID of backend nodes. Add backend nodes to the policy")
	flags.StringSliceVar(&removeBackendIDs, "remove-backend-id", nil, "Optional. BackendID of backend nodes. Remove those backend nodes from the policy")
	req.Type = flags.String("forward-method", "", "Optional. Forward method of policy, accept values:Domain and Path")
	req.Match = flags.String("expression", "", "Optional. Expression of domain or path, such as \"www.[123].demo.com\" or \"/path/img/*.jpg\"")

	cmd.MarkFlagRequired("ulb-id")
	cmd.MarkFlagRequired("vserver-id")
	cmd.MarkFlagRequired("policy-id")

	command.SetFlagValues(cmd, "forward-method", "Domain", "Path")
	command.SetCompletion(cmd, "ulb-id", func() []string {
		project := ctx.PickResourceID(*req.ProjectId)
		return getAllULBIDNames(ctx, project, *req.Region)
	})
	command.SetCompletion(cmd, "vserver-id", func() []string {
		return getAllVServerIDNames(ctx, *req.ULBId, *req.ProjectId, *req.Region)
	})
	command.SetCompletion(cmd, "backend-id", func() []string {
		return getAllBackendNodeIDNames(ctx, *req.ULBId, *req.VServerId, *req.ProjectId, *req.Region)
	})
	command.SetCompletion(cmd, "add-backend-id", func() []string {
		return getAllBackendNodeIDNames(ctx, *req.ULBId, *req.VServerId, *req.ProjectId, *req.Region)
	})
	command.SetCompletion(cmd, "remove-backend-id", func() []string {
		return getAllBackendNodeIDNames(ctx, *req.ULBId, *req.VServerId, *req.ProjectId, *req.Region)
	})

	return cmd
}
