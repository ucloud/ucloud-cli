package nlb

import (
	"fmt"

	"github.com/spf13/cobra"

	nlbsdk "github.com/ucloud/ucloud-sdk-go/services/nlb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newTargetRemove implements `nlb target remove`.
func newTargetRemove(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, nlbsdk.NewClient)
	req := client.NewRemoveNLBTargetsRequest()

	var nlbID, listenerID string
	var targetIDs []string
	var yes bool

	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove backend targets from an NLB listener",
		Long:  "Remove one or more backend service nodes from an NLB listener.",
		Run: func(c *cobra.Command, args []string) {
			ok, err := ctx.Confirm(yes, "Are you sure you want to remove the target(s)?")
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if !ok {
				return
			}
			req.ProjectId = sdk.String(ctx.PickResourceID(*req.ProjectId))
			req.NLBId = sdk.String(ctx.PickResourceID(nlbID))
			req.ListenerId = sdk.String(ctx.PickResourceID(listenerID))
			ids := make([]string, 0, len(targetIDs))
			for _, idName := range targetIDs {
				ids = append(ids, ctx.PickResourceID(idName))
			}
			req.Ids = ids

			if _, err := client.RemoveNLBTargets(req); err != nil {
				ctx.HandleError(err)
				return
			}
			results := make([]cli.OpResultRow, 0, len(ids))
			for _, id := range ids {
				fmt.Fprintf(ctx.ProgressWriter(), "nlb-target[%s] removed\n", id)
				results = append(results, cli.OpResultRow{ResourceID: id, Action: "remove-target", Status: "Removed"})
			}
			ctx.EmitResult(results...)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)

	flags.StringVar(&nlbID, resourceIDFlag, "", "Required. Resource ID of the NLB instance.")
	flags.StringVar(&listenerID, "listener-id", "", "Required. Resource ID of the listener.")
	flags.StringSliceVar(&targetIDs, "target-id", nil, "Required. Target ID(s) to remove (max 40 per request).")
	flags.BoolVarP(&yes, "yes", "y", false, "Optional. Skip the confirmation prompt.")

	cmd.MarkFlagRequired(resourceIDFlag)
	cmd.MarkFlagRequired("listener-id")
	cmd.MarkFlagRequired("target-id")
	command.SetCompletion(cmd, resourceIDFlag, func() []string {
		return getAllNLBIDNames(ctx, derefStr(req.ProjectId), derefStr(req.Region))
	})
	command.SetCompletion(cmd, "listener-id", func() []string {
		return getAllListenerIDNames(ctx, nlbID, derefStr(req.ProjectId), derefStr(req.Region))
	})
	command.SetCompletion(cmd, "target-id", func() []string {
		return getAllTargetIDNames(ctx, nlbID, listenerID, derefStr(req.ProjectId), derefStr(req.Region))
	})

	return cmd
}
