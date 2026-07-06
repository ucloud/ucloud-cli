package firewall

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/unet"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newCopy ucloud firewall copy
func newCopy(ctx *cli.Context) *cobra.Command {
	srcFirewall := ""
	srcRegion := ""
	client := cli.NewServiceClient(ctx, unet.NewClient)
	req := client.NewCreateFirewallRequest()
	cmd := &cobra.Command{
		Use:     "copy",
		Short:   "Copy firewall",
		Long:    "Copy firewall",
		Example: "ucloud firewall copy --src-fw firewall-xxx --target-region cn-bj2 --name test",
		Run: func(c *cobra.Command, args []string) {
			fwID := ctx.PickResourceID(srcFirewall)
			firewall, err := getFirewall(ctx, fwID, *req.ProjectId, srcRegion)

			if err != nil {
				ctx.HandleError(err)
				return
			}
			req.Tag = sdk.String(firewall.Tag)
			req.Remark = sdk.String(firewall.Remark)
			for _, r := range firewall.Rule {
				rstr := fmt.Sprintf("%s|%s|%s|%s|%s", r.ProtocolType, r.DstPort, r.SrcIP, r.RuleAction, r.Priority)
				req.Rule = append(req.Rule, rstr)
			}
			resp, err := client.CreateFirewall(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "firewall[%s] created from %s\n", resp.FWId, srcFirewall)
			ctx.EmitResult(cli.OpResultRow{ResourceID: resp.FWId, Action: "copy", Status: "Created"})
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	flags.StringVar(&srcFirewall, "src-fw", "", "Required. ResourceID or name of source firewall")
	req.Name = flags.String("name", "", "Required. Name of new firewall")
	flags.StringVar(&srcRegion, "region", ctx.DefaultRegion(), "Optional. Current region, used to fetch source firewall")
	req.Region = flags.String("target-region", ctx.DefaultRegion(), "Optional. Copy firewall to target region")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Project-id, see 'ucloud project list'")

	command.SetCompletion(cmd, "src-fw", func() []string {
		return getFirewallIDNames(ctx, *req.ProjectId, srcRegion)
	})
	command.SetCompletion(cmd, "target-region", ctx.RegionList)
	command.SetCompletion(cmd, "region", ctx.RegionList)

	cmd.MarkFlagRequired("src-fw")
	cmd.MarkFlagRequired("name")

	return cmd
}
