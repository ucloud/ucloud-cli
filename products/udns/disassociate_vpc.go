package udns

import (
	"fmt"

	"github.com/spf13/cobra"

	udnssdk "github.com/ucloud/ucloud-sdk-go/services/udns"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newDisassociateVPCCommand builds `udns disassociate-vpc` (DisassociateUDNSZoneVPC).
func newDisassociateVPCCommand(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, udnssdk.NewClient)
	req := client.NewDisassociateUDNSZoneVPCRequest()
	cmd := &cobra.Command{
		Use:   "disassociate-vpc",
		Short: "Disassociate a UDNS zone from a VPC",
		Long:  "Disassociate a UDNS zone from a VPC",
		Run: func(cmd *cobra.Command, args []string) {
			zoneID := ctx.PickResourceID(*req.DNSZoneId)
			req.DNSZoneId = &zoneID
			_, err := client.DisassociateUDNSZoneVPC(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "zone[%s] disassociated from vpc[%s]\n", zoneID, *req.VPCId)
			ctx.EmitResult(cli.OpResultRow{ResourceID: zoneID, Action: "disassociate-vpc", Status: "Disassociated"})
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.DNSZoneId = flags.String("zone-id", "", "Required. Zone resource ID")
	req.VPCId = flags.String("vpc-id", "", "Required. VPC resource ID")
	req.VPCProjectId = flags.String("vpc-project-id", "", "Required. Project ID that owns the VPC")
	ctx.BindRegion(cmd, req)
	cmd.MarkFlagRequired("zone-id")
	cmd.MarkFlagRequired("vpc-id")
	cmd.MarkFlagRequired("vpc-project-id")
	return cmd
}
