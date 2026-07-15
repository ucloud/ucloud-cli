package nlb

import (
	"fmt"

	"github.com/spf13/cobra"

	nlbsdk "github.com/ucloud/ucloud-sdk-go/services/nlb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newListenerDelete implements `nlb listener delete`.
func newListenerDelete(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, nlbsdk.NewClient)
	req := client.NewDeleteNLBListenerRequest()

	var nlbID string
	var listenerIDs []string
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete NLB listeners by resource ID",
		Long:  "Delete one or more listeners of an NLB instance.",
		Run: func(c *cobra.Command, args []string) {
			ok, err := ctx.Confirm(yes, "Are you sure you want to delete the listener(s)?")
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if !ok {
				return
			}
			req.ProjectId = sdk.String(ctx.PickResourceID(*req.ProjectId))
			req.NLBId = sdk.String(ctx.PickResourceID(nlbID))
			results := []cli.OpResultRow{}
			for _, idName := range listenerIDs {
				id := ctx.PickResourceID(idName)
				req.ListenerId = sdk.String(id)
				if _, err := client.DeleteNLBListener(req); err != nil {
					ctx.HandleError(err)
					continue
				}
				fmt.Fprintf(ctx.ProgressWriter(), "nlb-listener[%s] deleted\n", id)
				results = append(results, cli.OpResultRow{ResourceID: id, Action: "delete-listener", Status: "Deleted"})
			}
			ctx.EmitResult(results...)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)

	flags.StringVar(&nlbID, resourceIDFlag, "", "Required. Resource ID of the NLB instance.")
	flags.StringSliceVar(&listenerIDs, "listener-id", nil, "Required. Resource ID(s) of the listeners to delete.")
	flags.BoolVarP(&yes, "yes", "y", false, "Optional. Skip the confirmation prompt.")

	cmd.MarkFlagRequired(resourceIDFlag)
	cmd.MarkFlagRequired("listener-id")
	command.SetCompletion(cmd, resourceIDFlag, func() []string {
		return getAllNLBIDNames(ctx, derefStr(req.ProjectId), derefStr(req.Region))
	})
	command.SetCompletion(cmd, "listener-id", func() []string {
		return getAllListenerIDNames(ctx, nlbID, derefStr(req.ProjectId), derefStr(req.Region))
	})

	return cmd
}
