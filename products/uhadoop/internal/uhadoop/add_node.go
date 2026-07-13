package uhadoop

import (
	"encoding/base64"

	"github.com/spf13/cobra"

	uhadoopsdk "github.com/ucloud/ucloud-sdk-go/services/uhadoop"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newAddNode ucloud uhadoop add-node
func newAddNode(ctx *cli.Context) *cobra.Command {
	var rawPassword string
	client := cli.NewServiceClient(ctx, uhadoopsdk.NewClient)
	req := client.NewAddUHadoopInstanceNodeRequest()
	cmd := &cobra.Command{
		Use:   "add-node",
		Short: "Add nodes to a UHadoop cluster",
		Long:  `Add a number of nodes to an existing UHadoop cluster`,
		Run: func(cmd *cobra.Command, args []string) {
			if rawPassword != "" {
				req.Password = sdkStr(base64.StdEncoding.EncodeToString([]byte(rawPassword)))
			}
			resp, err := client.AddUHadoopInstanceNode(req)
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
	req.NodeRole = cmd.Flags().String("node-role", "", "Required. Node role: core|task|client")
	req.NodeType = cmd.Flags().String("node-type", "", "Required. Node type, e.g. o.hadoop2m.medium (from list-node-type)")
	req.NodeCount = cmd.Flags().Int("node-count", 1, "Optional. Number of nodes to add, default 1")
	cmd.Flags().StringVar(&rawPassword, "password", "", "Optional. Login password, plain text (NodeRole=client is required)")
	req.BootDiskSize = cmd.Flags().String("boot-disk-size", "50", "Optional. Boot disk size in GB for new-type nodes, default 50")
	req.BootDiskType = cmd.Flags().String("boot-disk-type", "CLOUD_RSSD", "Optional. Boot disk type for new-type nodes, default CLOUD_RSSD")
	req.DataDiskSize = cmd.Flags().String("data-disk-size", "200", "Optional. Data disk size in GB for new-type nodes, default 200")
	req.DataDiskNum = cmd.Flags().String("data-disk-num", "1", "Optional. Data disk number for new-type nodes, default 1")
	req.DataDiskType = cmd.Flags().String("data-disk-type", "CLOUD_RSSD", "Optional. Data disk type for new-type nodes, default CLOUD_RSSD")

	command.SetFlagValues(cmd, "node-role", "core", "task", "client")

	return cmd
}
