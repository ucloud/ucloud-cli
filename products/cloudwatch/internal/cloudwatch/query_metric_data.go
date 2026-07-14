package cloudwatch

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/internal/common"
	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// queryMetricDataResp is the local decode target for the QueryMetricDataSet
// envelope Data. Fields mirror SkymFlameAPI dto.QueryMetricDataResp (release).
type queryMetricDataResp struct {
	List               []metricInfo `json:"List"`
	InvalidResourceIds []string     `json:"InvalidResourceIds"`
}

type metricInfo struct {
	ErrCode int            `json:"ErrCode"`
	ErrMsg  string         `json:"ErrMsg"`
	Metric  string         `json:"Metric"`
	Results []metricValues `json:"Results"`
}

type metricValues struct {
	ResourceID   string            `json:"ResourceId"`
	ResourceName string            `json:"ResourceName"`
	TagMap       map[string]string `json:"TagMap"`
	Values       []metricPoint     `json:"Values"`
}

type metricPoint struct {
	Timestamp int64   `json:"Timestamp"`
	Value     float64 `json:"Value"`
}

// Note on request encoding: the SDK's generic form encoder expands maps and
// slices but rejects nested structs. MetricInfos is therefore built as a
// []map[string]interface{} (with each TagList entry also a map), not structs.

// newQueryMetricData ucloud cloudwatch query-metric-data
func newQueryMetricData(ctx *cli.Context) *cobra.Command {
	var product, calcMethod string
	var resourceIDs, metrics []string
	var startTime, endTime, period int64
	var tags []string
	client := newGenericClient(ctx)
	req := client.NewGenericRequest()

	cmd := &cobra.Command{
		Use:   "query-metric-data",
		Short: "Query metric data",
		Long:  "Query time-series data for one or more resources and metrics. Every resource is paired with every metric.",
		Example: `  # Query one metric on one resource for the default last-hour window
  ucloud cloudwatch query-metric-data --product uhost --resource-ids uhost-xxx --metrics uhost_cpu_used

  # Query two metrics on two resources using 5-minute averages
  ucloud cloudwatch query-metric-data --product uhost \
    --resource-ids uhost-a,uhost-b --metrics uhost_cpu_used,uhost_mem_used \
    --tags env=prod --calc-method avg --period 300`,
		Args: cobra.NoArgs,
		Run: func(c *cobra.Command, args []string) {
			if req.GetProjectId() == "" {
				ctx.HandleError(fmt.Errorf("project-id is required for query-metric-data; pass --project-id or configure a default project"))
				return
			}
			// default time window: the last hour
			now := time.Now().Unix()
			if endTime == 0 {
				endTime = now
			}
			if startTime == 0 {
				startTime = endTime - 3600
			}
			if startTime >= endTime {
				ctx.HandleError(fmt.Errorf("start-time must be earlier than end-time"))
				return
			}
			if err := validateEnum("calc-method", calcMethod, calcMethodValues); err != nil {
				ctx.HandleError(err)
				return
			}
			if period != 0 {
				if err := validateEnum("period", fmt.Sprint(period), periodValues); err != nil {
					ctx.HandleError(err)
					return
				}
			}

			tagList, err := parseTags(tags)
			if err != nil {
				ctx.HandleError(err)
				return
			}

			// --resource-ids / --metrics each accept both repeated flags and
			// comma-separated values within a single flag (freely mixable):
			//   --metrics a --metrics b,c == --metrics a,b,c
			expandedResourceIDs := splitCommaList(resourceIDs)
			expandedMetrics := splitCommaList(metrics)

			// Cartesian product of --resource-id x --metric: one MetricInfos
			// entry per (resource, metric) combination, all queried in the
			// same request. The backend enforces its own cap on the total
			// number of combinations (config.VM.ReqBatchMaxNum, not fixed at
			// compile time) — CLI does not pre-validate a count, an
			// over-limit request surfaces as a normal backend error via
			// ctx.HandleError.
			metricInfos := buildMetricInfos(ctx, expandedResourceIDs, expandedMetrics, tagList)

			payload := map[string]interface{}{
				"Action":      "QueryMetricDataSet",
				"ProductKey":  product,
				"StartTime":   startTime,
				"EndTime":     endTime,
				"CalcMethod":  calcMethod,
				"MetricInfos": metricInfos,
			}
			if period != 0 {
				payload["Period"] = period
			}
			out, err := invoke(client, req, payload)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			var resp queryMetricDataResp
			if err := decodeData(out, &resp); err != nil {
				ctx.HandleError(err)
				return
			}

			rows := make([]DataPointRow, 0)
			for _, mi := range resp.List {
				if mi.ErrCode != 0 {
					ctx.LogWarn(fmt.Sprintf("metric %s error: %s", mi.Metric, mi.ErrMsg))
					continue
				}
				for _, mv := range mi.Results {
					for _, pt := range mv.Values {
						rows = append(rows, DataPointRow{
							ResourceID:   mv.ResourceID,
							ResourceName: mv.ResourceName,
							Metric:       mi.Metric,
							Timestamp:    common.FormatDateTime(int(pt.Timestamp)),
							Value:        pt.Value,
							Tags:         flattenTagMap(mv.TagMap),
						})
					}
				}
			}
			if len(resp.InvalidResourceIds) > 0 {
				ctx.LogWarn(fmt.Sprintf("invalid resource ids: %v", resp.InvalidResourceIds))
			}
			if len(rows) == 0 {
				ctx.LogWarn("no data points in the given time range")
			}
			ctx.PrintList(rows)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	flags.StringVar(&product, "product", "", "Required. Product key returned by list-products, for example uhost")
	flags.StringArrayVar(&resourceIDs, "resource-ids", nil, "Required. Resource IDs; repeat the flag or separate values with commas")
	flags.StringArrayVar(&metrics, "metrics", nil, "Required. Metric keys returned by list-metrics; repeat the flag or separate values with commas")
	flags.Int64Var(&startTime, "start-time", 0, "Optional. Start time as Unix seconds; defaults to one hour before end-time")
	flags.Int64Var(&endTime, "end-time", 0, "Optional. End time as Unix seconds; defaults to the current time")
	flags.StringVar(&calcMethod, "calc-method", "raw", "Optional. Calculation method: raw, max, min, avg, or sum")
	flags.Int64Var(&period, "period", 0, "Optional. Data interval in seconds: 60, 300, 3600, 21600, or 86400; omit to choose automatically")
	flags.StringArrayVar(&tags, "tags", nil, "Optional. Tag filter in key=value1,value2 form; repeat for multiple tag keys")
	cmd.MarkFlagRequired("product")
	cmd.MarkFlagRequired("resource-ids")
	cmd.MarkFlagRequired("metrics")

	ctx.BindProjectID(cmd, req)
	cmd.Flags().Lookup("project-id").Usage = "Required. Project ID"
	cmd.Flags().Lookup("project-id").DefValue = ""
	ctx.BindRegion(cmd, req)
	cmd.Flags().Lookup("region").Usage = "Optional. Region"
	cmd.Flags().Lookup("region").DefValue = ""
	command.SetCompletion(cmd, "product", productKeyCandidates(ctx))
	registerQueryMetricDataCompletions(cmd)

	return cmd
}

func buildMetricInfos(ctx *cli.Context, resourceIDs, metrics []string, tagList []map[string]interface{}) []map[string]interface{} {
	metricInfos := make([]map[string]interface{}, 0, len(resourceIDs)*len(metrics))
	for _, rid := range resourceIDs {
		for _, metric := range metrics {
			info := map[string]interface{}{
				"Metric":     metric,
				"ResourceId": ctx.PickResourceID(rid),
			}
			if len(tagList) > 0 {
				info["TagList"] = tagList
			}
			metricInfos = append(metricInfos, info)
		}
	}
	return metricInfos
}

// parseTags converts repeated --tags "key=v1,v2" flags into TagList entries.
// Each entry is a map (not a struct) so the SDK generic form encoder can expand
// it into TagList.N.TagKey / TagList.N.TagValues.M form.
func parseTags(tags []string) ([]map[string]interface{}, error) {
	if len(tags) == 0 {
		return nil, nil
	}
	list := make([]map[string]interface{}, 0, len(tags))
	for _, t := range tags {
		kv := strings.SplitN(t, "=", 2)
		if len(kv) != 2 || kv[0] == "" {
			return nil, fmt.Errorf("invalid --tags %q, want key=v1,v2", t)
		}
		values := strings.Split(kv[1], ",")
		list = append(list, map[string]interface{}{"TagKey": kv[0], "TagValues": values})
	}
	return list, nil
}

// flattenTagMap renders a tag map as a stable "k=v, k2=v2" string.
func flattenTagMap(m map[string]string) string {
	if len(m) == 0 {
		return ""
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+"="+m[k])
	}
	return strings.Join(parts, ", ")
}

// splitCommaList expands repeatable flag values on commas and removes empty
// elements. It belongs to query-metric-data because that is its only caller.
func splitCommaList(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		for _, part := range strings.Split(value, ",") {
			part = strings.TrimSpace(part)
			if part != "" {
				out = append(out, part)
			}
		}
	}
	return out
}
