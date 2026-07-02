package mysql

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/udb"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newResetPassword ucloud udb reset-password
func newResetPassword(ctx *cli.Context) *cobra.Command {
	var idNames []string
	client := cli.NewServiceClient(ctx, udb.NewClient)
	req := client.NewModifyUDBInstancePasswordRequest()
	cmd := &cobra.Command{
		Use:   "reset-password",
		Short: "Reset password of MySQL instances",
		Long:  "Reset password of MySQL instances",
		Run: func(c *cobra.Command, args []string) {
			w := ctx.ProgressWriter()
			results := []cli.OpResultRow{}
			for _, idname := range idNames {
				id := ctx.PickResourceID(idname)
				req.DBId = &id
				_, err := client.ModifyUDBInstancePassword(req)
				if err != nil {
					ctx.HandleError(err)
					continue
				}
				fmt.Fprintf(w, "udb[%s]'s password modified\n", idname)
				results = append(results, cli.OpResultRow{ResourceID: id, Action: "reset-password", Status: "PasswordReset"})
			}
			ctx.EmitResult(results...)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&idNames, "udb-id", nil, "Required. Resource ID of UDB instances to reset password")
	req.Password = flags.String("password", "", "Required. New password")
	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)

	cmd.MarkFlagRequired("udb-id")
	cmd.MarkFlagRequired("password")

	command.SetCompletion(cmd, "udb-id", func() []string {
		return getUDBIDList(ctx, nil, "", *req.ProjectId, *req.Region, *req.Zone)
	})

	return cmd
}
