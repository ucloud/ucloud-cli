package udns

import (
	"fmt"

	"github.com/spf13/cobra"

	udnssdk "github.com/ucloud/ucloud-sdk-go/services/udns"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newAssociateVPCCommand builds `udns associate-vpc` (AssociateUDNSZoneVPC).
func newAssociateVPCCommand(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, udnssdk.NewClient)
	req := client.NewAssociateUDNSZoneVPCRequest()
	cmd := &cobra.Command{
		Use:   "associate-vpc",
		Short: "Associate a UDNS zone with a VPC",
		Long:  "Associate a UDNS zone with a VPC",
		Run: func(cmd *cobra.Command, args []string) {
			zoneID := ctx.PickResourceID(*req.DNSZoneId)
			req.DNSZoneId = &zoneID
			_, err := client.AssociateUDNSZoneVPC(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "zone[%s] associated with vpc[%s]\n", zoneID, *req.VPCId)
			ctx.EmitResult(cli.OpResultRow{ResourceID: zoneID, Action: "associate-vpc", Status: "Associated"})
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
