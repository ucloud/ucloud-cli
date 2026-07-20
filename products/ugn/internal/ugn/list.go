package ugn

import (
	"github.com/spf13/cobra"

	ugnsdk "github.com/ucloud/ucloud-sdk-go/services/ugn"

	"github.com/ucloud/ucloud-cli/internal/common"
	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newList ucloud ugn list
func newList(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, ugnsdk.NewClient)
	req := client.NewListUGNRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List ugn instances",
		Long:  "List ugn instances",
		Run: func(c *cobra.Command, args []string) {
			resp, err := client.ListUGN(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			list := []UGNRow{}
			for _, ugn := range resp.UGNs {
				list = append(list, UGNRow{
					ResourceID:     ugn.UGNID,
					Name:           ugn.Name,
					Remark:         ugn.Remark,
					NetworkCount:   ugn.NetworkCount,
					BwPackageCount: ugn.BwPackageCount,
					CreateTime:     common.FormatDate(ugn.CreateTime),
				})
			}
			ctx.PrintList(list)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	ctx.BindOffset(cmd, req)
	req.Limit = flags.Int("limit", 50, "Optional. Limit")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Project-id, see 'ucloud project list'")

	return cmd
}
