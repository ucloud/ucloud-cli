package css

import (
	"fmt"

	"github.com/spf13/cobra"

	uessdk "github.com/ucloud/ucloud-sdk-go/services/ues"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newNodeConf ucloud css node-conf
func newNodeConf(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, uessdk.NewClient)
	req := client.NewGetUESNodeConfRequest()
	cmd := &cobra.Command{
		Use:   "node-conf",
		Short: "List available UES node configurations",
		Long:  "List available UES node configurations",
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := client.GetUESNodeConf(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			list := []NodeConfRow{}
			for _, n := range resp.NodeConfList {
				list = append(list, NodeConfRow{
					NodeConf:   n.NodeConf,
					CPU:        fmt.Sprintf("%d", n.CPU),
					MemoryGB:   fmt.Sprintf("%d", n.Memory),
					DiskSizeGB: fmt.Sprintf("%d", n.DiskSize),
					DiskType:   n.DiskType,
					SecGroup:   fmt.Sprintf("%t", n.IsSecGroup),
				})
			}
			ctx.PrintList(list)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.AppVersion = flags.String("app-version", "", "Required. Application version, e.g. elasticsearch-7.10.0")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	req.Zone = flags.String("zone", ctx.DefaultZone(), "Optional. Assign availability zone")

	cmd.MarkFlagRequired("app-version")

	return cmd
}
