package ugn

import (
	"fmt"

	"github.com/spf13/cobra"

	ugnsdk "github.com/ucloud/ucloud-sdk-go/services/ugn"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newDelete ucloud ugn delete
func newDelete(ctx *cli.Context) *cobra.Command {
	idNames := []string{}
	client := cli.NewServiceClient(ctx, ugnsdk.NewClient)
	req := client.NewDelUGNRequest()
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete ugn instances",
		Long:  "Delete ugn instances",
		Run: func(c *cobra.Command, args []string) {
			ok, err := ctx.Confirm(yes, "Are you sure you want to delete the ugn instance(s)?")
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if !ok {
				return
			}
			results := []cli.OpResultRow{}
			for _, idname := range idNames {
				id := ctx.PickResourceID(idname)
				req.UGNID = sdk.String(id)
				_, err := client.DelUGN(req)
				if err != nil {
					ctx.HandleError(err)
					continue
				}
				fmt.Fprintf(ctx.ProgressWriter(), "ugn[%s] deleted\n", idname)
				results = append(results, cli.OpResultRow{ResourceID: id, Action: "delete", Status: "Deleted"})
			}
			ctx.EmitResult(results...)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&idNames, "ugn-id", nil, "Required. Resource ID of ugn instances to delete")
	flags.BoolVarP(&yes, "yes", "y", false, "Optional. Skip the confirmation prompt.")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Project-id, see 'ucloud project list'")

	cmd.MarkFlagRequired("ugn-id")
	command.SetCompletion(cmd, "ugn-id", func() []string {
		return getAllUGNIdNames(ctx, *req.ProjectId)
	})
	command.SetCompletion(cmd, "project-id", ctx.ProjectList)

	return cmd
}
