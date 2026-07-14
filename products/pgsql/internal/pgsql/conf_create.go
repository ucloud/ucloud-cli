package pgsql

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newConfCreate ucloud pgsql conf create
func newConfCreate(ctx *cli.Context) *cobra.Command {
	var srcConfID string
	client := newUPgSQLClient(ctx)
	req := client.NewCreateUPgSQLParamTemplateRequest()
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a UPgSQL parameter template from a base template",
		Long:  "Create a UPgSQL parameter template from a base template",
		Run: func(c *cobra.Command, args []string) {
			id, err := strconv.Atoi(ctx.PickResourceID(srcConfID))
			if err != nil {
				ctx.HandleError(fmt.Errorf("invalid src-conf-id: %w", err))
				return
			}
			req.SrcGroupID = sdk.Int(id)
			resp, err := client.CreateUPgSQLParamTemplate(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "conf[%d] created\n", resp.GroupID)
			ctx.EmitResult(cli.OpResultRow{ResourceID: strconv.Itoa(resp.GroupID), Action: "create", Status: "Created"})
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.GroupName = flags.String("name", "", "Required. Name of the parameter template")
	flags.StringVar(&srcConfID, "src-conf-id", "", "Required. Group ID of the base template to clone from")
	req.DBVersion = flags.String("db-version", "", "Required. DB version. Options: postgresql-10.4, postgresql-13.4")
	req.Description = flags.String("description", "", "Optional. Description of the parameter template")
	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)

	command.SetFlagValues(cmd, "db-version", pgsqlVersionList...)
	command.SetCompletion(cmd, "src-conf-id", func() []string {
		return listParamTemplateIDNames(ctx, req.GetProjectId(), req.GetRegion(), req.GetZone())
	})

	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("src-conf-id")
	cmd.MarkFlagRequired("db-version")

	return cmd
}
