package udns

import (
	"fmt"

	"github.com/spf13/cobra"

	udnssdk "github.com/ucloud/ucloud-sdk-go/services/udns"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newRecordModifyCommand builds `udns record modify` (ModifyUDNSRecord).
func newRecordModifyCommand(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, udnssdk.NewClient)
	req := client.NewModifyUDNSRecordRequest()
	cmd := &cobra.Command{
		Use:   "modify",
		Short: "Modify a DNS record in a UDNS zone",
		Long:  "Modify a DNS record in a UDNS zone",
		Run: func(cmd *cobra.Command, args []string) {
			_, err := client.ModifyUDNSRecord(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "record[%s] modified\n", *req.RecordId)
			ctx.EmitResult(cli.OpResultRow{ResourceID: *req.RecordId, Action: "modify", Status: "Modified"})
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.DNSZoneId = flags.String("zone-id", "", "Required. Zone resource ID")
	req.RecordId = flags.String("record-id", "", "Required. Record resource ID")
	req.Type = flags.String("type", "", "Optional. Record type: A, AAAA, CNAME, MX, TXT, SRV, PTR")
	req.Value = flags.String("value", "", `Optional. Value string: "IP|weight|enabled,..."`)
	req.ValueType = flags.String("value-type", "", "Optional. Normal or Multivalue")
	req.TTL = flags.Int("ttl", 0, "Optional. TTL in seconds (5-600); 0 means unchanged")
	req.Remark = flags.String("remark", "", "Optional. Remark")
	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)
	ctx.SetFlagValues(cmd, "type", "A", "AAAA", "CNAME", "MX", "TXT", "SRV", "PTR")
	ctx.SetFlagValues(cmd, "value-type", "Normal", "Multivalue")
	cmd.MarkFlagRequired("zone-id")
	cmd.MarkFlagRequired("record-id")
	return cmd
}
