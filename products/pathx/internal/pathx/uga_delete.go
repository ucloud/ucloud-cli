package pathx

import (
	"fmt"

	"github.com/spf13/cobra"

	ppathx "github.com/ucloud/ucloud-sdk-go/private/services/pathx"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newUGADelete ucloud pathx uga delete
func newUGADelete(ctx *cli.Context) *cobra.Command {
	idNames := []string{}
	client := cli.NewServiceClient(ctx, ppathx.NewClient)
	req := client.NewDeleteUGAInstanceRequest()
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete uga instances",
		Long:  "Delete uga instances",
		Run: func(c *cobra.Command, args []string) {
			results := []cli.OpResultRow{}
			for _, idname := range idNames {
				id := ctx.PickResourceID(idname)
				req.UGAId = &id
				_, err := client.DeleteUGAInstance(req)
				if err != nil {
					ctx.HandleError(err)
					continue
				}
				fmt.Fprintf(ctx.ProgressWriter(), "uga[%s] deleted\n", id)
				results = append(results, cli.OpResultRow{ResourceID: id, Action: "delete", Status: "Deleted"})
			}
			ctx.EmitResult(results...)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	ctx.BindProjectID(cmd, req)
	flags.StringSliceVar(&idNames, "uga-id", nil, "Required. Resource ID of uga instances to delete. Multiple resource ids separated by comma")
	cmd.MarkFlagRequired("uga-id")
	ctx.SetCompletion(cmd, "uga-id", func() []string {
		return getUGAIDList(ctx, *req.ProjectId)
	})
	return cmd
}
