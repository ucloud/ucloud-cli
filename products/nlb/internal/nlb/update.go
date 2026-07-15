package nlb

import (
	"fmt"

	"github.com/spf13/cobra"

	nlbsdk "github.com/ucloud/ucloud-sdk-go/services/nlb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newUpdate implements `nlb update`.
func newUpdate(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, nlbsdk.NewClient)
	req := client.NewUpdateNetworkLoadBalancerAttributeRequest()

	var idNames []string
	var name, remark, group string

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update NLB instance attributes",
		Long:  "Update the name, remark or business group of one or more NLB instances.",
		Run: func(c *cobra.Command, args []string) {
			if name == "" && remark == "" && group == "" {
				ctx.HandleError(fmt.Errorf("nothing to update: set at least one of --name/--remark/--group"))
				return
			}
			req.ProjectId = sdk.String(ctx.PickResourceID(*req.ProjectId))
			if name != "" {
				req.Name = &name
			}
			if remark != "" {
				req.Remark = &remark
			}
			if group != "" {
				req.Tag = &group
			}
			results := []cli.OpResultRow{}
			for _, idName := range idNames {
				id := ctx.PickResourceID(idName)
				req.NLBId = sdk.String(id)
				if _, err := client.UpdateNetworkLoadBalancerAttribute(req); err != nil {
					ctx.HandleError(err)
					continue
				}
				fmt.Fprintf(ctx.ProgressWriter(), "nlb[%s] updated\n", id)
				results = append(results, cli.OpResultRow{ResourceID: id, Action: "update", Status: "Updated"})
			}
			ctx.EmitResult(results...)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)

	flags.StringSliceVar(&idNames, resourceIDFlag, nil, "Required. Resource ID(s) of the NLB instances to update.")
	flags.StringVar(&name, "name", "", "Optional. New NLB instance name.")
	flags.StringVar(&remark, "remark", "", "Optional. New remark.")
	flags.StringVar(&group, "group", "", "Optional. New business group.")

	cmd.MarkFlagRequired(resourceIDFlag)
	command.SetCompletion(cmd, resourceIDFlag, func() []string {
		return getAllNLBIDNames(ctx, derefStr(req.ProjectId), derefStr(req.Region))
	})

	return cmd
}
