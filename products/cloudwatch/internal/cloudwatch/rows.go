package cloudwatch

// MonitorProductRow is one product definition returned by ListMonitorProduct.
type MonitorProductRow struct {
	ProductKey             string
	ProductName            string
	ProductChName          string
	IsSupportHighPrecision bool
}

// MetricRow is one table row for `cloudwatch list-metrics`.
// One row per metric of the queried product.
type MetricRow struct {
	Metric      string
	MetricName  string // MetricChName — Chinese display name, closer to console habit
	Unit        string // Unit.UnitChName (empty when Unit is nil)
	FrequencyMs int32
}

// DataPointRow is one table row for `cloudwatch query-metric-data`.
// The nested response (metric → resource → point[]) is flattened so each row
// is a single (resource, metric, timestamp) sample.
type DataPointRow struct {
	ResourceID   string
	ResourceName string
	Metric       string
	Timestamp    string // common.FormatDateTime(point.Timestamp)
	Value        float64
	Tags         string // TagMap flattened to "k=v, k2=v2"; empty when no tags
}
