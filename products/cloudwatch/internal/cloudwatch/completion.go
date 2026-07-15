package cloudwatch

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

var (
	calcMethodValues = []string{"raw", "max", "min", "avg", "sum"}
	periodValues     = []string{"60", "300", "3600", "21600", "86400"}
)

// productKeyCandidates returns the live product-key list for --product
// completion by calling ListMonitorProduct — the authoritative, dynamic
// source (products are added/retired over time; a hardcoded list would go
// stale). No request fields are required (empty Filter matches everything).
func productKeyCandidates(ctx *cli.Context) func() []string {
	return func() []string {
		client := newGenericClient(ctx)
		req := client.NewGenericRequest()
		out, err := invoke(client, req, map[string]interface{}{
			"Action": "ListMonitorProduct",
		})
		if err != nil {
			return nil
		}
		var resp listMonitorProductResp
		if err := decodeData(out, &resp); err != nil {
			return nil
		}
		keys := make([]string, 0, len(resp.List))
		for _, p := range resp.List {
			keys = append(keys, p.ProductKey)
		}
		return keys
	}
}

func registerQueryMetricDataCompletions(cmd *cobra.Command) {
	command.SetFlagValues(cmd, "calc-method", calcMethodValues...)
	command.SetFlagValues(cmd, "period", periodValues...)
}

func validateEnum(name, value string, allowed []string) error {
	for _, candidate := range allowed {
		if value == candidate {
			return nil
		}
	}
	return fmt.Errorf("%s must be one of: %s", name, joinEnumValues(allowed))
}

func joinEnumValues(values []string) string {
	result := ""
	for i, value := range values {
		if i > 0 {
			result += ", "
		}
		result += value
	}
	return result
}
