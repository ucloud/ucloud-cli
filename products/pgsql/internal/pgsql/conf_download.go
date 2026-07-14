package pgsql

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newConfDownload ucloud pgsql conf download
func newConfDownload(ctx *cli.Context) *cobra.Command {
	var confID string
	client := newUPgSQLClient(ctx)
	req := client.NewDownloadUPgSQLParamTemplateRequest()
	cmd := &cobra.Command{
		Use:   "download",
		Short: "Download a UPgSQL parameter template (base64 content)",
		Long:  "Download a UPgSQL parameter template and print its base64 content to stdout",
		Run: func(c *cobra.Command, args []string) {
			id, err := strconv.Atoi(ctx.PickResourceID(confID))
			if err != nil {
				ctx.HandleError(err)
				return
			}
			req.GroupID = sdk.Int(id)
			resp, err := client.DownloadUPgSQLParamTemplate(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprint(ctx.Out(), resp.Content)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringVar(&confID, "conf-id", "", "Required. Group ID of the parameter template to download")
	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)

	cmd.MarkFlagRequired("conf-id")
	command.SetCompletion(cmd, "conf-id", func() []string {
		return listParamTemplateIDNames(ctx, req.GetProjectId(), req.GetRegion(), req.GetZone())
	})

	return cmd
}
