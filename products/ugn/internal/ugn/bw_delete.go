package ugn

import (
	"fmt"

	"github.com/spf13/cobra"

	ugnsdk "github.com/ucloud/ucloud-sdk-go/services/ugn"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newBWDelete ucloud ugn bw delete
func newBWDelete(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, ugnsdk.NewClient)
	req := client.NewDeleteUGNBwPackageRequest()
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete ugn bandwidth packages",
		Long:  "Delete ugn bandwidth packages",
		Run: func(c *cobra.Command, args []string) {
			ok, err := ctx.Confirm(yes, "Are you sure you want to delete the bandwidth package?")
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if !ok {
				return
			}
			_, err = client.DeleteUGNBwPackage(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "ugn bw[%s] deleted\n", *req.BwPackageID)
			ctx.EmitResult(cli.OpResultRow{ResourceID: *req.BwPackageID, Action: "delete-bw", Status: "Deleted"})
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.BwPackageID = flags.String("bw-package-id", "", "Required. Resource ID of the bandwidth package to delete")
	req.UGNID = flags.String("ugn-id", "", "Required. Resource ID of the ugn instance")
	flags.BoolVarP(&yes, "yes", "y", false, "Optional. Skip the confirmation prompt.")

	cmd.MarkFlagRequired("bw-package-id")
	cmd.MarkFlagRequired("ugn-id")

	return cmd
}
