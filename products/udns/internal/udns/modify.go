package udns

import (
	"fmt"

	"github.com/spf13/cobra"

	udnssdk "github.com/ucloud/ucloud-sdk-go/services/udns"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

func newModifyCommand(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, udnssdk.NewClient)
	req := client.NewModifyUDNSZoneRequest()
	cmd := &cobra.Command{
		Use:   "modify",
		Short: "Modify a UDNS zone",
		Long:  "Modify a UDNS zone (recursion and remark only)",
		Run: func(cmd *cobra.Command, args []string) {
			id := ctx.PickResourceID(*req.DNSZoneId)
			req.DNSZoneId = &id
			_, err := client.ModifyUDNSZone(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "zone[%s] modified\n", id)
			ctx.EmitResult(cli.OpResultRow{ResourceID: id, Action: "modify", Status: "Modified"})
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.DNSZoneId = flags.String("zone-id", "", "Required. Zone resource ID")
	req.IsRecursionEnabled = flags.String("recursion", "", "Optional. enable or disable")
	req.Remark = flags.String("remark", "", "Optional. Remark")
	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)
	command.SetFlagValues(cmd, "recursion", "enable", "disable")
	cmd.MarkFlagRequired("zone-id")
	return cmd
}
