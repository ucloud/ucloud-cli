package clickhouse

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	uclickhousesdk "github.com/ucloud/ucloud-sdk-go/services/uclickhouse"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newCreateOption ucloud clickhouse create-option
func newCreateOption(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, uclickhousesdk.NewClient)
	req := client.NewGetUClickhouseClusterCreateOptionRequest()
	cmd := &cobra.Command{
		Use:   "create-option",
		Short: "List available UClickhouse creation options",
		Long:  "List available UClickhouse creation options",
		Args:  noArgs,
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := getUClickhouseClusterCreateOption(client, req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			ctx.PrintList(createOptionRows(resp.Data))
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	return cmd
}

func getUClickhouseClusterCreateOption(client *uclickhousesdk.UClickhouseClient, req *uclickhousesdk.GetUClickhouseClusterCreateOptionRequest) (*uclickhousesdk.GetUClickhouseClusterCreateOptionResponse, error) {
	var resp uclickhousesdk.GetUClickhouseClusterCreateOptionResponse
	reqCopier := *req
	err := invokeUClickhouseAction(client, "GetUClickhouseClusterCreateOption", &reqCopier, &resp)
	return &resp, err
}

func createOptionRows(data uclickhousesdk.GetCreateUClickhouseClusterOptionResponseData) []CreateOptionRow {
	rows := []CreateOptionRow{}
	for _, version := range data.ClickhouseVersions {
		rows = append(rows, CreateOptionRow{
			OptionType:   "version",
			Version:      version.Version,
			VersionName:  version.VersionName,
			MaxNodeCount: fmt.Sprintf("%d", data.MaxNodeCount),
		})
	}
	rows = append(rows, machineTypeRows("clickhouse", data.MaxNodeCount, data.ClickhouseMachineTypes)...)
	rows = append(rows, machineTypeRows("zookeeper", data.MaxNodeCount, data.ZookeeperMachineTypes)...)
	return rows
}

func machineTypeRows(nodeType string, maxNodeCount int, machineTypes []uclickhousesdk.ClickhouseMachineType) []CreateOptionRow {
	rows := []CreateOptionRow{}
	for _, machineType := range machineTypes {
		for _, option := range machineType.ClickhouseMachineTypeOptions {
			base := CreateOptionRow{
				OptionType:      "machine-type",
				NodeType:        nodeType,
				MachineTypeID:   option.ClickhouseMachineTypeId,
				MachineTypeName: machineType.ClickhouseMachineTypeName,
				MachineType:     option.MachineType,
				CPU:             fmt.Sprintf("%d", option.CPU),
				MemoryGB:        fmt.Sprintf("%d", option.Memory),
				NodeCounts:      joinInts(option.NodeCounts),
				IsSecGroup:      machineType.IsSecgroupMachineType,
				MaxNodeCount:    fmt.Sprintf("%d", maxNodeCount),
			}
			rows = append(rows, base)
			for _, disk := range option.DataDisks {
				diskRow := base
				diskRow.OptionType = "data-disk"
				diskRow.DiskType = disk.DiskType
				diskRow.MinSizeGB = fmt.Sprintf("%d", disk.MinDiskSize)
				diskRow.MaxSizeGB = fmt.Sprintf("%d", disk.MaxDiskSize)
				diskRow.DefaultSizeGB = fmt.Sprintf("%d", disk.DefaultDataDiskSize)
				diskRow.StepGB = fmt.Sprintf("%d", disk.Step)
				rows = append(rows, diskRow)
			}
		}
	}
	return rows
}

func joinInts(values []int) string {
	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, fmt.Sprintf("%d", value))
	}
	return strings.Join(parts, ",")
}
