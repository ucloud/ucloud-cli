package uk8s

import (
	"encoding/base64"
	"fmt"

	"github.com/spf13/cobra"

	uk8ssdk "github.com/ucloud/ucloud-sdk-go/services/uk8s"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// masterCount is fixed at 3 by the platform contract (CreateUK8SClusterV2
// builds a 3-master HA control plane; the SDK requires 3 Master entries).
const masterCount = 3

// newCreate implements `ucloud uk8s create`.
//
// Platform APIs exercised: cli.NewServiceClient, ctx.BindCommonParams,
// ctx.PollerTo(...).Spoll (wait path), ctx.ProgressWriter, ctx.EmitResult,
// ctx.HandleError, command.SetFlagValues, command.SetCompletion,
// MarkFlagRequired with "Required." descriptions, the --async pattern.
func newCreate(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, uk8ssdk.NewClient)
	req := client.NewCreateUK8SClusterV2Request()

	var (
		async         bool
		masterZones   []string
		nodeZone      string
		kubeProxyMode string
	)

	cmd := &cobra.Command{
		Use:           "create",
		Short:         "Create a UK8S cluster",
		Long:          "Create a UK8S (UCloud Kubernetes Service) cluster and, unless --async is set, wait for it to become RUNNING.",
		SilenceUsage:  true,
		SilenceErrors: false,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Pad masterZones to masterCount so the user can supply 1 or 3
			// zones (single-AZ dev clusters vs. multi-AZ HA).
			switch len(masterZones) {
			case 1:
				masterZones = []string{masterZones[0], masterZones[0], masterZones[0]}
			case masterCount:
				// already a triple
			default:
				return fmt.Errorf("--master-zone requires exactly 1 or %d entries (got %d)", masterCount, len(masterZones))
			}
			if nodeZone == "" {
				return fmt.Errorf("--node-zone is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Base64-encode the password: the SDK docstring requires the
			// caller to base64-encode (echo -n Password1 | base64).
			if req.Password != nil && *req.Password != "" {
				encoded := base64.StdEncoding.EncodeToString([]byte(*req.Password))
				req.Password = sdk.String(encoded)
			}

			// Build Master slice from --master-zone (pre-padded in PreRunE).
			masters := make([]uk8ssdk.CreateUK8SClusterV2ParamMaster, 0, masterCount)
			for _, z := range masterZones {
				masters = append(masters, uk8ssdk.CreateUK8SClusterV2ParamMaster{Zone: sdk.String(z)})
			}
			req.Master = masters

			// Build a single Nodes group; Node.* fields are bound to the
			// first entry of the slice (count=1 group, multiple VMs).
			node := uk8ssdk.CreateUK8SClusterV2ParamNodes{
				Zone:               sdk.String(nodeZone),
				CPU:                req.Nodes[0].CPU,
				Count:              req.Nodes[0].Count,
				Mem:                req.Nodes[0].Mem,
				MachineType:        req.Nodes[0].MachineType,
				BootDiskType:       req.Nodes[0].BootDiskType,
				BootDiskSIze:       req.Nodes[0].BootDiskSIze,
				DataDiskType:       req.Nodes[0].DataDiskType,
				DataDiskSize:       req.Nodes[0].DataDiskSize,
				MinimalCpuPlatform: req.Nodes[0].MinimalCpuPlatform,
				MaxPods:            req.Nodes[0].MaxPods,
			}
			req.Nodes = []uk8ssdk.CreateUK8SClusterV2ParamNodes{node}

			// Resolve id/name forms for VPC/Subnet/Image (PickResourceID
			// strips the "/Name" tail that completion hands back).
			req.VPCId = sdk.String(ctx.PickResourceID(*req.VPCId))
			req.SubnetId = sdk.String(ctx.PickResourceID(*req.SubnetId))
			if req.ImageId != nil && *req.ImageId != "" {
				id := ctx.PickResourceID(*req.ImageId)
				req.ImageId = sdk.String(id)
			}

			// Wire kube-proxy from the bound flag.
			if kubeProxyMode != "" {
				req.KubeProxy = &uk8ssdk.CreateUK8SClusterV2ParamKubeProxy{Mode: sdk.String(kubeProxyMode)}
			}

			w := ctx.ProgressWriter()
			resp, err := client.CreateUK8SClusterV2(req)
			if err != nil {
				ctx.HandleError(err)
				return nil
			}

			text := fmt.Sprintf("uk8s[%s] is creating", resp.ClusterId)
			if async {
				fmt.Fprintln(w, text)
			} else {
				ctx.PollerTo(w, describeByID(ctx)).Spoll(resp.ClusterId, text, []string{
					CLUSTER_RUNNING, CLUSTER_CREATEFAILED, CLUSTER_ERROR, CLUSTER_ABNORMAL,
				})
			}

			// json/yaml: structured row on stdout; table: no-op (text above
			// is the result).
			ctx.EmitResult(cli.OpResultRow{ResourceID: resp.ClusterId, Action: "create", Status: "Creating"})
			return nil
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	// Pre-allocate the single Nodes group so flag bindings can write into
	// its fields (the SDK request constructor returns an empty slice).
	req.Nodes = []uk8ssdk.CreateUK8SClusterV2ParamNodes{{}}

	// --- Required: cluster identification ---
	req.ClusterName = flags.String("name", "", "Required. Cluster name.")
	req.K8sVersion = flags.String("k8s-version", "", "Required. Kubernetes version (e.g. 1.30.10). See 'ucloud api --action GetUK8SVersions'.")
	req.Password = flags.String("password", "", "Required. Password for cluster nodes (Master + Node). Must include at least one uppercase letter. Base64-encoded automatically before submission.")

	// --- Required: network ---
	req.VPCId = flags.String("vpc-id", "", "Required. VPC ID. See 'ucloud vpc list'.")
	req.SubnetId = flags.String("subnet-id", "", "Required. Subnet ID where nodes and pods live. See 'ucloud subnet list'.")
	req.ServiceCIDR = flags.String("service-cidr", "172.17.0.0/16", "Required. Service CIDR for ClusterIP allocation (e.g. 172.17.0.0/16). Must not overlap with the VPC CIDR.")

	// --- Required: master ---
	req.MasterCPU = flags.Int("master-cpu", 4, "Required. vCPU cores per Master node. Range [2, 64].")
	req.MasterMem = flags.Int("master-memory-mb", 8192, "Required. Memory per Master node. Unit: MB. Range [4096, 262144], multiple of 1024.")
	req.MasterMachineType = flags.String("master-machine-type", "N", "Required. Master machine type. One of N, C, O, OS.")
	flags.StringSliceVar(&masterZones, "master-zone", nil, "Required. Availability zone(s) for the 3 Master nodes. Pass 1 zone (replicated 3x) or 3 zones for multi-AZ HA.")

	// --- Required: nodes (first/only group) ---
	req.Nodes[0].CPU = flags.Int("node-cpu", 4, "Required. vCPU cores per Node. Range [2, 64].")
	req.Nodes[0].Count = flags.Int("node-count", 2, "Required. Node count per group. Range [1, 10].")
	req.Nodes[0].Mem = flags.Int("node-memory-mb", 8192, "Required. Memory per Node. Unit: MB. Range [4096, 262144], multiple of 1024.")
	req.Nodes[0].MachineType = flags.String("node-machine-type", "N", "Required. Node machine type. One of N, C, O, OS.")
	flags.StringVar(&nodeZone, "node-zone", "", "Required. Availability zone for the node group.")

	// --- Optional: image / extras ---
	req.ImageId = flags.String("image-id", "", "Optional. Image ID for both Master and Node. Random if empty. See 'ucloud api --Action DescribeUK8SImage'.")
	req.ExternalApiServer = flags.String("external-api-server", "No", "Optional. Expose the API server on the public internet. Accept values: Yes, No.")
	req.ClusterDomain = flags.String("cluster-domain", "", "Optional. Custom cluster domain.")
	flags.BoolVar(&async, "async", false, "Optional. Do not wait for the cluster to become RUNNING.")

	// Master disk / platform (optional)
	req.MasterBootDiskType = flags.String("master-boot-disk-type", "CLOUD_SSD", "Optional. Master system disk type. See uhost disk types.")
	req.MasterBootDiskSize = flags.Int("master-boot-disk-size-gb", 40, "Optional. Master system disk size in GB. Range [40, 500].")
	req.MasterDataDiskType = flags.String("master-data-disk-type", "", "Optional. Master data disk type.")
	req.MasterDataDiskSize = flags.Int("master-data-disk-size-gb", 0, "Optional. Master data disk size in GB. Range [20, 1000].")
	req.MasterMinimalCpuPlatform = flags.String("master-cpu-platform", "", "Optional. Minimum CPU platform. E.g. Intel/Cascadelake.")

	// Node disk / platform (optional)
	req.Nodes[0].BootDiskType = flags.String("node-boot-disk-type", "CLOUD_SSD", "Optional. Node system disk type.")
	req.Nodes[0].BootDiskSIze = flags.Int("node-boot-disk-size-gb", 40, "Optional. Node system disk size in GB. Range [40, 500].")
	req.Nodes[0].DataDiskType = flags.String("node-data-disk-type", "", "Optional. Node data disk type.")
	req.Nodes[0].DataDiskSize = flags.Int("node-data-disk-size-gb", 0, "Optional. Node data disk size in GB. Range [20, 1000].")
	req.Nodes[0].MinimalCpuPlatform = flags.String("node-cpu-platform", "", "Optional. Minimum CPU platform.")
	req.Nodes[0].MaxPods = flags.Int("node-max-pods", 110, "Optional. Maximum pods per node.")

	// kube-proxy
	flags.StringVar(&kubeProxyMode, "kube-proxy-mode", "iptables", "Optional. kube-proxy mode. Accept values: iptables, ipvs.")

	// Bind charge-type, quantity, region, zone, project-id (only those exist
	// on CreateUK8SClusterV2Request).
	ctx.BindCommonParams(cmd, req)

	// Static enum candidates.
	command.SetFlagValues(cmd, "master-machine-type", "N", "C", "O", "OS")
	command.SetFlagValues(cmd, "node-machine-type", "N", "C", "O", "OS")
	command.SetFlagValues(cmd, "external-api-server", "Yes", "No")
	command.SetFlagValues(cmd, "kube-proxy-mode", "iptables", "ipvs")
	command.SetFlagValues(cmd, "master-boot-disk-type", "CLOUD_SSD", "CLOUD_NORMAL", "LOCAL_SSD", "LOCAL_NORMAL", "CLOUD_RSSD", "EXCLUSIVE_LOCAL_DISK")
	command.SetFlagValues(cmd, "master-data-disk-type", "CLOUD_SSD", "CLOUD_NORMAL", "LOCAL_SSD", "LOCAL_NORMAL", "CLOUD_RSSD", "EXCLUSIVE_LOCAL_DISK", "")
	command.SetFlagValues(cmd, "node-boot-disk-type", "CLOUD_SSD", "CLOUD_NORMAL", "LOCAL_SSD", "LOCAL_NORMAL", "CLOUD_RSSD", "EXCLUSIVE_LOCAL_DISK")
	command.SetFlagValues(cmd, "node-data-disk-type", "CLOUD_SSD", "CLOUD_NORMAL", "LOCAL_SSD", "LOCAL_NORMAL", "CLOUD_RSSD", "EXCLUSIVE_LOCAL_DISK", "")
	command.SetFlagValues(cmd, "master-cpu-platform", "Intel/Auto", "Intel/IvyBridge", "Intel/Haswell", "Intel/Broadwell", "Intel/Skylake", "Intel/Cascadelake")
	command.SetFlagValues(cmd, "node-cpu-platform", "Intel/Auto", "Intel/IvyBridge", "Intel/Haswell", "Intel/Broadwell", "Intel/Skylake", "Intel/Cascadelake")

	// Dynamic completion for resource ids.
	command.SetCompletion(cmd, "vpc-id", func() []string {
		return listVPCIDs(ctx, derefStr(req.ProjectId), derefStr(req.Region))
	})
	command.SetCompletion(cmd, "subnet-id", func() []string {
		return listSubnetIDs(ctx, ctx.PickResourceID(*req.VPCId), derefStr(req.ProjectId), derefStr(req.Region))
	})

	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("k8s-version")
	cmd.MarkFlagRequired("password")
	cmd.MarkFlagRequired("vpc-id")
	cmd.MarkFlagRequired("subnet-id")
	cmd.MarkFlagRequired("service-cidr")
	cmd.MarkFlagRequired("master-cpu")
	cmd.MarkFlagRequired("master-memory-mb")
	cmd.MarkFlagRequired("master-machine-type")
	cmd.MarkFlagRequired("master-zone")
	cmd.MarkFlagRequired("node-cpu")
	cmd.MarkFlagRequired("node-count")
	cmd.MarkFlagRequired("node-memory-mb")
	cmd.MarkFlagRequired("node-machine-type")
	cmd.MarkFlagRequired("node-zone")

	return cmd
}
