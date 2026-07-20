package ugn

import (
	"github.com/spf13/cobra"

	ugnsdk "github.com/ucloud/ucloud-sdk-go/services/ugn"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// regionRow is the table row for `ucloud ugn region list`.
type regionRow struct {
	Region     string
	RegionID   string
	IsOnline   bool
	IsOverseas bool
}

// newRegionList ucloud ugn region list
func newRegionList(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, ugnsdk.NewClient)
	req := client.NewListUGNRegionsRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List ugn regions",
		Long:  "List ugn regions",
		Run: func(c *cobra.Command, args []string) {
			resp, err := client.ListUGNRegions(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			rows := make([]regionRow, 0, len(resp.RegionLIst))
			for _, r := range resp.RegionLIst {
				rows = append(rows, regionRow{
					Region:     r.Region,
					RegionID:   r.RegIonId,
					IsOnline:   r.IsOnline,
					IsOverseas: r.IsOverseas,
				})
			}
			ctx.PrintList(rows)
		},
	}

	return cmd
}
