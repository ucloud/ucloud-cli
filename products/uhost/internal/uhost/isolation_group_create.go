package uhost

import (
	"fmt"
	"regexp"

	"github.com/spf13/cobra"

	uhostsdk "github.com/ucloud/ucloud-sdk-go/services/uhost"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newIsolationCreate ucloud uhost isolation-group create
func newIsolationCreate(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, uhostsdk.NewClient)
	req := client.NewCreateIsolationGroupRequest()
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create isolation group instance",
		Long:  "Create isolation group instance",
		Run: func(c *cobra.Command, args []string) {
			re := regexp.MustCompile(REGEXP_NAME)
			if !re.Match([]byte(*req.GroupName)) {
				ctx.LogError(fmt.Sprintf("group-name %s is invalid! Length 1~63, only English,Chinese,number and '-_.' are allowed", *req.GroupName))
				return
			}
			resp, err := client.CreateIsolationGroup(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "isolation group %s created\n", resp.GroupId)
			ctx.EmitResult(cli.OpResultRow{ResourceID: resp.GroupId, Action: "create", Status: "Created"})
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	req.GroupName = flags.String("group-name", "", "Required. Name of isolation group. Length 1~63, only English,Chinese,number and '-_.' are allowed")
	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)
	req.Remark = flags.String("remark", "", "Optional. Remark ok isolation group")

	cmd.MarkFlagRequired("group-name")
	return cmd
}
