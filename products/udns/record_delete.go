package udns

import (
	"fmt"

	"github.com/spf13/cobra"

	udnssdk "github.com/ucloud/ucloud-sdk-go/services/udns"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newRecordDeleteCommand builds `udns record delete` (DeleteUDNSRecord).
func newRecordDeleteCommand(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, udnssdk.NewClient)
	req := client.NewDeleteUDNSRecordRequest()
	var recordIDs []string
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete DNS records from a UDNS zone",
		Long:  "Delete DNS records from a UDNS zone",
		Run: func(cmd *cobra.Command, args []string) {
			for i, id := range recordIDs {
				recordIDs[i] = ctx.PickResourceID(id)
			}
			req.RecordIds = recordIDs
			_, err := client.DeleteUDNSRecord(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "records %v deleted from zone[%s]\n", recordIDs, *req.DNSZoneId)
			results := make([]cli.OpResultRow, 0, len(recordIDs))
			for _, id := range recordIDs {
				results = append(results, cli.OpResultRow{ResourceID: id, Action: "delete", Status: "Deleted"})
			}
			ctx.EmitResult(results...)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.DNSZoneId = flags.String("zone-id", "", "Required. Zone resource ID")
	flags.StringSliceVar(&recordIDs, "record-id", nil, "Required. Record resource ID (repeatable)")
	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)
	cmd.MarkFlagRequired("zone-id")
	cmd.MarkFlagRequired("record-id")
	return cmd
}
