package uhadoop

import (
	"github.com/spf13/cobra"

	uhadoopsdk "github.com/ucloud/ucloud-sdk-go/services/uhadoop"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newUpgradeNode ucloud uhadoop upgrade-node
func newUpgradeNode(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, uhadoopsdk.NewClient)
	req := client.NewUpgradeUHadoopNodeRequest()
	var nodeNames []string
	cmd := &cobra.Command{
		Use:   "upgrade-node",
		Short: "Upgrade UHadoop node instance type",
		Long:  `Upgrade UHadoop node to a new instance type`,
		Run: func(cmd *cobra.Command, args []string) {
			req.NodeNames = nodeNames
			resp, err := client.UpgradeUHadoopNode(req)
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
	req.NodeType = cmd.Flags().String("node-type", "", "Required. New node type (from list-node-type)")
	cmd.Flags().StringSliceVar(&nodeNames, "node-name", nil, "Optional. Node names, required when NodeRole is not master")

	command.SetFlagValues(cmd, "node-role", "master", "core", "task", "client")

	return cmd
}
