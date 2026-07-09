package pathx

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	ppathx "github.com/ucloud/ucloud-sdk-go/private/services/pathx"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newUpathList ucloud pathx upath list
func newUpathList(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, ppathx.NewClient)
	req := client.NewDescribeUPathRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list upath instances",
		Long:  "list upath instances",
		Run: func(c *cobra.Command, args []string) {
			resp, err := client.DescribeUPath(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			list := make([]upathRow, 0)
			for _, ins := range resp.UPathSet {
				ids := []string{}
				for _, ga := range ins.UGAList {
					ids = append(ids, ga.UGAId)
				}
				list = append(list, upathRow{
					ResourceID:      ins.UPathId,
					UPathName:       ins.Name,
					AcceleratedPath: fmt.Sprintf("%s->%s %dM", ins.LineFromName, ins.LineToName, ins.Bandwidth),
					BoundUGA:        strings.Join(ids, ","),
				})
			}
			ctx.PrintList(list)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	ctx.BindProjectID(cmd, req)
	req.UPathId = flags.String("upath-id", "", "Optional. Resource ID of upath instance to list")
	return cmd
}
