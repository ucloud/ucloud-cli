package ulhost

import (
	"fmt"

	"github.com/spf13/cobra"

	ucompsharesdk "github.com/ucloud/ucloud-sdk-go/services/ucompshare"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newBundles ucloud ulhost bundles
func newBundles(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, ucompsharesdk.NewClient)
	req := client.NewDescribeULHostBundlesRequest()
	cmd := &cobra.Command{
		Use:   "bundles",
		Short: "List all ULHost bundles",
		Long:  `List all ULHost bundles (套餐列表)`,
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := client.DescribeULHostBundles(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			rows := make([]bundleRow, 0, len(resp.Bundles))
			for _, bundle := range resp.Bundles {
				row := bundleRow{
					BundleID:      bundle.BundleId,
					CPU:           fmt.Sprintf("%d", bundle.CPU),
					Memory:        fmt.Sprintf("%dG", bundle.Memory/1024),
					SysDiskSpace:  fmt.Sprintf("%dG", bundle.SysDiskSpace),
					Bandwidth:     fmt.Sprintf("%dM", bundle.Bandwidth),
					TrafficPacket: fmt.Sprintf("%dG", bundle.TrafficPacket),
				}
				rows = append(rows, row)
			}
			ctx.PrintList(rows)
		},
	}
	cmd.Flags().SortFlags = false
	req.ProjectId = cmd.Flags().String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = cmd.Flags().String("region", ctx.DefaultRegion(), "Optional. Assign region")
	command.SetCompletion(cmd, "project-id", ctx.ProjectList)
	command.SetCompletion(cmd, "region", ctx.RegionList)

	return cmd
}
