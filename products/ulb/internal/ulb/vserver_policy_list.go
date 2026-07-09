package ulb

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	ulbsdk "github.com/ucloud/ucloud-sdk-go/services/ulb"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newPolicyList returns ucloud ulb vserver policy list.
func newPolicyList(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, ulbsdk.NewClient)
	req := client.NewDescribeVServerRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List content forward policies of the VServer instance",
		Long:  "List content forward policies of the VServer instance",
		Run: func(c *cobra.Command, args []string) {
			ulbID := ctx.PickResourceID(*req.ULBId)
			vserverID := ctx.PickResourceID(*req.VServerId)
			vsList, err := getAllVServers(ctx, ulbID, vserverID, *req.ProjectId, *req.Region)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if len(vsList) == 1 {
				vs := vsList[0]
				list := []PolicyRow{}
				for _, p := range vs.PolicySet {
					row := PolicyRow{}
					row.ForwardMethod = p.Type
					row.Expression = p.Match
					row.PolicyID = p.PolicyId
					row.PolicyType = p.PolicyType
					nodes := []string{}
					for _, b := range p.BackendSet {
						nodes = append(nodes, fmt.Sprintf("%s|%s:%d|%s", b.BackendId, b.PrivateIP, b.Port, b.ResourceName))
					}
					row.Backends = strings.Join(nodes, ",")
					list = append(list, row)
				}
				ctx.PrintList(list)
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)
	req.ULBId = flags.String("ulb-id", "", "Required. Resource ID of ULB")
	req.VServerId = flags.String("vserver-id", "", "Required. Resource ID of VServer")

	command.SetCompletion(cmd, "ulb-id", func() []string {
		return getAllULBIDNames(ctx, *req.ProjectId, *req.Region)
	})
	command.SetCompletion(cmd, "vserver-id", func() []string {
		ulb := ctx.PickResourceID(*req.ULBId)
		return getAllVServerIDNames(ctx, ulb, *req.ProjectId, *req.Region)
	})
	cmd.MarkFlagRequired("ulb-id")
	cmd.MarkFlagRequired("vserver-id")
	return cmd
}
