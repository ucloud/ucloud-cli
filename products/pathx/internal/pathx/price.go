package pathx

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newPrice builds `ucloud pathx price`.
func newPrice(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "price",
		Short: "List all the acceleration area price",
		Long:  "List all the acceleration area price",
	}
	cmd.AddCommand(newPriceList(ctx))
	return cmd
}
