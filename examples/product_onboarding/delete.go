package onboarding

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/udb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newDelete implements `example delete`.
//
// Platform APIs exercised: cli.NewServiceClient, ctx.BindCommonParams,
// ctx.Confirm (the destructive-op guard), the --yes/-y pattern,
// ctx.PickResourceID, ctx.ProgressWriter, ctx.EmitResult, ctx.HandleError,
// command.SetCompletion.
func newDelete(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, udb.NewClient)
	req := client.NewDeleteUDBInstanceRequest()

	var ids []string
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete example instances",
		Long:  "Delete one or more example instances by resource ID.",
		Run: func(c *cobra.Command, args []string) {
			// Destructive: gate on confirmation unless --yes was passed.
			ok, err := ctx.Confirm(yes, "Are you sure you want to delete the instance(s)?")
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if !ok {
				return
			}
			w := ctx.ProgressWriter()
			results := []cli.OpResultRow{}
			for _, idName := range ids {
				id := ctx.PickResourceID(idName)
				req.DBId = sdk.String(id)
				if _, err := client.DeleteUDBInstance(req); err != nil {
					ctx.HandleError(err)
					continue
				}
				fmt.Fprintf(w, "%s[%s] deleted\n", productName, id)
				results = append(results, cli.OpResultRow{ResourceID: id, Action: "delete", Status: "Deleted"})
			}
			ctx.EmitResult(results...)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&ids, resourceIDFlag, nil, "Required. Resource ID(s) of instances to delete.")
	flags.BoolVarP(&yes, "yes", "y", false, "Optional. Skip the confirmation prompt.")
	ctx.BindCommonParams(cmd, req)

	cmd.MarkFlagRequired(resourceIDFlag)
	command.SetCompletion(cmd, resourceIDFlag, func() []string {
		return listResourceIDs(ctx, nil, derefStr(req.Region), derefStr(req.Zone), derefStr(req.ProjectId))
	})

	return cmd
}
