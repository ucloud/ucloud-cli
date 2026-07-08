package vpc

import (
	"fmt"

	"github.com/spf13/cobra"

	vpcsdk "github.com/ucloud/ucloud-sdk-go/services/vpc"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newDeletePeer returns ucloud vpc delete-intercome.
func newDeletePeer(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, vpcsdk.NewClient)
	req := client.NewDeleteVPCIntercomRequest()
	cmd := &cobra.Command{
		Use:     "delete-intercome",
		Short:   "delete the vpc intercome",
		Long:    "delete the vpc intercome",
		Example: "ucloud vpc delete-intercome --vpc-id xxx --dst-vpc-id xxx",
		Run: func(cmd *cobra.Command, args []string) {
			req.VPCId = sdk.String(ctx.PickResourceID(*req.VPCId))
			req.DstVPCId = sdk.String(ctx.PickResourceID(*req.DstVPCId))
			_, err := client.DeleteVPCIntercom(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "intercome [%s<-->%s] deleted\n", *req.VPCId, *req.DstVPCId)
			ctx.EmitResult(cli.OpResultRow{ResourceID: *req.VPCId, Action: "delete-intercome", Status: "Deleted"})
		},
	}

	cmd.Flags().SortFlags = false

	req.VPCId = cmd.Flags().String("vpc-id", "", "Required. Resource ID of source VPC to disconnect with destination VPC")
	req.DstVPCId = cmd.Flags().String("dst-vpc-id", "", "Required. Resource ID of destination VPC to disconnect with source VPC")
	req.ProjectId = cmd.Flags().String("project-id", ctx.DefaultProjectID(), "Optional. The project id of source vpc")
	req.Region = cmd.Flags().String("region", ctx.DefaultRegion(), "Optional. The region of source vpc to disconnect")
	req.DstRegion = cmd.Flags().String("dst-region", "", "Optional. The region of dest vpc to disconnect")

	cmd.MarkFlagRequired("vpc-id")
	cmd.MarkFlagRequired("dst-vpc-id")
	cmd.MarkFlagRequired("dst-region")

	command.SetCompletion(cmd, "vpc-id", func() []string {
		return getAllVPCIdNames(ctx, *req.ProjectId, *req.Region)
	})
	command.SetCompletion(cmd, "dst-region", ctx.RegionList)

	return cmd
}
