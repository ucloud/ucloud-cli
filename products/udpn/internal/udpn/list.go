package udpn

import (
	"fmt"

	"github.com/spf13/cobra"

	udpnsdk "github.com/ucloud/ucloud-sdk-go/services/udpn"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/internal/common"
	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newList ucloud udpn list
func newList(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, udpnsdk.NewClient)
	req := client.NewDescribeUDPNRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List udpn instances",
		Long:  "List udpn instances",
		Run: func(c *cobra.Command, args []string) {
			req.UDPNId = sdk.String(ctx.PickResourceID(*req.UDPNId))
			resp, err := client.DescribeUDPN(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			list := []UDPNRow{}
			for _, udpn := range resp.DataSet {
				list = append(list, UDPNRow{
					ResourceID:   udpn.UDPNId,
					Peers:        fmt.Sprintf("%s <--> %s", udpn.Peer1, udpn.Peer2),
					Bandwidth:    fmt.Sprintf("%dMb", udpn.Bandwidth),
					ChargeType:   udpn.ChargeType,
					CreationTime: common.FormatDate(udpn.CreateTime),
				})
			}
			ctx.PrintList(list)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.UDPNId = flags.String("udpn-id", "", "Optional. Resource ID of udpn instances to list")
	ctx.BindOffset(cmd, req)
	req.Limit = flags.Int("limit", 50, "Optional. Limit")
	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)

	ctx.SetCompletion(cmd, "udpn-id", func() []string {
		return getAllUDPNIdNames(ctx, *req.ProjectId, *req.Region)
	})

	return cmd
}
