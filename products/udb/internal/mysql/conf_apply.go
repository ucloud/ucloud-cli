package mysql

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/udb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newUDBConfApply ucloud udb conf apply
func newUDBConfApply(ctx *cli.Context) *cobra.Command {
	var confID string
	var udbIDs []string
	var restart, yes, async bool

	client := cli.NewServiceClient(ctx, udb.NewClient)
	req := client.NewChangeUDBParamGroupRequest()
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply configuration for UDB instances",
		Long:  "Apply configuration for UDB instances",
		Run: func(c *cobra.Command, args []string) {
			req.GroupId = sdk.String(ctx.PickResourceID(confID))
			w := ctx.ProgressWriter()
			results := []cli.OpResultRow{}
			for _, idname := range udbIDs {
				req.DBId = sdk.String(ctx.PickResourceID(idname))
				if restart {
					ok, err := ctx.Confirm(yes, fmt.Sprintf("udb[%s] is about to restart, do you want to continue?", idname))
					if err != nil {
						ctx.HandleError(err)
						continue
					}
					if !ok {
						continue
					}
				}
				_, err := client.ChangeUDBParamGroup(req)
				if err != nil {
					ctx.HandleError(err)
					continue
				}
				fmt.Fprintf(w, "conf[%s] has applied for udb[%s]\n", confID, idname)
				results = append(results, cli.OpResultRow{ResourceID: *req.DBId, Action: "apply", Status: "Applied"})
				if !restart {
					continue
				}
				restartReq := client.NewRestartUDBInstanceRequest()
				restartReq.Region = req.Region
				restartReq.Zone = req.Zone
				restartReq.ProjectId = req.ProjectId
				restartReq.DBId = req.DBId
				_, err = client.RestartUDBInstance(restartReq)
				if err != nil {
					ctx.HandleError(err)
					continue
				}
				if async {
					fmt.Fprintf(w, "udb[%s] is restarting\n", idname)
				} else {
					text := fmt.Sprintf("udb[%s] is restarting", idname)
					ctx.PollerTo(w, describeUdbByID(ctx)).Spoll(*req.DBId, text, []string{UDB_FAIL, UDB_RUNNING})
				}
			}
			ctx.EmitResult(results...)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringVar(&confID, "conf-id", "", "Required. ConfID of the configuration to be applied")
	flags.StringSliceVar(&udbIDs, "udb-id", nil, "Required. Resource ID of UDB instances to change configuration")
	flags.BoolVar(&restart, "restart-after-apply", true, "Optional. The new configuration will take effect after DB restarts")
	flags.BoolVarP(&yes, "yes", "y", false, "Optional. Do not prompt for confirmation")
	flags.BoolVarP(&async, "async", "a", false, "Optional. Do not wait for the long-running operation to finish.")
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)

	cmd.MarkFlagRequired("conf-id")
	cmd.MarkFlagRequired("udb-id")

	command.SetCompletion(cmd, "conf-id", func() []string {
		return getModifiableConfIDList(ctx, "", *req.ProjectId, *req.Region, *req.Zone)
	})
	command.SetCompletion(cmd, "udb-id", func() []string {
		return getUDBIDList(ctx, nil, "", *req.ProjectId, *req.Region, *req.Zone)
	})

	return cmd
}
