package mysql

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/udb"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newUDBConfDownload ucloud udb conf download
func newUDBConfDownload(ctx *cli.Context) *cobra.Command {
	var confID string
	client := cli.NewServiceClient(ctx, udb.NewClient)
	req := client.NewExtractUDBParamGroupRequest()
	cmd := &cobra.Command{
		Use:   "download",
		Short: "Download UDB configuration",
		Long:  "Download UDB configuration",
		Run: func(c *cobra.Command, args []string) {
			id, err := strconv.Atoi(ctx.PickResourceID(confID))
			if err != nil {
				ctx.HandleError(err)
				return
			}

			req.GroupId = &id
			resp, err := client.ExtractUDBParamGroup(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprint(ctx.Out(), resp.Content)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringVar(&confID, "conf-id", "", "Required. ConfID of configuration to download")
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)

	cmd.MarkFlagRequired("conf-id")

	command.SetCompletion(cmd, "conf-id", func() []string {
		return getConfIDList(ctx, "sql", *req.ProjectId, *req.Region, *req.Zone)
	})

	return cmd
}
