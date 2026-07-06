package onboarding

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/udb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newRestart implements `example restart`.
func newRestart(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, udb.NewClient)
	req := client.NewRestartUDBInstanceRequest()

	var ids []string
	var async bool

	cmd := &cobra.Command{
		Use:   "restart",
		Short: "Restart example instances",
		Long:  "Restart one or more example instances.",
		Run: func(c *cobra.Command, args []string) {
			w := ctx.ProgressWriter()
			results := []cli.OpResultRow{}
			for _, idName := range ids {
				id := ctx.PickResourceID(idName)
				req.DBId = sdk.String(id)
				if _, err := client.RestartUDBInstance(req); err != nil {
					ctx.HandleError(err)
					continue
				}
				text := fmt.Sprintf("%s[%s] is restarting", productName, id)
				if async {
					fmt.Fprintln(w, text)
				} else {
					ctx.PollerTo(w, describeByID(ctx)).Spoll(id, text, []string{stateRunning, stateFail})
				}
				results = append(results, cli.OpResultRow{ResourceID: id, Action: "restart", Status: "Restarting"})
			}
			ctx.EmitResult(results...)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&ids, resourceIDFlag, nil, "Required. Resource ID(s) of instances to restart.")
	flags.BoolVarP(&async, "async", "a", false, "Optional. Do not wait for the operation to finish.")
	ctx.BindCommonParams(cmd, req)

	cmd.MarkFlagRequired(resourceIDFlag)
	command.SetCompletion(cmd, resourceIDFlag, func() []string {
		return listResourceIDs(ctx, nil, derefStr(req.Region), derefStr(req.Zone), derefStr(req.ProjectId))
	})

	return cmd
}
