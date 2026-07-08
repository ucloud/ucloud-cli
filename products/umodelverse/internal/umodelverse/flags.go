package umodelverse

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func bindProject(cmd *cobra.Command, req interface{ SetProjectIdRef(*string) error }, defaultProject string) {
	projectID := defaultProject
	cmd.Flags().StringVar(&projectID, "project-id", defaultProject, "Optional. Override default project-id for this command invocation, see 'ucloud project list'")
	_ = req.SetProjectIdRef(&projectID)
}

func bindTimeRange(cmd *cobra.Command, req *orderRequest) {
	req.StartTime = cmd.Flags().Int64("start-time", 0, "Required. Query start time, Unix timestamp in seconds.")
	req.EndTime = cmd.Flags().Int64("end-time", 0, "Required. Query end time, Unix timestamp in seconds.")
	cmd.MarkFlagRequired("start-time")
	cmd.MarkFlagRequired("end-time")
}

func bindOrderFilters(cmd *cobra.Command, req *orderRequest) {
	flags := cmd.Flags()
	flags.StringSliceVar(&req.ResourceIds, "resource-id", nil, "Optional. Resource ID filter. Can be repeated or comma-separated.")
	flags.StringSliceVar(&req.ModelIds, "model-id", nil, "Optional. Model ID filter. Can be repeated or comma-separated.")
	flags.IntSliceVar(&req.PricingUnits, "pricing-unit", nil, "Optional. Pricing unit filter. Can be repeated or comma-separated.")
	flags.StringSliceVar(&req.PricingSkus, "pricing-sku", nil, "Optional. Pricing SKU filter. Can be repeated or comma-separated.")
	flags.IntSliceVar(&req.OrderTypes, "order-type", nil, "Optional. Order type filter. Can be repeated or comma-separated.")
	flags.IntSliceVar(&req.OrganizationIds, "organization-id", nil, "Optional. Organization ID filter. Can be repeated or comma-separated.")
	flags.StringSliceVar(&req.Regions, "order-region", nil, "Optional. Order region filter. Can be repeated or comma-separated.")
	flags.StringSliceVar(&req.ProductCodes, "product-code", nil, "Optional. Product code filter, e.g. modelverse or sandbox.")
}

func bindPage(cmd *cobra.Command, req *orderRequest) {
	req.Page = cmd.Flags().Int("page", 1, "Required. Page number, starting from 1.")
	req.PageSize = cmd.Flags().Int("page-size", 20, "Required. Page size.")
	cmd.MarkFlagRequired("page")
	cmd.MarkFlagRequired("page-size")
}

func cleanMultilineFlag(s string) string {
	return strings.ReplaceAll(s, "\\n", "\n")
}

func stringSliceJSONRef(values []string) *string {
	values = normalizeStringSliceValues(values)
	if len(values) == 0 {
		return nil
	}
	b, _ := json.Marshal(values)
	s := string(b)
	return &s
}

func intSliceJSONRef(values []int) *string {
	if len(values) == 0 {
		return nil
	}
	b, _ := json.Marshal(values)
	s := string(b)
	return &s
}

func bindOrderChargeTypes(cmd *cobra.Command, req *orderRequest) {
	cmd.Flags().IntSliceVar(&req.ChargeTypes, "charge-type-code", nil, "Optional. Charge type code filter. Can be repeated or comma-separated.")
}

func normalizeStringSliceValues(values []string) []string {
	if len(values) != 1 {
		return values
	}
	raw := strings.TrimSpace(values[0])
	if len(raw) < 2 || raw[0] != '[' || raw[len(raw)-1] != ']' {
		return values
	}
	var parsed []string
	if err := json.Unmarshal([]byte(raw), &parsed); err == nil {
		return parsed
	}
	raw = strings.TrimSpace(raw[1 : len(raw)-1])
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	parsed = make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.Trim(strings.TrimSpace(part), `"'`)
		if item != "" {
			parsed = append(parsed, item)
		}
	}
	return parsed
}

func clearStringIfUnchanged(flags *pflag.FlagSet, name string, target **string) {
	if !flags.Changed(name) {
		*target = nil
	}
}

func clearIntIfUnchanged(flags *pflag.FlagSet, name string, target **int) {
	if !flags.Changed(name) {
		*target = nil
	}
}

func clearInt64IfUnchanged(flags *pflag.FlagSet, name string, target **int64) {
	if !flags.Changed(name) {
		*target = nil
	}
}

func clearBoolIfUnchanged(flags *pflag.FlagSet, name string, target **bool) {
	if !flags.Changed(name) {
		*target = nil
	}
}
