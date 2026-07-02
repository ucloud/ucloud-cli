package firewall

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/unet"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newUpdate ucloud firewall update
func newUpdate(ctx *cli.Context) *cobra.Command {
	fwIDs := []string{}
	client := cli.NewServiceClient(ctx, unet.NewClient)
	req := client.NewUpdateFirewallAttributeRequest()
	cmd := &cobra.Command{
		Use:     "update",
		Short:   "Update firewall attribute, such as name,group and remark.",
		Long:    "Update firewall attribute, such as name,group and remark.",
		Example: `ucloud firewall update --fw-id firewall-2xxxx/test2 --name test_update.1 --remark "this is a remark"`,
		Run: func(c *cobra.Command, args []string) {
			if *req.Name == "" && *req.Tag == "" && *req.Remark == "" {
				fmt.Fprintln(ctx.Err(), "Error: name, group and remark can't be all empty")
				return
			}
			if *req.Name == "" {
				req.Name = nil
			}
			if *req.Tag == "" {
				req.Tag = nil
			}
			if *req.Remark == "" {
				req.Remark = nil
			}
			results := []cli.OpResultRow{}
			for _, id := range fwIDs {
				rid := ctx.PickResourceID(id)
				req.FWId = sdk.String(rid)
				_, err := client.UpdateFirewallAttribute(req)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				fmt.Fprintf(ctx.ProgressWriter(), "firewall[%s] updated\n", id)
				results = append(results, cli.OpResultRow{ResourceID: rid, Action: "update", Status: "Updated"})
			}
			ctx.EmitResult(results...)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&fwIDs, "fw-id", nil, "Required. Resource ID of firewalls")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Region, see 'ucloud region'")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Project-id, see 'ucloud project list'")
	req.Name = flags.String("name", "", "Name of firewall")
	req.Tag = flags.String("group", "", "Group of firewall")
	req.Remark = flags.String("remark", "", "Remark of firewall")

	command.SetCompletion(cmd, "fw-id", func() []string {
		return getFirewallIDNames(ctx, *req.ProjectId, *req.Region)
	})

	cmd.MarkFlagRequired("fw-id")

	return cmd
}
