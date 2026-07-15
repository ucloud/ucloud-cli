package uk8s

import (
	"github.com/spf13/cobra"

	uk8ssdk "github.com/ucloud/ucloud-sdk-go/services/uk8s"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

func newImageList(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, uk8ssdk.NewClient)
	req := client.NewDescribeUK8SImageRequest()

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List images supported by UK8S",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := client.DescribeUK8SImage(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if ctx.Format() != cli.OutputTable {
				ctx.PrintList(resp)
				return
			}
			ctx.PrintList(imageRows(resp))
		},
	}

	cmd.Flags().SortFlags = false
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)
	return cmd
}
