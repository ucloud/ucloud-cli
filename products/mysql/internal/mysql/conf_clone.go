package mysql

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/udb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newUDBConfClone ucloud udb conf clone
func newUDBConfClone(ctx *cli.Context) *cobra.Command {
	var srcConfID string
	client := cli.NewServiceClient(ctx, udb.NewClient)
	req := client.NewCreateUDBParamGroupRequest()
	cmd := &cobra.Command{
		Use:   "clone",
		Short: "Create configuration file by cloning existed configuration",
		Long:  "Create configuration file by cloning existed configuration",
		Run: func(c *cobra.Command, args []string) {
			id, err := strconv.Atoi(ctx.PickResourceID(srcConfID))
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if *req.DBTypeId == "" {
				confIns, err := getConfByID(ctx, id, *req.ProjectId, *req.Region, *req.Zone)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				req.DBTypeId = sdk.String(confIns.DBTypeId)
			}
			req.SrcGroupId = &id
			resp, err := client.CreateUDBParamGroup(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "conf[%d] created\n", resp.GroupId)
			ctx.EmitResult(cli.OpResultRow{ResourceID: strconv.Itoa(resp.GroupId), Action: "clone", Status: "Created"})
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.DBTypeId = flags.String("db-version", "", fmt.Sprintf("Required. Version of DB. Accept values:%s", strings.Join(dbVersionList, ", ")))
	req.GroupName = flags.String("name", "", "Required. Name of configuration. It's length should be between 6 and 63")
	req.Description = flags.String("description", " ", "Optional. Description of the configuration to clone")
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)
	flags.StringVar(&srcConfID, "src-conf-id", "", "Optional. The ConfID of source configuration which to be cloned from")

	command.SetFlagValues(cmd, "db-version", dbVersionList...)
	command.SetCompletion(cmd, "src-conf-id", func() []string {
		return getConfIDList(ctx, "sql", *req.ProjectId, *req.Region, *req.Zone)
	})

	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("src-conf-id")
	return cmd
}

func getConfByID(ctx *cli.Context, confID int, project, region, zone string) (*udb.UDBParamGroupSet, error) {
	client := cli.NewServiceClient(ctx, udb.NewClient)
	req := client.NewDescribeUDBParamGroupRequest()
	req.ProjectId = &project
	req.Region = &region
	req.Zone = &zone
	req.GroupId = &confID
	resp, err := client.DescribeUDBParamGroup(req)
	if err != nil {
		return nil, err
	}
	if len(resp.DataSet) != 1 {
		return nil, fmt.Errorf("conf-id[%d] may not exist", *req.GroupId)
	}
	return &resp.DataSet[0], nil
}
