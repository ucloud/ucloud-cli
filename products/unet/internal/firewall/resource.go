package firewall

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/unet"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newResource ucloud firewall resource
func newResource(ctx *cli.Context) *cobra.Command {
	fwID := ""
	client := cli.NewServiceClient(ctx, unet.NewClient)
	req := client.NewDescribeFirewallResourceRequest()
	cmd := &cobra.Command{
		Use:   "resource",
		Short: "List resources that has been applied the firewall",
		Long:  "List resources that has been applied the firewall",
		Run: func(c *cobra.Command, args []string) {
			req.FWId = sdk.String(ctx.PickResourceID(fwID))
			resp, err := client.DescribeFirewallResource(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			list := []FirewallResourceRow{}
			for _, rs := range resp.ResourceSet {
				row := FirewallResourceRow{}
				row.ResourceName = rs.Name
				row.ResourceID = rs.ResourceID
				row.ResourceType = rs.ResourceType
				row.IntranetIP = rs.PrivateIP
				row.Group = rs.Tag
				row.Remark = rs.Remark
				list = append(list, row)
			}
			ctx.PrintList(list)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringVar(&fwID, "fw-id", "", "Required. Resource ID of firewall")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Region, see 'ucloud region'")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Project-id, see 'ucloud project list'")
	req.Offset = flags.Int("offset", 0, "Optional. Offset")
	req.Limit = flags.Int("limit", 50, "Optional. Limit")

	command.SetCompletion(cmd, "fw-id", func() []string {
		return getFirewallIDNames(ctx, *req.ProjectId, *req.Region)
	})

	cmd.MarkFlagRequired("fw-id")

	return cmd
}
