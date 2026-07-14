package uhadoop

import (
	"fmt"

	"github.com/spf13/cobra"

	uhadoopsdk "github.com/ucloud/ucloud-sdk-go/services/uhadoop"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

func newUpgradeNode(ctx *cli.Context) *cobra.Command {
	var yes, async *bool
	client := cli.NewServiceClient(ctx, uhadoopsdk.NewClient)
	req := client.NewUpgradeUHadoopNodeRequest()
	var nodeNames []string
	cmd := &cobra.Command{
		Use:          "upgrade-node",
		Short:        "Upgrade UHadoop node instance type",
		Long:         `Upgrade UHadoop node to a new instance type`,
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, args []string) {
			w := ctx.ProgressWriter()
			ok, err := ctx.Confirm(*yes, fmt.Sprintf("Upgrade %s nodes on cluster %s to %s?", *req.NodeRole, *req.InstanceId, *req.NodeType))
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if !ok {
				return
			}
			req.NodeNames = nodeNames
			resp, err := client.UpgradeUHadoopNode(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if resp.RetCode != 0 {
				ctx.HandleError(fmt.Errorf("[%d] %s", resp.RetCode, resp.Message))
				return
			}
			text := fmt.Sprintf("uhadoop[%s] upgrading %s nodes", *req.InstanceId, *req.NodeRole)
			if *async {
				fmt.Fprintln(w, text)
			} else {
				ctx.PollerTo(w, describeClusterForPoll(ctx, client)).Spoll(*req.InstanceId, text, []string{StateRunning})
			}
			ctx.EmitResult(cli.OpResultRow{ResourceID: *req.InstanceId, Action: "upgrade-node", Status: "Upgrading"})
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	req.Zone = flags.String("zone", "", "Optional. Assign availability zone")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.InstanceId = flags.String("instance-id", "", "Required. Cluster instance ID")
	req.NodeRole = flags.String("node-role", "", "Required. Node role: master|core|task")
	req.NodeType = flags.String("node-type", "", "Required. New node type")
	yes = flags.BoolP("yes", "y", false, "Do not prompt for confirmation")
	async = flags.Bool("async", false, "Optional. Do not wait for upgrade to complete")
	flags.StringSliceVar(&nodeNames, "node-name", nil, "Node names, required when NodeRole != master")
	command.SetFlagValues(cmd, "node-role", "master", "core", "task", "client")
	cmd.MarkFlagRequired("instance-id")
	cmd.MarkFlagRequired("node-role")
	cmd.MarkFlagRequired("node-type")
	cmd.MarkFlagRequired("region")
	cmd.MarkFlagRequired("zone")
	return cmd
}
