package vpc

import (
	"fmt"

	"github.com/spf13/cobra"

	vpcsdk "github.com/ucloud/ucloud-sdk-go/services/vpc"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newCreate returns ucloud vpc create.
func newCreate(ctx *cli.Context) *cobra.Command {
	var segments *[]string
	client := cli.NewServiceClient(ctx, vpcsdk.NewClient)
	req := client.NewCreateVPCRequest()
	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Create vpc network",
		Long:    "Create vpc network",
		Example: "ucloud vpc create --name xxx --segment 192.168.0.0/16",
		Args:    cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			req.Network = *segments
			req.ProjectId = sdk.String(ctx.PickResourceID(*req.ProjectId))
			resp, err := client.CreateVPC(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "vpc[%s] created\n", resp.VPCId)
			ctx.EmitResult(cli.OpResultRow{ResourceID: resp.VPCId, Action: "create", Status: "Created"})
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	req.Name = flags.String("name", "", "Required. Name of the vpc network.")
	segments = flags.StringSlice("segment", nil, "Required. The segment for private network.")
	req.Tag = flags.String("group", "", "Optional. Business group.")
	req.Remark = flags.String("remark", "", "Optional. The description of the vpc.")
	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)

	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("segment")

	return cmd
}
