package uk8s

import (
	"fmt"

	"github.com/spf13/cobra"

	uk8ssdk "github.com/ucloud/ucloud-sdk-go/services/uk8s"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

func newNodeGroupAdd(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, uk8ssdk.NewClient)
	req := client.NewAddUK8SNodeGroupRequest()

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a UK8S node group",
		Args:  cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Flags().Changed("machine-type") && !oneOf(*req.MachineType, "N", "C", "G", "O", "OS") {
				return fmt.Errorf("--machine-type must be one of N, C, G, O, or OS")
			}
			if cmd.Flags().Changed("cpu") && (*req.CPU < 2 || *req.CPU > 64) {
				return fmt.Errorf("--cpu must be between 2 and 64")
			}
			if cmd.Flags().Changed("memory-mb") && (*req.Mem < 4096 || *req.Mem > 262144 || *req.Mem%1024 != 0) {
				return fmt.Errorf("--memory-mb must be between 4096 and 262144 and a multiple of 1024")
			}
			for _, item := range []struct {
				name     string
				value    *int
				min, max int
			}{
				{"boot-disk-size-gb", req.BootDiskSize, 40, 500},
				{"data-disk-size-gb", req.DataDiskSize, 20, 1000},
			} {
				if cmd.Flags().Changed(item.name) && (*item.value < item.min || *item.value > item.max) {
					return fmt.Errorf("--%s must be between %d and %d", item.name, item.min, item.max)
				}
			}
			if cmd.Flags().Changed("charge-type") && !oneOf(*req.ChargeType, "Dynamic", "Month", "Year") {
				return fmt.Errorf("--charge-type must be one of Dynamic, Month, or Year")
			}
			if req.MachineType != nil && *req.MachineType == "G" {
				if !cmd.Flags().Changed("gpu") || !cmd.Flags().Changed("gpu-type") {
					return fmt.Errorf("--gpu and --gpu-type are required when --machine-type is G")
				}
			} else if cmd.Flags().Changed("gpu") || cmd.Flags().Changed("gpu-type") {
				return fmt.Errorf("--gpu and --gpu-type require --machine-type G")
			}
			if cmd.Flags().Changed("gpu") && *req.GPU < 1 {
				return fmt.Errorf("--gpu must be greater than 0")
			}
			if cmd.Flags().Changed("gpu-type") && !oneOf(*req.GpuType, "K80", "P40", "V100") {
				return fmt.Errorf("--gpu-type must be one of K80, P40, or V100")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			*req.ClusterId = ctx.PickResourceID(*req.ClusterId)
			for _, value := range []*string{req.ImageId, req.SubnetId} {
				if value != nil && *value != "" {
					*value = ctx.PickResourceID(*value)
				}
			}
			resp, err := client.AddUK8SNodeGroup(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "uk8s nodegroup[%s] added\n", resp.NodeGroupId)
			ctx.EmitResult(cli.OpResultRow{ResourceID: resp.NodeGroupId, Action: "add", Status: "Created"})
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	req.ClusterId = flags.String("cluster-id", "", "Required. Cluster ID.")
	req.NodeGroupName = flags.String("name", "", "Required. Node group name.")
	req.MachineType = flags.String("machine-type", "", "Optional. Node machine type.")
	req.CPU = flags.Int("cpu", 0, "Optional. vCPU cores.")
	req.Mem = flags.Int("memory-mb", 0, "Optional. Memory in MB.")
	req.ImageId = flags.String("image-id", "", "Optional. Node image ID.")
	req.SubnetId = flags.String("subnet-id", "", "Optional. Subnet ID.")
	req.BootDiskType = flags.String("boot-disk-type", "", "Optional. Boot disk type.")
	req.BootDiskSize = flags.Int("boot-disk-size-gb", 0, "Optional. Boot disk size in GB.")
	req.DataDiskType = flags.String("data-disk-type", "", "Optional. Data disk type.")
	req.DataDiskSize = flags.Int("data-disk-size-gb", 0, "Optional. Data disk size in GB.")
	req.MinimalCpuPlatform = flags.String("cpu-platform", "", "Optional. Minimum CPU platform.")
	req.ChargeType = flags.String("charge-type", "", "Optional. Billing mode.")
	req.Tag = flags.String("group", "", "Optional. Business group.")
	req.GPU = flags.Int("gpu", 0, "Optional. GPU count; requires machine type G.")
	req.GpuType = flags.String("gpu-type", "", "Optional. GPU type: K80, P40, or V100.")
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)
	cmd.MarkFlagRequired("cluster-id")
	cmd.MarkFlagRequired("name")
	command.SetCompletion(cmd, "cluster-id", func() []string {
		return listClusterIDs(ctx, nil, derefStr(req.Region), derefStr(req.ProjectId))
	})
	command.SetFlagValues(cmd, "machine-type", "N", "C", "G", "O", "OS")
	command.SetFlagValues(cmd, "charge-type", "Dynamic", "Month", "Year")
	command.SetFlagValues(cmd, "gpu-type", "K80", "P40", "V100")
	command.SetCompletion(cmd, "image-id", func() []string {
		return listUK8SImageIDs(ctx, derefStr(req.ProjectId), derefStr(req.Region), derefStr(req.Zone))
	})
	return cmd
}
