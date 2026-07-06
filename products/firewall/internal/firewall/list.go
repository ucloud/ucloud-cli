package firewall

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/unet"

	"github.com/ucloud/ucloud-cli/internal/common"
	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newList ucloud firewall list
func newList(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, unet.NewClient)
	req := client.NewDescribeFirewallRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List extranet firewall",
		Long:  `List extranet firewall`,
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := client.DescribeFirewall(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			list := []FirewallRow{}
			for _, fw := range resp.DataSet {
				row := FirewallRow{}
				row.ResourceID = fw.FWId
				row.FirewallName = fw.Name
				row.Group = fw.Tag
				row.RuleAmount = len(fw.Rule)
				row.BoundResourceAmount = fw.ResourceCount
				row.CreationTime = common.FormatDate(fw.CreateTime)
				if fw.Remark != "" {
					row.FirewallName += "\nremark:" + fw.Remark + "\n"
				}
				for _, r := range fw.Rule {
					rule := fmt.Sprintf("%s|%s|%s|%s|%s", r.ProtocolType, r.DstPort, r.SrcIP, r.RuleAction, r.Priority)
					row.Rule += rule + "\n"
				}
				list = append(list, row)
			}
			ctx.PrintList(list)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Region, see 'ucloud region'")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Project-id, see 'ucloud project list'")
	req.FWId = flags.String("firewall-id", "", "Optional. The Rsource ID of firewall. Return all firewalls by default.")
	req.ResourceType = flags.String("bound-resource-type", "", "Optional. The type of resource bound on the firewall")
	req.ResourceId = flags.String("bound-resource-id", "", "Optional. The resource ID of resource bound on the firewall")
	req.Offset = flags.Int("offset", 0, "Optional. Offset")
	req.Limit = flags.Int("limit", 50, "Optional. Limit")
	return cmd
}
