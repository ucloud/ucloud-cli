package udns

import (
	"fmt"

	"github.com/spf13/cobra"

	udnssdk "github.com/ucloud/ucloud-sdk-go/services/udns"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// NewCreateCommand builds `udns create` (CreateUDNSZone).
// Exported because product.go already references it by name.
func NewCreateCommand(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, udnssdk.NewClient)
	req := client.NewCreateUDNSZoneRequest()
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a UDNS private DNS zone",
		Long:  "Create a UDNS private DNS zone",
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := client.CreateUDNSZone(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "zone[%s] created\n", resp.DNSZoneId)
			ctx.EmitResult(cli.OpResultRow{ResourceID: resp.DNSZoneId, Action: "create", Status: "Created"})
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.DNSZoneName = flags.String("zone-name", "", "Required. Domain name string")
	req.Type = flags.String("type", "", "Required. Zone type: private or public")
	req.ChargeType = flags.String("charge-type", "Month", "Optional. Year, Month, or Dynamic; default Month")
	req.Quantity = flags.Int("quantity", 1, "Optional. Purchase duration; default 1")
	req.IsRecursionEnabled = flags.String("recursion", "", "Optional. enable or disable")
	req.Tag = flags.String("tag", "", "Optional. Business group")
	req.Remark = flags.String("remark", "", "Optional. Remark")
	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)
	ctx.SetFlagValues(cmd, "type", "private", "public")
	ctx.SetFlagValues(cmd, "charge-type", "Month", "Year", "Dynamic")
	ctx.SetFlagValues(cmd, "recursion", "enable", "disable")
	cmd.MarkFlagRequired("zone-name")
	cmd.MarkFlagRequired("type")
	return cmd
}
