package nlb

import (
	"fmt"

	"github.com/spf13/cobra"

	nlbsdk "github.com/ucloud/ucloud-sdk-go/services/nlb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newDelete implements `nlb delete`.
func newDelete(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, nlbsdk.NewClient)
	req := client.NewDeleteNetworkLoadBalancerRequest()

	var idNames []string
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete NLB instances by resource ID",
		Long:  "Delete one or more NLB instances by resource ID.",
		Run: func(c *cobra.Command, args []string) {
			ok, err := ctx.Confirm(yes, "Are you sure you want to delete the NLB instance(s)?")
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if !ok {
				return
			}
			req.ProjectId = sdk.String(ctx.PickResourceID(*req.ProjectId))
			results := []cli.OpResultRow{}
			for _, idName := range idNames {
				id := ctx.PickResourceID(idName)
				req.NLBId = sdk.String(id)
				if _, err := client.DeleteNetworkLoadBalancer(req); err != nil {
					ctx.HandleError(err)
					continue
				}
				fmt.Fprintf(ctx.ProgressWriter(), "nlb[%s] deleted\n", id)
				results = append(results, cli.OpResultRow{ResourceID: id, Action: "delete", Status: "Deleted"})
			}
			ctx.EmitResult(results...)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)

	flags.StringSliceVar(&idNames, resourceIDFlag, nil, "Required. Resource ID(s) of the NLB instances to delete.")
	flags.BoolVarP(&yes, "yes", "y", false, "Optional. Skip the confirmation prompt.")
	req.ReleaseEIP = flags.Bool("release-eip", false, "Optional. Release the EIP bound to the NLB when deleting.")

	cmd.MarkFlagRequired(resourceIDFlag)
	command.SetCompletion(cmd, resourceIDFlag, func() []string {
		return getAllNLBIDNames(ctx, derefStr(req.ProjectId), derefStr(req.Region))
	})

	return cmd
}
