package udns

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	udnssdk "github.com/ucloud/ucloud-sdk-go/services/udns"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

func newRecordListCommand(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, udnssdk.NewClient)
	req := client.NewDescribeUDNSRecordRequest()
	var recordIDs []string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List DNS records in a UDNS zone",
		Long:  "List DNS records in a UDNS zone",
		Run: func(cmd *cobra.Command, args []string) {
			req.RecordIds = recordIDs
			resp, err := client.DescribeUDNSRecord(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			rows := make([]recordRow, 0, len(resp.RecordInfos))
			for _, r := range resp.RecordInfos {
				rows = append(rows, toRecordRow(r))
			}
			ctx.PrintList(rows)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.DNSZoneId = flags.String("zone-id", "", "Required. Zone resource ID")
	flags.StringSliceVar(&recordIDs, "record-id", nil, "Optional. Filter by record ID (repeatable)")
	req.Query = flags.String("query", "", "Optional. Fuzzy search string")
	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)
	req.Offset = flags.Int("offset", 0, "Optional. Pagination offset; default 0")
	req.Limit = flags.Int("limit", 20, "Optional. Pagination limit; default 20")
	cmd.MarkFlagRequired("zone-id")
	return cmd
}

func toRecordRow(r udnssdk.RecordInfo) recordRow {
	values := make([]string, 0, len(r.ValueSet))
	for _, v := range r.ValueSet {
		values = append(values, fmt.Sprintf("%s|%d|%d", v.Data, v.Weight, v.IsEnabled))
	}
	return recordRow{
		RecordID:  r.RecordId,
		Name:      r.Name,
		Type:      r.Type,
		TTL:       fmt.Sprintf("%d", r.TTL),
		Values:    strings.Join(values, ","),
		ValueType: r.ValueType,
		Remark:    r.Remark,
	}
}
