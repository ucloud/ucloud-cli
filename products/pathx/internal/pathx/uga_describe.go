package pathx

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	ppathx "github.com/ucloud/ucloud-sdk-go/private/services/pathx"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newUGADescribe ucloud pathx uga describe
func newUGADescribe(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, ppathx.NewClient)
	req := client.NewDescribeUGAInstanceRequest()
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Display detail informations about uga instances",
		Long:  "Display detail informations about uga instances",
		Run: func(c *cobra.Command, args []string) {
			*req.UGAId = ctx.PickResourceID(*req.UGAId)
			resp, err := client.DescribeUGAInstance(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if len(resp.UGAList) != 1 {
				ctx.HandleError(fmt.Errorf("uga[%s] may not exist", *req.UGAId))
				return
			}
			ins := resp.UGAList[0]
			list := []describeRow{
				{Attribute: "ResourceID", Content: ins.UGAId},
				{Attribute: "UGAName", Content: ins.UGAName},
				{Attribute: "Origin", Content: fmt.Sprintf("%s%s", ins.Domain, strings.Join(ins.IPList, ","))},
				{Attribute: "CName", Content: ins.CName},
				{Attribute: "AcceleratedPath", Content: getUpathStr(ins.UPathSet)},
				{Attribute: "OutIP", Content: getOutIPStr(ins.OutPublicIpList)},
				{Attribute: "Port", Content: getPortStr(ins.TaskSet)},
			}
			ctx.PrintList(list)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.UGAId = flags.String("uga-id", "", "Required. Resource ID of uga instance")
	ctx.BindProjectID(cmd, req)
	cmd.MarkFlagRequired("uga-id")
	ctx.SetCompletion(cmd, "uga-id", func() []string {
		return getUGAIDList(ctx, *req.ProjectId)
	})
	return cmd
}
