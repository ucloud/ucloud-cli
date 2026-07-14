package cloudwatch

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

type monitorProductItem struct {
	ProductKey             string `json:"ProductKey"`
	ProductName            string `json:"ProductName"`
	ProductChName          string `json:"ProductChName"`
	IsSupportHighPrecision bool   `json:"IsSupportHighPrecision"`
}

type listMonitorProductResp struct {
	Total int                  `json:"Total"`
	List  []monitorProductItem `json:"List"`
}

func newListProducts(ctx *cli.Context) *cobra.Command {
	client := newGenericClient(ctx)
	req := client.NewGenericRequest()

	cmd := &cobra.Command{
		Use:   "list-products",
		Short: "List monitored products",
		Long:  "List products that can be queried with CloudWatch.",
		Example: `  # List monitored products
  ucloud cloudwatch list-products

  # Print only product keys as JSON
  ucloud cloudwatch list-products --output json | jq -r '.[].Product'`,
		Args: cobra.NoArgs,
		Run: func(c *cobra.Command, args []string) {
			out, err := invoke(client, req, map[string]interface{}{
				"Action": "ListMonitorProduct",
			})
			if err != nil {
				ctx.HandleError(err)
				return
			}

			var resp listMonitorProductResp
			if err := decodeData(out, &resp); err != nil {
				ctx.HandleError(err)
				return
			}

			rows := make([]MonitorProductRow, 0, len(resp.List))
			for _, p := range resp.List {
				rows = append(rows, MonitorProductRow{
					Product:                p.ProductKey,
					ProductName:            p.ProductName,
					ProductChName:          p.ProductChName,
					IsSupportHighPrecision: p.IsSupportHighPrecision,
				})
			}
			ctx.PrintList(rows)
		},
	}

	return cmd
}
