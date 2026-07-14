package tidb

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/tidb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newList ucloud utidb list.
// ListTiDBClusterService Limit/Offset are *string in the SDK, so
// ctx.BindCommonParams (BindLimit expects *int) cannot be used here.
func newList(ctx *cli.Context) *cobra.Command {
	var limit, offset string
	client := cli.NewServiceClient(ctx, tidb.NewClient)
	req := client.NewListTiDBClusterServiceRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List UTiDB instances",
		Long:  "List UTiDB instances",
		Run: func(c *cobra.Command, args []string) {
			if limit != "" {
				req.Limit = sdk.String(limit)
			}
			if offset != "" {
				req.Offset = sdk.String(offset)
			}
			resp, err := client.ListTiDBClusterService(req)
			if err != nil {
				handleAPIError(ctx, err)
				return
			}
			rows := []instanceRow{}
			for _, d := range resp.Data {
				rows = append(rows, newInstanceRowFromData(d))
			}
			ctx.PrintList(rows)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)
	flags.StringVar(&limit, "limit", "", "Optional. The maximum number of resources per page")
	flags.StringVar(&offset, "offset", "", "Optional. The index of resource which start to list")

	return cmd
}
