package cloudwatch

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// NewCommand builds the `cloudwatch` root command and mounts its public verbs.
func NewCommand(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cloudwatch",
		Short: "Discover and query CloudWatch metrics",
		Long:  "List monitored products and metrics, then query metric data.",
	}
	cmd.AddCommand(newListProducts(ctx))
	cmd.AddCommand(newListMetrics(ctx))
	cmd.AddCommand(newQueryMetricData(ctx))
	return cmd
}
