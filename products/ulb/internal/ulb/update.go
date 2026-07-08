package ulb

import (
	"fmt"

	"github.com/spf13/cobra"

	ulbsdk "github.com/ucloud/ucloud-sdk-go/services/ulb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newUpdate returns ucloud ulb update.
func newUpdate(ctx *cli.Context) *cobra.Command {
	var name, group, remark string
	idNames := []string{}
	client := cli.NewServiceClient(ctx, ulbsdk.NewClient)
	req := client.NewUpdateULBAttributeRequest()
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update ULB instance",
		Long:  "Update ULB instance",
		Run: func(c *cobra.Command, args []string) {
			req.ProjectId = sdk.String(ctx.PickResourceID(*req.ProjectId))
			results := []cli.OpResultRow{}
			for _, idname := range idNames {
				id := ctx.PickResourceID(idname)
				req.ULBId = sdk.String(id)
				if name == "" && group == "" && remark == "" {
					fmt.Fprintln(ctx.ProgressWriter(), "Error, name, remark and group can't be all empty")
					return
				}
				if name != "" {
					req.Name = &name
				}
				if group != "" {
					req.Tag = &group
				}
				if remark != "" {
					req.Remark = &remark
				}
				_, err := client.UpdateULBAttribute(req)
				if err != nil {
					ctx.HandleError(err)
					continue
				}
				fmt.Fprintf(ctx.ProgressWriter(), "ulb[%s] updated\n", *req.ULBId)
				results = append(results, cli.OpResultRow{ResourceID: id, Action: "update", Status: "Updated"})
			}
			ctx.EmitResult(results...)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)
	flags.StringSliceVar(&idNames, "ulb-id", nil, "Required. Resource ID of ULB instances to update")
	flags.StringVar(&name, "name", "", "Optional, Name of ULB instance")
	flags.StringVar(&remark, "remark", "", "Optional, Remark of ULB instance")
	flags.StringVar(&group, "group", "", "Optional, Business group of ULB instance")

	command.SetCompletion(cmd, "ulb-id", func() []string {
		return getAllULBIDNames(ctx, *req.ProjectId, *req.Region)
	})

	cmd.MarkFlagRequired("ulb-id")

	return cmd
}
