package firewall

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/unet"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newDelete ucloud firewall delete
func newDelete(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, unet.NewClient)
	req := client.NewDeleteFirewallRequest()
	ids := []string{}
	cmd := &cobra.Command{
		Use:     "delete",
		Short:   "Delete firewall by resource ids or names",
		Long:    "Delete firewall by resource ids or names",
		Example: "ucloud firewall delete --fw-id firewall-xxx",
		Run: func(c *cobra.Command, args []string) {
			results := []cli.OpResultRow{}
			for _, id := range ids {
				rid := ctx.PickResourceID(id)
				req.FWId = sdk.String(rid)
				_, err := client.DeleteFirewall(req)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				fmt.Fprintf(ctx.ProgressWriter(), "firewall[%s] deleted\n", id)
				results = append(results, cli.OpResultRow{ResourceID: rid, Action: "delete", Status: "Deleted"})
			}
			ctx.EmitResult(results...)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	flags.StringSliceVar(&ids, "fw-id", nil, "Required. Resource IDs of firewall to delete")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Region, see 'ucloud region'")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Project-id, see 'ucloud project list'")

	cmd.MarkFlagRequired("fw-id")
	command.SetCompletion(cmd, "fw-id", func() []string {
		return getFirewallIDNames(ctx, *req.ProjectId, *req.Region)
	})

	return cmd
}
