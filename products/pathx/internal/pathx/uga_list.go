package pathx

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	ppathx "github.com/ucloud/ucloud-sdk-go/private/services/pathx"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newUGAList ucloud pathx uga list
func newUGAList(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, ppathx.NewClient)
	req := client.NewDescribeUGAInstanceRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list uga instances",
		Long:  "list uga instances",
		Run: func(c *cobra.Command, args []string) {
			*req.UGAId = ctx.PickResourceID(*req.UGAId)
			resp, err := client.DescribeUGAInstance(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			list := make([]UGARow, 0)
			for _, ins := range resp.UGAList {
				list = append(list, UGARow{
					ResourceID:      ins.UGAId,
					UGAName:         ins.UGAName,
					CName:           ins.CName,
					Origin:          fmt.Sprintf("%s%s", strings.Join(ins.IPList, ","), ins.Domain),
					AcceleratedPath: getUpathStr(ins.UPathSet),
				})
			}
			ctx.PrintList(list)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.UGAId = flags.String("uga-id", "", "Optional. Resource ID of uga instance")
	ctx.BindProjectID(cmd, req)
	return cmd
}
