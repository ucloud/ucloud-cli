package bw

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/unet"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newPkgDelete returns ucloud bw pkg delete.
func newPkgDelete(ctx *cli.Context) *cobra.Command {
	ids := []string{}
	client := cli.NewServiceClient(ctx, unet.NewClient)
	req := client.NewDeleteBandwidthPackageRequest()
	cmd := &cobra.Command{
		Use:     "delete",
		Short:   "Delete bandwidth packages",
		Long:    "Delete bandwidth packages",
		Example: "ucloud bw pkg delete --resource-id bwpack-xxx",
		Run: func(c *cobra.Command, args []string) {
			results := []cli.OpResultRow{}
			for _, idname := range ids {
				id := ctx.PickResourceID(idname)
				req.BandwidthPackageId = &id
				_, err := client.DeleteBandwidthPackage(req)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				fmt.Fprintf(ctx.ProgressWriter(), "bandwidth package[%s] deleted\n", id)
				results = append(results, cli.OpResultRow{ResourceID: id, Action: "delete", Status: "Deleted"})
			}
			ctx.EmitResult(results...)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	flags.StringSliceVar(&ids, "resource-id", nil, "Required, Resource ID of bandwidth package to delete")
	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)

	return cmd
}
