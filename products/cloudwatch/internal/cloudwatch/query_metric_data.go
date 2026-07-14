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
		  ucloud cloudwatch query-metric-data --product uhost --resource-id uhost-xxx --metric uhost_cpu_used

		  # Repeat --resource-id and --metric for multiple resources and metrics
		  ucloud cloudwatch query-metric-data --product uhost \
		    --resource-id uhost-a --resource-id uhost-b \
		    --metric uhost_cpu_used --metric uhost_mem_used \
		    --tag env=prod --tag role=web --calc-method avg --period 300

		  # Values for the same tag key are OR-ed; different keys are AND-ed
		  ucloud cloudwatch query-metric-data --product uhost --resource-id uhost-a \
		    --metric uhost_cpu_used --tag env=prod --tag env=staging --tag role=web`,
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

			// Cartesian product of --resource-id x --metric: one MetricInfos
			// entry per (resource, metric) combination, all queried in the
			// same request. The backend enforces its own cap on the total
			// number of combinations (config.VM.ReqBatchMaxNum, not fixed at
			// compile time) — CLI does not pre-validate a count, an
			// over-limit request surfaces as a normal backend error via
			// ctx.HandleError.
			metricInfos := buildMetricInfos(ctx, resourceIDs, metrics, tagList)

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
	flags.StringArrayVar(&resourceIDs, "resource-id", nil, "Required. Resource ID; repeat --resource-id to query multiple resources (values are not comma-split)")
	flags.StringArrayVar(&metrics, "metric", nil, "Required. Metric key returned by list-metrics; repeat --metric to query multiple metrics (values are not comma-split)")
	flags.Int64Var(&startTime, "start-time", 0, "Optional. Start time as Unix seconds; defaults to one hour before end-time")
	flags.Int64Var(&endTime, "end-time", 0, "Optional. End time as Unix seconds; defaults to the current time")
	flags.StringVar(&calcMethod, "calc-method", "raw", "Optional. Calculation method: raw, max, min, avg, or sum")
	flags.Int64Var(&period, "period", 0, "Optional. Data interval in seconds: 60, 300, 3600, 21600, or 86400; omit to choose automatically")
	flags.StringArrayVar(&tags, "tag", nil, "Optional. Tag filter as key=value; repeat --tag for multiple values or keys. Same-key values are OR-ed, different keys are AND-ed; commas in values are preserved")
	cmd.MarkFlagRequired("product")
	cmd.MarkFlagRequired("resource-id")
	cmd.MarkFlagRequired("metric")

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

// parseTags converts repeated --tag "key=value" flags into TagList entries.
// Each entry is a map (not a struct) so the SDK generic form encoder can expand
// it into TagList.N.TagKey / TagList.N.TagValues.M form. Values for the same
// key are grouped into one entry (OR); separate keys remain separate entries
// (AND).
func parseTags(tags []string) ([]map[string]interface{}, error) {
	if len(tags) == 0 {
		return nil, nil
	}
	valuesByKey := make(map[string][]string, len(tags))
	keys := make([]string, 0, len(tags))
	for _, t := range tags {
		kv := strings.SplitN(t, "=", 2)
		if len(kv) != 2 || kv[0] == "" || kv[1] == "" {
			return nil, fmt.Errorf("invalid --tag %q, want key=value; repeat --tag for multiple values", t)
		}
		if _, exists := valuesByKey[kv[0]]; !exists {
			keys = append(keys, kv[0])
		}
		valuesByKey[kv[0]] = append(valuesByKey[kv[0]], kv[1])
	}
	list := make([]map[string]interface{}, 0, len(keys))
	for _, key := range keys {
		list = append(list, map[string]interface{}{"TagKey": key, "TagValues": valuesByKey[key]})
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
