package pgsql

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newUpdateName ucloud pgsql db update-name
func newUpdateName(ctx *cli.Context) *cobra.Command {
	client := newUPgSQLClient(ctx)
	req := client.NewUpdateUPgSQLAttributeRequest()
	cmd := &cobra.Command{
		Use:   "update-name",
		Short: "Update the name of a UPgSQL instance",
		Long:  "Update the name of a UPgSQL instance",
		Run: func(c *cobra.Command, args []string) {
			*req.InstanceID = ctx.PickResourceID(*req.InstanceID)
			_, err := client.UpdateUPgSQLAttribute(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			ctx.EmitResult(cli.OpResultRow{ResourceID: *req.InstanceID, Action: "update-name", Status: "Updated"})
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.InstanceID = flags.String("instance-id", "", "Required. Resource ID of the UPgSQL instance")
	req.Name = flags.String("name", "", "Required. New instance name")
	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)

	cmd.MarkFlagRequired("instance-id")
	cmd.MarkFlagRequired("name")
	command.SetCompletion(cmd, "instance-id", func() []string {
		return getUPgSQLIDList(ctx, req.GetProjectId(), req.GetRegion(), req.GetZone())
	})

	return cmd
}
