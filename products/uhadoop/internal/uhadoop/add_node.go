package uhadoop

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	uhadoopsdk "github.com/ucloud/ucloud-sdk-go/services/uhadoop"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

func newAddNode(ctx *cli.Context) *cobra.Command {
	var (
		async       *bool
		rawPassword string
	)
	client := cli.NewServiceClient(ctx, uhadoopsdk.NewClient)
	req := client.NewAddUHadoopInstanceNodeRequest()
	cmd := &cobra.Command{
		Use:          "add-node",
		Short:        "Add nodes to a UHadoop cluster",
		Long:         `Add a number of nodes to an existing UHadoop cluster`,
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, args []string) {
			w := ctx.ProgressWriter()
			if rawPassword != "" {
				req.Password = sdkStr(base64.StdEncoding.EncodeToString([]byte(rawPassword)))
			}
			resp, err := client.AddUHadoopInstanceNode(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if resp.RetCode != 0 {
				ctx.HandleError(fmt.Errorf("[%d] %s", resp.RetCode, resp.Message))
				return
			}
			text := fmt.Sprintf("uhadoop[%s] adding %d %s node(s)", *req.InstanceId, *req.NodeCount, *req.NodeRole)
			if *async {
				fmt.Fprintln(w, text)
			} else {
				ctx.PollerTo(w, describeClusterForPoll(ctx, client), cli.WithTimeout(60*time.Minute)).Spoll(*req.InstanceId, text, []string{StateRunning})
			}
			ctx.EmitResult(cli.OpResultRow{ResourceID: *req.InstanceId, Action: "add-node", Status: "Scaling"})
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	req.Zone = flags.String("zone", "", "Optional. Assign availability zone")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.InstanceId = flags.String("instance-id", "", "Required. Cluster instance ID")
	req.NodeRole = flags.String("node-role", "", "Required. Node role: core|task|client")
	req.NodeType = flags.String("node-type", "", "Required. Node type")
	req.NodeCount = flags.Int("node-count", 1, "Number of nodes, default 1")
	flags.StringVar(&rawPassword, "password", "", "Login password (client role requires)")
	req.BootDiskSize = flags.String("boot-disk-size", "50", "Boot disk GB, default 50")
	req.BootDiskType = flags.String("boot-disk-type", "CLOUD_RSSD", "Boot disk type, default CLOUD_RSSD")
	req.DataDiskSize = flags.String("data-disk-size", "200", "Data disk GB, default 200")
	req.DataDiskNum = flags.String("data-disk-num", "1", "Data disk num, default 1")
	req.DataDiskType = flags.String("data-disk-type", "CLOUD_RSSD", "Data disk type, default CLOUD_RSSD")
	async = flags.Bool("async", false, "Optional. Do not wait for node addition to finish")
	command.SetFlagValues(cmd, "node-role", "core", "task", "client")
	cmd.MarkFlagRequired("instance-id")
	cmd.MarkFlagRequired("node-role")
	cmd.MarkFlagRequired("node-type")
	cmd.MarkFlagRequired("region")
	cmd.MarkFlagRequired("zone")
	return cmd
}
