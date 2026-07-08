package umodelverse

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

func newFilterOptions(ctx *cli.Context) *cobra.Command {
	client := newClient(ctx)
	req := &filterOptionsRequest{}
	newRequest(client, req, true)

	cmd := &cobra.Command{
		Use:   "filter-options",
		Short: "Get uModelVerse order filter options",
		Long:  "Get uModelVerse order filter options.",
		Run: func(c *cobra.Command, args []string) {
			clearStringIfUnchanged(c.Flags(), "product-code", &req.ProductCode)
			resp, err := invokeUMAction(client, "GetFilterOptions", req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			printResponse(ctx, resp)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	req.ProductCode = flags.String("product-code", "", "Optional. Product code, e.g. modelverse or sandbox.")
	return cmd
}
