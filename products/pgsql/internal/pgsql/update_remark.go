package pgsql

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/upgsql"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newUpdateRemark ucloud pgsql db update-remark
func newUpdateRemark(ctx *cli.Context) *cobra.Command {
	client := newUPgSQLClient(ctx)
	req := client.NewUpdateUPgSQLAttributeRequest()
	cmd := &cobra.Command{
		Use:   "update-remark",
		Short: "Update the remark of a UPgSQL instance",
		Long:  "Update the remark of a UPgSQL instance",
		Run: func(c *cobra.Command, args []string) {
			*req.InstanceID = ctx.PickResourceID(*req.InstanceID)
			// UpdateUPgSQLAttribute requires Name to be present even when only the
			// remark changes — omitting it returns RetCode 230 "Params [Name] not
			// available". Fetch the current name and re-send it alongside Remark.
			any, err := describePgsqlByID(ctx)(*req.InstanceID, nil)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			ins, ok := any.(*upgsql.UDBInstance)
			if !ok {
				ctx.HandleError(fmt.Errorf("fetch pgsql[%s] instance", *req.InstanceID))
				return
			}
			req.Name = sdk.String(ins.Name)
			_, err = client.UpdateUPgSQLAttribute(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			ctx.EmitResult(cli.OpResultRow{ResourceID: *req.InstanceID, Action: "update-remark", Status: "Updated"})
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.InstanceID = flags.String("instance-id", "", "Required. Resource ID of the UPgSQL instance")
	req.Remark = flags.String("remark", "", "Required. New remark")
	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)

	cmd.MarkFlagRequired("instance-id")
	cmd.MarkFlagRequired("remark")
	command.SetCompletion(cmd, "instance-id", func() []string {
		return getUPgSQLIDList(ctx, req.GetProjectId(), req.GetRegion(), req.GetZone())
	})

	return cmd
}
