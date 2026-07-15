package udns

import (
	"fmt"

	"github.com/spf13/cobra"

	udnssdk "github.com/ucloud/ucloud-sdk-go/services/udns"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

func newRecordCreateCommand(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, udnssdk.NewClient)
	req := client.NewCreateUDNSRecordRequest()
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a DNS record in a UDNS zone",
		Long:  "Create a DNS record in a UDNS zone",
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := client.CreateUDNSRecord(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "record[%s] created\n", resp.DNSRecordId)
			ctx.EmitResult(cli.OpResultRow{ResourceID: resp.DNSRecordId, Action: "create", Status: "Created"})
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.DNSZoneId = flags.String("zone-id", "", "Required. Zone resource ID")
	req.Name = flags.String("name", "", "Required. Host record (subdomain prefix)")
	req.Type = flags.String("type", "", "Required. Record type: A, AAAA, CNAME, MX, TXT, SRV, PTR")
	req.Value = flags.String("value", "", `Required. Value string: "IP|weight|enabled,..." e.g. "192.168.1.1|1|1"`)
	req.ValueType = flags.String("value-type", "", "Required. Normal or Multivalue")
	req.TTL = flags.Int("ttl", 5, "Optional. TTL in seconds (5-600); default 5")
	req.Remark = flags.String("remark", "", "Optional. Remark")
	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)
	command.SetFlagValues(cmd, "type", "A", "AAAA", "CNAME", "MX", "TXT", "SRV", "PTR")
	command.SetFlagValues(cmd, "value-type", "Normal", "Multivalue")
	cmd.MarkFlagRequired("zone-id")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("type")
	cmd.MarkFlagRequired("value")
	cmd.MarkFlagRequired("value-type")
	return cmd
}
