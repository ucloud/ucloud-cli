package vpc

import (
	"fmt"

	"github.com/spf13/cobra"

	vpcsdk "github.com/ucloud/ucloud-sdk-go/services/vpc"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newCreatePeer returns ucloud vpc create-intercome.
func newCreatePeer(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, vpcsdk.NewClient)
	req := client.NewCreateVPCIntercomRequest()
	cmd := &cobra.Command{
		Use:     "create-intercome",
		Short:   "Create intercome with other vpc",
		Long:    "Create intercome with other vpc",
		Example: "ucloud vpc create-intercome --vpc-id xx --dst-vpc-id xx --dst-region xx",
		Run: func(cmd *cobra.Command, args []string) {
			req.ProjectId = sdk.String(ctx.PickResourceID(*req.ProjectId))
			req.DstProjectId = sdk.String(ctx.PickResourceID(*req.DstProjectId))
			req.VPCId = sdk.String(ctx.PickResourceID(*req.VPCId))
			req.DstVPCId = sdk.String(ctx.PickResourceID(*req.DstVPCId))
			_, err := client.CreateVPCIntercom(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "intercome [%s<-->%s] establish", *req.VPCId, *req.DstVPCId)
			ctx.EmitResult(cli.OpResultRow{ResourceID: *req.VPCId, Action: "create-intercome", Status: "Created"})
		},
	}

	cmd.Flags().SortFlags = false

	req.VPCId = cmd.Flags().String("vpc-id", "", "Required. The source vpc you want to establish the intercome")
	req.DstVPCId = cmd.Flags().String("dst-vpc-id", "", "Required. The target vpc you want to establish the intercome")
	req.DstRegion = cmd.Flags().String("dst-region", ctx.DefaultRegion(), "Required. If the intercome established across different regions")
	req.Region = cmd.Flags().String("region", ctx.DefaultRegion(), "Optioanl. The region of source vpc which will establish the intercome")
	req.ProjectId = cmd.Flags().String("project-id", ctx.DefaultProjectID(), "Optional. The project id of the source vpc")
	req.DstProjectId = cmd.Flags().String("dst-project-id", ctx.DefaultProjectID(), "Optional. The project id of the source vpc")

	cmd.MarkFlagRequired("vpc-id")
	cmd.MarkFlagRequired("dst-vpc-id")

	command.SetCompletion(cmd, "vpc-id", func() []string {
		return getAllVPCIdNames(ctx, *req.ProjectId, *req.Region)
	})
	command.SetCompletion(cmd, "dst-vpc-id", func() []string {
		return getAllVPCIdNames(ctx, *req.DstProjectId, *req.DstRegion)
	})
	command.SetCompletion(cmd, "region", ctx.RegionList)
	command.SetCompletion(cmd, "dst-region", ctx.RegionList)
	command.SetCompletion(cmd, "project-id", ctx.ProjectList)
	command.SetCompletion(cmd, "dst-project-id", ctx.ProjectList)

	return cmd
}
