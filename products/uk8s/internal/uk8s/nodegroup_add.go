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
			for _, name := range []string{"cluster-id", "name", "machine-type", "cpu", "memory-mb", "image-id", "subnet-id", "boot-disk-type", "boot-disk-size-gb"} {
				if !cmd.Flags().Changed(name) {
					return fmt.Errorf("--%s is required", name)
				}
			}
			if req.MachineType == nil || *req.MachineType == "" {
				return fmt.Errorf("--machine-type is required")
			}
			if req.CPU == nil {
				return fmt.Errorf("--cpu is required")
			}
			if req.Mem == nil {
				return fmt.Errorf("--memory-mb is required")
			}
			if req.SubnetId == nil || *req.SubnetId == "" {
				return fmt.Errorf("--subnet-id is required")
			}
			if req.ImageId == nil || *req.ImageId == "" {
				return fmt.Errorf("--image-id is required")
			}
			if req.BootDiskType == nil || *req.BootDiskType == "" {
				return fmt.Errorf("--boot-disk-type is required")
			}
			if req.BootDiskSize == nil {
				return fmt.Errorf("--boot-disk-size-gb is required")
			}
			if !oneOf(*req.MachineType, "N", "C", "G", "O", "OS") {
				return fmt.Errorf("--machine-type must be one of N, C, G, O, or OS")
			}
			if *req.CPU < 2 || *req.CPU > 64 {
				return fmt.Errorf("--cpu must be between 2 and 64")
			}
			if *req.Mem < 4096 || *req.Mem > 262144 || *req.Mem%1024 != 0 {
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
				if *item.value != 0 && (*item.value < item.min || *item.value > item.max) {
					return fmt.Errorf("--%s must be between %d and %d", item.name, item.min, item.max)
				}
			}
			if *req.BootDiskSize < 40 || *req.BootDiskSize > 500 {
				return fmt.Errorf("--boot-disk-size-gb must be between 40 and 500")
			}
			if *req.BootDiskType != "CLOUD_RSSD" {
				return fmt.Errorf("--boot-disk-type must be CLOUD_RSSD")
			}
			if !oneOf(*req.ChargeType, "Dynamic", "Month", "Year") {
				return fmt.Errorf("--charge-type must be one of Dynamic, Month, or Year")
			}
			if !oneOf(*req.MinimalCpuPlatform, "Intel/Auto", "Intel/IvyBridge", "Intel/Haswell", "Intel/Broadwell", "Intel/Skylake", "Intel/Cascadelake", "Intel/CascadelakeR", "Amd/Epyc2", "Amd/Auto") {
				return fmt.Errorf("--cpu-platform must be one of Intel/Auto, Intel/IvyBridge, Intel/Haswell, Intel/Broadwell, Intel/Skylake, Intel/Cascadelake, Intel/CascadelakeR, Amd/Epyc2, Amd/Auto")
			}
			if cmd.Flags().Changed("machine-type") && *req.MachineType == "G" {
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
			// Keep the boot disk type explicit because the backend rejects a
			// node-group request with an empty Disks.0.Type. Every other
			// product field remains nil unless the user supplied its flag.
			for name, clear := range map[string]func(){
				"data-disk-type":    func() { req.DataDiskType = nil },
				"data-disk-size-gb": func() { req.DataDiskSize = nil },
				"group":             func() { req.Tag = nil },
				"gpu":               func() { req.GPU = nil },
				"gpu-type":          func() { req.GpuType = nil },
			} {
				if !cmd.Flags().Changed(name) {
					clear()
				}
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
	req.MachineType = flags.String("machine-type", "", "Required. Node machine type. One of N, C, G, O, OS. G requires --gpu and --gpu-type.")
	req.CPU = flags.Int("cpu", 0, "Required. vCPU cores per node. Range 2-64.")
	req.Mem = flags.Int("memory-mb", 0, "Required. Memory in MB per node. Range 4096-262144, multiple of 1024.")
	req.ImageId = flags.String("image-id", "", "Required. Compatible UK8S node image ID. Choose one with 'ucloud uk8s image list'.")
	req.SubnetId = flags.String("subnet-id", "", "Required. Subnet ID; must belong to the cluster's VPC.")
	req.BootDiskType = flags.String("boot-disk-type", "", "Required. System disk type. Only CLOUD_RSSD is supported for UK8S node pools.")
	req.BootDiskSize = flags.Int("boot-disk-size-gb", 0, "Required. Boot disk size in GB. Range 40-500.")
	req.DataDiskType = flags.String("data-disk-type", "", "Optional. Data disk type.")
	req.DataDiskSize = flags.Int("data-disk-size-gb", 0, "Optional. Data disk size in GB.")
	req.MinimalCpuPlatform = flags.String("cpu-platform", "Intel/Auto", "Required. Minimum CPU platform. Defaults to Intel/Auto. One of Intel/Auto, Intel/IvyBridge, Intel/Haswell, Intel/Broadwell, Intel/Skylake, Intel/Cascadelake, Intel/CascadelakeR, Amd/Epyc2, Amd/Auto.")
	req.ChargeType = flags.String("charge-type", "Month", "Required. Billing mode. Defaults to Month.")
	req.Tag = flags.String("group", "", "Optional. Business group.")
	req.GPU = flags.Int("gpu", 0, "Optional. GPU count; requires machine type G.")
	req.GpuType = flags.String("gpu-type", "", "Optional. GPU type: K80, P40, or V100.")
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)
	cmd.MarkFlagRequired("cluster-id")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("machine-type")
	cmd.MarkFlagRequired("cpu")
	cmd.MarkFlagRequired("memory-mb")
	cmd.MarkFlagRequired("image-id")
	cmd.MarkFlagRequired("subnet-id")
	cmd.MarkFlagRequired("boot-disk-type")
	cmd.MarkFlagRequired("boot-disk-size-gb")
	cmd.MarkFlagRequired("charge-type")
	cmd.MarkFlagRequired("cpu-platform")
	command.SetCompletion(cmd, "cluster-id", func() []string {
		return listClusterIDs(ctx, nil, derefStr(req.Region), derefStr(req.ProjectId))
	})
	command.SetFlagValues(cmd, "machine-type", "N", "C", "G", "O", "OS")
	command.SetFlagValues(cmd, "charge-type", "Dynamic", "Month", "Year")
	command.SetFlagValues(cmd, "gpu-type", "K80", "P40", "V100")
	command.SetFlagValues(cmd, "boot-disk-type", "CLOUD_RSSD")
	command.SetFlagValues(cmd, "cpu-platform", "Intel/Auto", "Intel/IvyBridge", "Intel/Haswell", "Intel/Broadwell", "Intel/Skylake", "Intel/Cascadelake", "Intel/CascadelakeR", "Amd/Epyc2", "Amd/Auto")
	command.SetFlagValues(cmd, "data-disk-type", "CLOUD_SSD", "CLOUD_NORMAL", "LOCAL_SSD", "LOCAL_NORMAL", "CLOUD_RSSD", "EXCLUSIVE_LOCAL_DISK")
	command.SetCompletion(cmd, "image-id", func() []string {
		return listUK8SImageIDs(ctx, derefStr(req.ProjectId), derefStr(req.Region), derefStr(req.Zone))
	})
	return cmd
}
