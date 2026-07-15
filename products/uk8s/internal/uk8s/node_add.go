package uk8s

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	uk8ssdk "github.com/ucloud/ucloud-sdk-go/services/uk8s"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

func newNodeAdd(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, uk8ssdk.NewClient)
	req := client.NewAddUK8SUHostNodeRequest()
	var userData, userDataB64, initScript, initScriptB64 string

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add UHost nodes to a UK8S cluster",
		Args:  cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if *req.CPU < 2 || *req.CPU > 64 {
				return fmt.Errorf("--cpu must be between 2 and 64")
			}
			if *req.Mem < 4096 || *req.Mem > 262144 || *req.Mem%1024 != 0 {
				return fmt.Errorf("--memory-mb must be between 4096 and 262144 and a multiple of 1024")
			}
			if *req.Count < 1 || *req.Count > 50 {
				return fmt.Errorf("--count must be between 1 and 50")
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
			if cmd.Flags().Changed("quantity") && *req.Quantity < 0 {
				return fmt.Errorf("--quantity must not be negative")
			}
			if !oneOf(*req.ChargeType, "Dynamic", "Month", "Year", "Postpay") {
				return fmt.Errorf("--charge-type must be one of Dynamic, Month, Year, or Postpay")
			}
			if cmd.Flags().Changed("machine-type") && !oneOf(*req.MachineType, "N", "C", "G", "O", "OS") {
				return fmt.Errorf("--machine-type must be one of N, C, G, O, or OS")
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
			if req.IsolationGroup != nil && *req.IsolationGroup != "" && *req.Count > 8 {
				return fmt.Errorf("--count cannot exceed 8 when --isolation-group-id is set")
			}
			if cmd.Flags().Changed("max-pods") && *req.MaxPods < 1 {
				return fmt.Errorf("--max-pods must be greater than 0")
			}
			for name, value := range map[string]*string{"labels": req.Labels, "taints": req.Taints} {
				if value != nil && *value != "" && len(strings.Split(*value, ",")) > 5 {
					return fmt.Errorf("--%s accepts at most 5 comma-separated entries", name)
				}
			}
			if *req.ChargeType == "Dynamic" && cmd.Flags().Changed("quantity") {
				return fmt.Errorf("--quantity must not be set when --charge-type is Dynamic")
			}
			if err := validatePassword(*req.Password); err != nil {
				return err
			}
			if err := bindEncodedValue(cmd, "user-data", userData, "user-data-base64", userDataB64, &req.UserData); err != nil {
				return err
			}
			return bindEncodedValue(cmd, "init-script", initScript, "init-script-base64", initScriptB64, &req.InitScript)
		},
		Run: func(cmd *cobra.Command, args []string) {
			*req.ClusterId = ctx.PickResourceID(*req.ClusterId)
			if req.NodeGroupId != nil && *req.NodeGroupId != "" {
				*req.NodeGroupId = ctx.PickResourceID(*req.NodeGroupId)
			}
			for _, value := range []*string{req.SubnetId, req.ImageId, req.IsolationGroup} {
				if value != nil && *value != "" {
					*value = ctx.PickResourceID(*value)
				}
			}
			req.Password = sdk.String(base64.StdEncoding.EncodeToString([]byte(*req.Password)))
			resp, err := client.AddUK8SUHostNode(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if ctx.Format() != cli.OutputTable {
				ctx.PrintList(resp)
				return
			}
			for _, id := range resp.NodeIds {
				fmt.Fprintf(ctx.ProgressWriter(), "uk8s node[%s] is being added\n", id)
			}
			ctx.PrintList(responseRows(resp))
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	req.ClusterId = flags.String("cluster-id", "", "Required. Cluster ID.")
	req.CPU = flags.Int("cpu", 0, "Required. vCPU cores per node.")
	req.Mem = flags.Int("memory-mb", 0, "Required. Memory in MB per node.")
	req.Count = flags.Int("count", 0, "Required. Number of nodes, 1-50.")
	req.ChargeType = flags.String("charge-type", "", "Required. Billing mode.")
	req.Password = flags.String("password", "", "Required. Plaintext node password. 8-30 chars from A-Z, a-z, 0-9 and ()~!@#$%^&*-+=_|{}[]:;'\\<>,.?/; must include at least 2 of {uppercase, lowercase, digit, special symbol}. Base64-encoded automatically before submission.")
	req.MachineType = flags.String("machine-type", "", "Optional. Node machine type.")
	req.NodeGroupId = flags.String("nodegroup-id", "", "Optional. Node group ID.")
	req.SubnetId = flags.String("subnet-id", "", "Optional. Subnet ID.")
	req.ImageId = flags.String("image-id", "", "Optional. Image ID.")
	req.BootDiskType = flags.String("boot-disk-type", "", "Optional. Boot disk type.")
	req.BootDiskSize = flags.Int("boot-disk-size-gb", 0, "Optional. Boot disk size in GB.")
	req.DataDiskType = flags.String("data-disk-type", "", "Optional. Data disk type.")
	req.DataDiskSize = flags.Int("data-disk-size-gb", 0, "Optional. Data disk size in GB.")
	req.MinimalCpuPlatform = flags.String("cpu-platform", "", "Optional. Minimum CPU platform.")
	req.MaxPods = flags.Int("max-pods", 0, "Optional. Maximum pods per node.")
	req.Quantity = flags.Int("quantity", 0, "Optional. Purchase duration.")
	req.GPU = flags.Int("gpu", 0, "Optional. GPU count.")
	req.GpuType = flags.String("gpu-type", "", "Optional. GPU type.")
	req.IsolationGroup = flags.String("isolation-group-id", "", "Optional. Isolation group ID.")
	req.Labels = flags.String("labels", "", "Optional. Comma-separated node labels.")
	req.Taints = flags.String("taints", "", "Optional. Comma-separated node taints.")
	req.DisableSchedule = flags.Bool("disable-schedule", false, "Optional. Disable scheduling on the new nodes.")
	flags.StringVar(&userData, "user-data", "", "Optional. User data; base64-encoded automatically. Maximum decoded size: 16 KiB.")
	flags.StringVar(&userDataB64, "user-data-base64", "", "Optional. Pre-encoded user data. Conflicts with --user-data.")
	flags.StringVar(&initScript, "init-script", "", "Optional. Post-install script; base64-encoded automatically. Maximum decoded size: 16 KiB.")
	flags.StringVar(&initScriptB64, "init-script-base64", "", "Optional. Pre-encoded post-install script. Conflicts with --init-script.")
	req.Tag = flags.String("group", "", "Optional. Business group.")
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)
	for _, name := range []string{"cluster-id", "cpu", "memory-mb", "count", "charge-type", "password"} {
		cmd.MarkFlagRequired(name)
	}
	command.SetCompletion(cmd, "cluster-id", func() []string {
		return listClusterIDs(ctx, []string{CLUSTER_RUNNING}, derefStr(req.Region), derefStr(req.ProjectId))
	})
	command.SetCompletion(cmd, "nodegroup-id", func() []string {
		return listNodeGroupIDs(ctx, derefStr(req.ClusterId), derefStr(req.ProjectId), derefStr(req.Region))
	})
	command.SetCompletion(cmd, "image-id", func() []string {
		return listUK8SImageIDs(ctx, derefStr(req.ProjectId), derefStr(req.Region), derefStr(req.Zone))
	})
	command.SetCompletion(cmd, "isolation-group-id", func() []string {
		return listIsolationGroupIDs(ctx, derefStr(req.ProjectId), derefStr(req.Region))
	})
	command.SetFlagValues(cmd, "machine-type", "N", "C", "G", "O", "OS")
	command.SetFlagValues(cmd, "charge-type", "Dynamic", "Month", "Year", "Postpay")
	command.SetFlagValues(cmd, "gpu-type", "K80", "P40", "V100")
	return cmd
}
