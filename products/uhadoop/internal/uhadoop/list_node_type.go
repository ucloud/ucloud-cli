package uhadoop

import (
	"strings"

	"github.com/spf13/cobra"

	uhadoopsdk "github.com/ucloud/ucloud-sdk-go/services/uhadoop"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newListNodeType ucloud uhadoop list-node-type
func newListNodeType(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, uhadoopsdk.NewClient)
	req := client.NewGetUHadoopNodeTypeRequest()
	cmd := &cobra.Command{
		Use:   "list-node-type",
		Short: "List available node types for UHadoop",
		Long:  `List available node/instance types for UHadoop clusters`,
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := client.GetUHadoopNodeType(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			listNodeTypes(ctx, resp.InstanceTypeSet)
		},
	}
	cmd.Flags().SortFlags = false

	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	req.Framework = cmd.Flags().String("framework", "", "Optional. Filter by framework: Hadoop|HDFS|MR|StarRocks-Shared-Nothing|StarRocks-Shared-Data")
	req.FrameworkVersion = cmd.Flags().String("framework-version", "", "Optional. Filter by framework version, e.g. 3.3.4-udh3.2")
	req.NodeRole = cmd.Flags().String("node-role", "", "Optional. Filter by node role: master|core|task")
	req.NodeType = cmd.Flags().String("node-type", "", "Optional. Filter by node type name")

	command.SetFlagValues(cmd, "node-role", "master", "core", "task", "client")

	return cmd
}

func listNodeTypes(ctx *cli.Context, types []uhadoopsdk.InstanceType) {
	list := make([]instanceTypeRow, 0, len(types))
	for _, t := range types {
		row := instanceTypeRow{
			NodeType:         t.NodeType,
			HostType:         t.HostType,
			CPU:              t.CPU,
			Memory:           t.Memory,
			CPUToMemoryRatio: t.CPUToMemoryRatio,
			SuitableRole:     strings.Join(t.SuitableRole, ","),
			IsUsable:         t.IsUsable,
			GpuType:          t.GpuType,
			GpuCount:         t.GpuCount,
		}
		if len(t.DiskSet) > 0 {
			// Find the Data disk info
			for _, d := range t.DiskSet {
				if d.Type == "Data" {
					row.DiskType = strings.Join(d.DiskType, ",")
					row.DiskMinSize = d.DiskMinSize
					row.DiskMaxSize = d.DiskMaxSize
					row.DiskMinNum = d.DiskMinNum
					row.DiskMaxNum = d.DiskMaxNum
					break
				}
			}
		}
		list = append(list, row)
	}
	ctx.PrintList(list)
}
