package bw

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/unet"

	"github.com/ucloud/ucloud-cli/internal/common"
	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newSharedList returns ucloud bw shared list.
func newSharedList(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, unet.NewClient)
	req := client.NewDescribeShareBandwidthRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List shared bandwidth instances",
		Long:  "List shared bandwidth instances",
		Run: func(c *cobra.Command, args []string) {
			resp, err := client.DescribeShareBandwidth(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			list := []SharedBWRow{}
			for _, sb := range resp.DataSet {
				row := SharedBWRow{}
				row.Name = sb.Name
				row.ResourceID = sb.ShareBandwidthId
				row.ChargeType = sb.ChargeType
				row.Bandwidth = strconv.Itoa(sb.ShareBandwidth) + "Mb"
				row.ExpirationTime = common.FormatDate(sb.ExpireTime)
				eipList := []string{}
				for _, eip := range sb.EIPSet {
					eipText := ""
					eipText += eip.EIPId
					for _, ip := range eip.EIPAddr {
						eipText += fmt.Sprintf("/%s/%s", ip.IP, ip.OperatorName)
					}
					eipList = append(eipList, eipText)
				}
				row.EIP = strings.Join(eipList, "\n")
				list = append(list, row)
			}
			ctx.PrintList(list)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)
	flags.StringSliceVar(&req.ShareBandwidthIds, "shared-bw-id", nil, "Resource ID of shared bandwidth instances to list")

	return cmd
}
