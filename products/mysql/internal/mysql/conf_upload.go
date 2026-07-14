package mysql

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/udb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/internal/common"
	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

var udbSubtypeMap = map[string]int{
	"unknow":               0,
	"Shardsvr-MMAPv1":      1,
	"Shardsvr-WiredTiger":  2,
	"Configsvr-MMAPv1":     3,
	"Configsvr-WiredTiger": 4,
	"Mongos":               5,
	"Mysql":                10,
	"Postgresql":           20,
}

var subtypeList = []string{"Shardsvr-MMAPv1", "Shardsvr-WiredTiger", "Configsvr-MMAPv1", "Configsvr-WiredTiger", "Mongos", "Mysql", "Postgresql"}

// newUDBConfUpload ucloud udb conf upload
func newUDBConfUpload(ctx *cli.Context) *cobra.Command {
	var file string
	client := cli.NewServiceClient(ctx, udb.NewClient)
	req := client.NewUploadUDBParamGroupRequest()
	cmd := &cobra.Command{
		Use:   "upload",
		Short: "Create configuration file by uploading local DB configuration file",
		Long:  "Create configuration file by uploading local DB configuration file",
		Run: func(c *cobra.Command, args []string) {
			content, err := cli.ReadFile(file)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if l := len(*req.GroupName); l < 6 || l > 63 {
				ctx.HandleError(fmt.Errorf("length of name shoud be between 6 and 63"))
				return
			}
			req.Content = sdk.String(base64.StdEncoding.EncodeToString([]byte(content)))
			req.ParamGroupTypeId = sdk.Int(10)
			resp, err := client.UploadUDBParamGroup(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "conf[%d] uploaded\n", resp.GroupId)
			ctx.EmitResult(cli.OpResultRow{ResourceID: strconv.Itoa(resp.GroupId), Action: "upload", Status: "Uploaded"})
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringVar(&file, "conf-file", "", "Required. Path of local configuration file")
	req.DBTypeId = flags.String("db-version", "", fmt.Sprintf("Required. Version of DB. Accept values:%s", strings.Join(dbVersionList, ", ")))
	req.GroupName = flags.String("name", "", "Required. Name of configuration. It's length should be between 6 and 63")
	req.Description = flags.String("description", " ", "Optional. Description of the configuration to clone")
	// flags.StringVar(&subtype, "db-type", "", fmt.Sprintf("Optional. DB type. Accept values: %s", strings.Join(subtypeList, ", ")))
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)

	cmd.MarkFlagRequired("conf-file")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("db-version")
	// cmd.MarkFlagRequired("db-type")

	command.SetFlagValues(cmd, "db-version", dbVersionList...)
	// command.SetFlagValues(cmd, "db-type", subtypeList...)
	command.SetCompletion(cmd, "conf-file", func() []string {
		return common.GetFileList("")
	})
	return cmd
}
