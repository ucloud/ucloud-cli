package eip

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/unet"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newRelease ucloud eip release
func newRelease(ctx *cli.Context) *cobra.Command {
	var ids []string
	client := cli.NewServiceClient(ctx, unet.NewClient)
	req := client.NewReleaseEIPRequest()
	cmd := &cobra.Command{
		Use:     "release",
		Short:   "Release EIP",
		Long:    "Release EIP",
		Example: "ucloud eip release --eip-id eip-xx1,eip-xx2",
		Run: func(cmd *cobra.Command, args []string) {
			req.ProjectId = sdk.String(ctx.PickResourceID(*req.ProjectId))
			results := []cli.OpResultRow{}
			for _, id := range ids {
				req.EIPId = sdk.String(ctx.PickResourceID(id))
				_, err := client.ReleaseEIP(req)
				if err != nil {
					ctx.HandleError(err)
				} else {
					fmt.Fprintf(ctx.ProgressWriter(), "eip[%s] released\n", *req.EIPId)
					results = append(results, cli.OpResultRow{ResourceID: *req.EIPId, Action: "release", Status: "Released"})
				}
			}
			ctx.EmitResult(results...)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	flags.StringSliceVarP(&ids, "eip-id", "", nil, "Required. Resource ID of the EIPs you want to release")
	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)
	cmd.MarkFlagRequired("eip-id")
	command.SetCompletion(cmd, "eip-id", func() []string {
		return getAllEip(ctx, *req.ProjectId, *req.Region, []string{EIP_FREE}, nil)
	})

	return cmd
}
