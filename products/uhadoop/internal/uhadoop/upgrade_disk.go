package uhadoop

import (
	"fmt"

	"github.com/spf13/cobra"

	uhadoopsdk "github.com/ucloud/ucloud-sdk-go/services/uhadoop"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

func newUpgradeDisk(ctx *cli.Context) *cobra.Command {
	var yes *bool
	client := cli.NewServiceClient(ctx, uhadoopsdk.NewClient)
	req := client.NewUpgradeUHadoopNodeDiskRequest()
	var nodeNames []string
	cmd := &cobra.Command{
		Use:          "upgrade-disk",
		Short:        "Upgrade UHadoop node disk size",
		Long:         `Upgrade UHadoop node data disk (and optionally boot disk) size`,
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, args []string) {
			w := ctx.ProgressWriter()
			ok, err := ctx.Confirm(*yes, fmt.Sprintf("Upgrade disk on %s nodes of cluster %s to %d GB?", *req.NodeRole, *req.InstanceId, *req.DataDiskSize))
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if !ok {
				return
			}
			req.NodeNames = nodeNames
			resp, err := client.UpgradeUHadoopNodeDisk(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if resp.RetCode != 0 {
				ctx.HandleError(fmt.Errorf("[%d] %s", resp.RetCode, resp.Message))
				return
			}
			text := fmt.Sprintf("uhadoop[%s] upgrading disk on %s nodes", *req.InstanceId, *req.NodeRole)
			ctx.PollerTo(w, describeClusterForPoll(ctx, client)).Spoll(*req.InstanceId, text, []string{StateRunning})
			ctx.EmitResult(cli.OpResultRow{ResourceID: *req.InstanceId, Action: "upgrade-disk", Status: "Upgrading"})
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	req.Zone = flags.String("zone", "", "Optional. Assign availability zone")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.InstanceId = flags.String("instance-id", "", "Required. Cluster instance ID")
	req.NodeRole = flags.String("node-role", "", "Required. Node role: master|core|task")
	req.DataDiskSize = flags.Int("data-disk-size", 0, "Required. New data disk size in GB")
	req.BootDiskSize = flags.Int("boot-disk-size", 0, "Optional. New boot disk size in GB")
	yes = flags.BoolP("yes", "y", false, "Do not prompt for confirmation")
	flags.StringSliceVar(&nodeNames, "node-name", nil, "Node names, required when NodeRole != master")
	command.SetFlagValues(cmd, "node-role", "master", "core", "task", "client")
	cmd.MarkFlagRequired("instance-id")
	cmd.MarkFlagRequired("node-role")
	cmd.MarkFlagRequired("data-disk-size")
	return cmd
}
