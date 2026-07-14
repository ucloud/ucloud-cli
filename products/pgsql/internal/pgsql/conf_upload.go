package pgsql

import (
	"encoding/base64"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/internal/common"
	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newConfUpload ucloud pgsql conf upload
func newConfUpload(ctx *cli.Context) *cobra.Command {
	var file string
	client := newUPgSQLClient(ctx)
	req := client.NewUploadUPgSQLParamTemplateRequest()
	cmd := &cobra.Command{
		Use:   "upload",
		Short: "Create a UPgSQL parameter template by uploading a local config file",
		Long:  "Create a UPgSQL parameter template by uploading a local config file",
		Run: func(c *cobra.Command, args []string) {
			content, err := cli.ReadFile(file)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			req.Content = sdk.String(base64.StdEncoding.EncodeToString([]byte(content)))
			resp, err := client.UploadUPgSQLParamTemplate(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "conf[%d] uploaded\n", resp.GroupID)
			ctx.EmitResult(cli.OpResultRow{ResourceID: strconv.Itoa(resp.GroupID), Action: "upload", Status: "Uploaded"})
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringVar(&file, "conf-file", "", "Required. Path of the local configuration file")
	req.GroupName = flags.String("name", "", "Required. Name of the parameter template")
	req.DBVersion = flags.String("db-version", "", "Required. DB version. Options: postgresql-10.4, postgresql-13.4")
	req.Description = flags.String("description", "", "Optional. Description of the parameter template")
	req.ParamGroupType = flags.String("param-group-type", "", "Optional. Parameter group type")
	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)

	command.SetFlagValues(cmd, "db-version", pgsqlVersionList...)
	command.SetCompletion(cmd, "conf-file", func() []string {
		return common.GetFileList("")
	})

	cmd.MarkFlagRequired("conf-file")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("db-version")

	return cmd
}
