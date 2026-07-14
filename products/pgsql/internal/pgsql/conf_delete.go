package pgsql

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newConfDelete ucloud pgsql conf delete
func newConfDelete(ctx *cli.Context) *cobra.Command {
	var confID string
	client := newUPgSQLClient(ctx)
	req := client.NewDeleteUPgSQLParamTemplateRequest()
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a UPgSQL parameter template",
		Long:  "Delete a UPgSQL parameter template",
		Run: func(c *cobra.Command, args []string) {
			id, err := strconv.Atoi(ctx.PickResourceID(confID))
			if err != nil {
				ctx.HandleError(fmt.Errorf("invalid conf-id: %w", err))
				return
			}
			req.GroupID = sdk.Int(id)
			_, err = client.DeleteUPgSQLParamTemplate(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "conf[%s] deleted\n", confID)
			ctx.EmitResult(cli.OpResultRow{ResourceID: strconv.Itoa(id), Action: "delete", Status: "Deleted"})
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringVar(&confID, "conf-id", "", "Required. Group ID of the parameter template to delete")
	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)

	cmd.MarkFlagRequired("conf-id")
	command.SetCompletion(cmd, "conf-id", func() []string {
		return listParamTemplateIDNames(ctx, req.GetProjectId(), req.GetRegion(), req.GetZone())
	})

	return cmd
}
