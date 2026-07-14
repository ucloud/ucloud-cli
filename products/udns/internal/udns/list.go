package udns

import (
	"strings"
	"time"

	"github.com/spf13/cobra"

	udnssdk "github.com/ucloud/ucloud-sdk-go/services/udns"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

func newListCommand(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, udnssdk.NewClient)
	req := client.NewDescribeUDNSZoneRequest()
	var zoneIDs []string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List UDNS zones",
		Long:  "List UDNS zones",
		Run: func(cmd *cobra.Command, args []string) {
			req.DNSZoneIds = zoneIDs
			resp, err := client.DescribeUDNSZone(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			rows := make([]zoneRow, 0, len(resp.DNSZoneInfos))
			for _, z := range resp.DNSZoneInfos {
				rows = append(rows, toZoneRow(z))
			}
			ctx.PrintList(rows)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	flags.StringSliceVar(&zoneIDs, "zone-id", nil, "Optional. Filter by zone ID (repeatable)")
	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)
	req.Offset = flags.Int("offset", 0, "Optional. Pagination offset; default 0")
	req.Limit = flags.Int("limit", 20, "Optional. Pagination limit; default 20")
	return cmd
}

func toZoneRow(z udnssdk.ZoneInfo) zoneRow {
	vpcIDs := make([]string, 0, len(z.VPCInfos))
	for _, v := range z.VPCInfos {
		vpcIDs = append(vpcIDs, v.VPCId)
	}
	return zoneRow{
		ZoneID:     z.DNSZoneId,
		Name:       z.DNSZoneName,
		ChargeType: z.ChargeType,
		Recursion:  z.IsRecursionEnabled,
		VPCs:       strings.Join(vpcIDs, ","),
		Tag:        z.Tag,
		Remark:     z.Remark,
		CreateTime: time.Unix(int64(z.CreateTime), 0).Format("2006-01-02"),
		ExpireTime: time.Unix(int64(z.ExpireTime), 0).Format("2006-01-02"),
	}
}
