package cloudwatch

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// getProductMetricResp is the local decode target for the GetProductMetrics
// envelope Data. Fields mirror SkymFlameAPI dto.GetProductMetricListResp
// (release branch) — only what the CLI renders is declared.
type getProductMetricResp struct {
	Total int64        `json:"Total"`
	List  []metricItem `json:"List"`
}

type metricItem struct {
	Metric       string    `json:"Metric"`
	MetricName   string    `json:"MetricName"`
	MetricChName string    `json:"MetricChName"`
	FrequencyMs  int32     `json:"FrequencyMs"`
	Unit         *unitItem `json:"Unit"`
}

type unitItem struct {
	UnitChName string `json:"UnitChName"`
	UnitName   string `json:"UnitName"`
}

func newListMetrics(ctx *cli.Context) *cobra.Command {
	var product, monitorType string
	client := newGenericClient(ctx)
	req := client.NewGenericRequest()

	cmd := &cobra.Command{
		Use:   "list-metrics",
		Short: "List metrics for a product",
		Long:  "List the metrics available for one monitored product.",
		Example: `  # List all UHost metrics
  ucloud cloudwatch list-metrics --product uhost

	# List only basic UHost metrics
	ucloud cloudwatch list-metrics --product uhost --monitor-type basic`,
		Args: cobra.NoArgs,
		Run: func(c *cobra.Command, args []string) {
			payload := map[string]interface{}{
				"Action":     "GetProductMetrics",
				"ProductKey": product,
			}
			if monitorType != "" {
				payload["MonitorType"] = monitorType
			}
			out, err := invoke(client, req, payload)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			var resp getProductMetricResp
			if err := decodeData(out, &resp); err != nil {
				ctx.HandleError(err)
				return
			}
			rows := make([]MetricRow, 0, len(resp.List))
			for _, m := range resp.List {
				name := m.MetricChName
				if name == "" {
					name = m.MetricName
				}
				unit := ""
				if m.Unit != nil {
					unit = m.Unit.UnitChName
					if unit == "" {
						unit = m.Unit.UnitName
					}
				}
				rows = append(rows, MetricRow{
					Metric:      m.Metric,
					MetricName:  name,
					Unit:        unit,
					FrequencyMs: m.FrequencyMs,
				})
			}
			ctx.PrintList(rows)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	flags.StringVar(&product, "product", "", "Required. Product key returned by list-products, for example uhost")
	flags.StringVar(&monitorType, "monitor-type", "", "Optional. Metric type filter; omit to list all types")
	cmd.MarkFlagRequired("product")

	command.SetCompletion(cmd, "product", productKeyCandidates(ctx))

	return cmd
}
