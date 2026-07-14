package uhadoop

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	uhadoopsdk "github.com/ucloud/ucloud-sdk-go/services/uhadoop"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

func newUpgradeNode(ctx *cli.Context) *cobra.Command {
	var yes bool
	client := cli.NewServiceClient(ctx, uhadoopsdk.NewClient)
	req := client.NewUpgradeUHadoopNodeRequest()
	var nodeNames []string
	cmd := &cobra.Command{
		Use:   "upgrade-node",
		Short: "Upgrade UHadoop node instance type",
		Long:  `Upgrade UHadoop node to a new instance type`,
		Run: func(cmd *cobra.Command, args []string) {
			ok, err := ctx.Confirm(yes, fmt.Sprintf("Upgrade %s nodes on cluster %s to %s?", *req.NodeRole, *req.InstanceId, *req.NodeType))
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if !ok {
				return
			}
			w := ctx.ProgressWriter()
			req.NodeNames = nodeNames
			_, err = client.UpgradeUHadoopNode(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			text := fmt.Sprintf("uhadoop[%s] upgrading %s nodes to %s", *req.InstanceId, *req.NodeRole, *req.NodeType)
			fmt.Fprintln(w, text)
			ctx.PollerTo(w, describeClusterForPoll(ctx, client), cli.WithTimeout(40*time.Minute)).Spoll(*req.InstanceId, text, []string{stateRunning})
			ctx.EmitResult(cli.OpResultRow{ResourceID: *req.InstanceId, Action: "upgrade-node", Status: "Upgrading"})
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)
	req.InstanceId = flags.String("instance-id", "", "Required. Cluster instance ID")
	req.NodeRole = flags.String("node-role", "", "Required. Node role: master|core|task|client")
	req.NodeType = flags.String("node-type", "", "Required. New node type")
	flags.BoolVarP(&yes, "yes", "y", false, "Do not prompt for confirmation")
	flags.StringSliceVar(&nodeNames, "node-name", nil, "Node names, required when NodeRole != master")
	command.SetFlagValues(cmd, "node-role", "master", "core", "task", "client")
	cmd.MarkFlagRequired("instance-id")
	cmd.MarkFlagRequired("node-role")
	cmd.MarkFlagRequired("node-type")
	cmd.MarkFlagRequired("region")
	cmd.MarkFlagRequired("zone")
	return cmd
}
