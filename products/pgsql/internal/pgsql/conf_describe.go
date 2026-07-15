package pgsql

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/upgsql"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newConfDescribe ucloud pgsql conf describe
func newConfDescribe(ctx *cli.Context) *cobra.Command {
	var confID string
	client := newUPgSQLClient(ctx)
	req := client.NewGetUPgSQLParamTemplateRequest()
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Display details of a UPgSQL parameter template",
		Long:  "Display details of a UPgSQL parameter template",
		Run: func(c *cobra.Command, args []string) {
			id, err := strconv.Atoi(ctx.PickResourceID(confID))
			if err != nil {
				ctx.HandleError(fmt.Errorf("invalid conf-id: %w", err))
				return
			}
			req.GroupID = sdk.Int(id)
			resp, err := client.GetUPgSQLParamTemplate(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}

			// Template metadata comes from ListUPgSQLParamTemplate (GetUPgSQLParamTemplate
			// returns only the param list). Fetch and filter by GroupID.
			templates, err := listParamTemplates(ctx, req.GetProjectId(), req.GetRegion(), req.GetZone())
			if err != nil {
				ctx.HandleError(err)
				return
			}
			var group upgsql.TemplateGroup
			for _, t := range templates {
				if t.GroupID == id {
					group = t
					break
				}
			}
			attrs := []cli.DescribeRow{
				{Attribute: "GroupID", Content: strconv.Itoa(group.GroupID)},
				{Attribute: "GroupName", Content: group.GroupName},
				{Attribute: "DBVersion", Content: group.DBVersion},
				{Attribute: "Description", Content: group.Description},
				{Attribute: "Modifiable", Content: strconv.FormatBool(group.Modifiable)},
			}
			fmt.Fprintln(ctx.ProgressWriter(), "Attributes:")
			ctx.PrintList(attrs)

			params := []PgsqlConfParamRow{}
			for _, p := range resp.Data {
				if p.Key == "" {
					continue
				}
				params = append(params, PgsqlConfParamRow{
					Key:        p.Key,
					Value:      p.Value,
					Modifiable: p.Modifiable,
				})
			}
			fmt.Fprintln(ctx.ProgressWriter(), "\nParameters:")
			ctx.PrintList(params)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringVar(&confID, "conf-id", "", "Required. Group ID of the parameter template to describe")
	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)

	cmd.MarkFlagRequired("conf-id")
	command.SetCompletion(cmd, "conf-id", func() []string {
		return listParamTemplateIDNames(ctx, req.GetProjectId(), req.GetRegion(), req.GetZone())
	})

	return cmd
}
