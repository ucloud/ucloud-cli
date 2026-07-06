package firewall

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/unet"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newApply ucloud firewall apply
func newApply(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, unet.NewClient)
	req := client.NewGrantFirewallRequest()
	resourceIDs := []string{}
	fwID := ""
	cmd := &cobra.Command{
		Use:     "apply",
		Short:   "Applay firewall to ucloud service",
		Long:    "Applay firewall to ucloud service",
		Example: "ucloud firewall apply --fw-id firewall-xxx --resource-id uhost-xxx --resource-type uhost",
		Run: func(c *cobra.Command, args []string) {
			req.FWId = sdk.String(ctx.PickResourceID(fwID))
			results := []cli.OpResultRow{}
			for _, id := range resourceIDs {
				req.ResourceId = sdk.String(id)
				_, err := client.GrantFirewall(req)
				if err != nil {
					ctx.HandleError(err)
					continue
				}
				fmt.Fprintf(ctx.ProgressWriter(), "firewall[%s] applied to %s[%s]\n", fwID, *req.ResourceType, id)
				results = append(results, cli.OpResultRow{ResourceID: id, Action: "apply", Status: "Applied"})
			}
			ctx.EmitResult(results...)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringVar(&fwID, "fw-id", "", "Required. Resource ID of firewall to apply to some ucloud resource")
	req.ResourceType = flags.String("resource-type", "", "Required. Resource type of resource to be applied firewall. Range 'uhost','unatgw','upm','hadoophost','fortresshost','udhost','udockhost','dbaudit'.")
	flags.StringSliceVar(&resourceIDs, "resource-id", nil, "Resource ID of resources to be applied firewall")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Region, see 'ucloud region'")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Project-id, see 'ucloud project list'")

	command.SetFlagValues(cmd, "resource-type", "uhost", "unatgw", "upm", "hadoophost", "fortresshost", "udhost", "udockhost", "dbaudit")
	command.SetCompletion(cmd, "fw-id", func() []string {
		return getFirewallIDNames(ctx, *req.ProjectId, *req.Region)
	})

	cmd.MarkFlagRequired("fw-id")
	cmd.MarkFlagRequired("resource-id")
	cmd.MarkFlagRequired("resource-type")

	return cmd
}
