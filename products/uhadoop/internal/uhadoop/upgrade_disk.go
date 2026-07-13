package uhadoop

import (
	"github.com/spf13/cobra"

	uhadoopsdk "github.com/ucloud/ucloud-sdk-go/services/uhadoop"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newUpgradeDisk ucloud uhadoop upgrade-disk
func newUpgradeDisk(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, uhadoopsdk.NewClient)
	req := client.NewUpgradeUHadoopNodeDiskRequest()
	var nodeNames []string
	cmd := &cobra.Command{
		Use:   "upgrade-disk",
		Short: "Upgrade UHadoop node disk size",
		Long:  `Upgrade UHadoop node data disk (and optionally boot disk) size`,
		Run: func(cmd *cobra.Command, args []string) {
			req.NodeNames = nodeNames
			resp, err := client.UpgradeUHadoopNodeDisk(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			ctx.PrintJSON(resp)
		},
	}
	cmd.Flags().SortFlags = false

	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)
	req.InstanceId = cmd.Flags().String("instance-id", "", "Required. Cluster instance ID")
	req.NodeRole = cmd.Flags().String("node-role", "", "Required. Node role: master|core|task")
	req.DataDiskSize = cmd.Flags().Int("data-disk-size", 0, "Required. New data disk size in GB")
	req.BootDiskSize = cmd.Flags().Int("boot-disk-size", 0, "Optional. New boot disk size in GB")
	cmd.Flags().StringSliceVar(&nodeNames, "node-name", nil, "Optional. Node names, required when NodeRole is not master")

	command.SetFlagValues(cmd, "node-role", "master", "core", "task", "client")

	return cmd
}
