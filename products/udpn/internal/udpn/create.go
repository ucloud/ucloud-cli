package udpn

import (
	"fmt"

	"github.com/spf13/cobra"

	udpnsdk "github.com/ucloud/ucloud-sdk-go/services/udpn"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newCreate ucloud udpn create
func newCreate(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, udpnsdk.NewClient)
	req := client.NewAllocateUDPNRequest()
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create UDPN tunnel",
		Long:  "Create UDPN tunnel",
		Run: func(c *cobra.Command, args []string) {
			if *req.Peer1 == *req.Peer2 {
				fmt.Fprintln(ctx.ProgressWriter(), "Error, flags peer1 and peer2 can't be equal")
				return
			}
			req.ProjectId = sdk.String(ctx.PickResourceID(*req.ProjectId))
			resp, err := client.AllocateUDPN(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "udpn[%s] created\n", resp.UDPNId)
			ctx.EmitResult(cli.OpResultRow{ResourceID: resp.UDPNId, Action: "create", Status: "Created"})
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.Peer1 = flags.String("peer1", ctx.DefaultRegion(), "Required. One end of the tunnel to create")
	req.Peer2 = flags.String("peer2", "", "Required. The other end of the tunnel create")
	req.Bandwidth = flags.Int("bandwidth-mb", 0, "Required. Bandwidth of the tunnel to create. Unit:Mb. Rnange [2,1000]")
	req.ChargeType = flags.String("charge-type", "", "Optional. Enumeration value.'Year',pay yearly;'Month',pay monthly;'Dynamic', pay hourly")
	req.Quantity = flags.Int("quantity", 1, "Optional. The duration of the instance. N years/months.")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Project-id, see 'ucloud project list'")

	command.SetFlagValues(cmd, "charge-type", "Month", "Year", "Dynamic")
	ctx.SetCompletion(cmd, "project-id", ctx.ProjectList)
	ctx.SetCompletion(cmd, "peer1", ctx.RegionList)
	ctx.SetCompletion(cmd, "peer2", func() []string {
		regions := ctx.RegionList()
		list := []string{}
		for _, r := range regions {
			if r != *req.Peer1 {
				list = append(list, r)
			}
		}
		return list
	})

	cmd.MarkFlagRequired("peer1")
	cmd.MarkFlagRequired("peer2")
	cmd.MarkFlagRequired("bandwidth-mb")

	return cmd
}
