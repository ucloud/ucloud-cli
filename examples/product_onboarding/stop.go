package onboarding

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/udb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newStop implements `example stop`.
func newStop(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, udb.NewClient)
	req := client.NewStopUDBInstanceRequest()

	var ids []string
	var async bool

	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop example instances",
		Long:  "Stop one or more running example instances.",
		Run: func(c *cobra.Command, args []string) {
			w := ctx.ProgressWriter()
			results := []cli.OpResultRow{}
			for _, idName := range ids {
				id := ctx.PickResourceID(idName)
				req.DBId = sdk.String(id)
				if _, err := client.StopUDBInstance(req); err != nil {
					ctx.HandleError(err)
					continue
				}
				text := fmt.Sprintf("%s[%s] is stopping", productName, id)
				if async {
					fmt.Fprintln(w, text)
				} else {
					ctx.PollerTo(w, describeByID(ctx)).Spoll(id, text, []string{stateShutoff, stateFail})
				}
				results = append(results, cli.OpResultRow{ResourceID: id, Action: "stop", Status: "Stopping"})
			}
			ctx.EmitResult(results...)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&ids, resourceIDFlag, nil, "Required. Resource ID(s) of instances to stop.")
	req.ForceToKill = flags.Bool("force", false, "Optional. Force-stop the instance(s).")
	flags.BoolVarP(&async, "async", "a", false, "Optional. Do not wait for the operation to finish.")
	ctx.BindCommonParams(cmd, req)

	cmd.MarkFlagRequired(resourceIDFlag)
	command.SetFlagValues(cmd, "force", "true", "false")
	command.SetCompletion(cmd, resourceIDFlag, func() []string {
		// Only running instances are stoppable.
		return listResourceIDs(ctx, []string{stateRunning}, derefStr(req.Region), derefStr(req.Zone), derefStr(req.ProjectId))
	})

	return cmd
}
