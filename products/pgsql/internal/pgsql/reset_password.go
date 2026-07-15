package pgsql

import (
	"github.com/spf13/cobra"

	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newResetPassword ucloud pgsql db reset-password
func newResetPassword(ctx *cli.Context) *cobra.Command {
	var idNames []string
	client := newUPgSQLClient(ctx)
	req := client.NewUpdateUPgSQLPasswordRequest()
	cmd := &cobra.Command{
		Use:   "reset-password",
		Short: "Reset the admin password of UPgSQL instances",
		Long:  "Reset the admin password of UPgSQL instances",
		Run: func(c *cobra.Command, args []string) {
			results := []cli.OpResultRow{}
			for _, idname := range idNames {
				id := ctx.PickResourceID(idname)
				req.InstanceID = sdk.String(id)
				_, err := client.UpdateUPgSQLPassword(req)
				if err != nil {
					ctx.HandleError(err)
					continue
				}
				results = append(results, cli.OpResultRow{ResourceID: id, Action: "reset-password", Status: "PasswordReset"})
			}
			ctx.EmitResult(results...)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&idNames, "instance-id", nil, "Required. Resource ID of UPgSQL instances to reset password")
	req.Password = flags.String("password", "", "Required. New password")
	req.Name = flags.String("name", "", "Optional. Database user name, default root")
	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)

	cmd.MarkFlagRequired("instance-id")
	cmd.MarkFlagRequired("password")
	command.SetCompletion(cmd, "instance-id", func() []string {
		return getUPgSQLIDList(ctx, req.GetProjectId(), req.GetRegion(), req.GetZone())
	})

	return cmd
}
