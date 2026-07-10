package uk8s

import (
	"encoding/base64"
	"fmt"
	"net"
	"strings"

	"github.com/spf13/cobra"

	uk8ssdk "github.com/ucloud/ucloud-sdk-go/services/uk8s"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/internal/common"
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
		chargeType    string
		quantity      int
		userData      string
		userDataB64   string
		initScript    string
		initScriptB64 string
	)

	cmd := &cobra.Command{
		Use:           "create",
		Short:         "Create a UK8S cluster",
		Long:          "Create a UK8S (UCloud Kubernetes Service) cluster and, unless --async is set, wait for it to become RUNNING.",
		SilenceUsage:  true,
		SilenceErrors: false,
		Args:          cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := validateCreateCommon(req.Region, req.ProjectId, *req.ServiceCIDR); err != nil {
				return err
			}
			if err := validateCreateShape("master", *req.MasterCPU, *req.MasterMem); err != nil {
				return err
			}
			if err := validateCreateShape("node", *req.Nodes[0].CPU, *req.Nodes[0].Mem); err != nil {
				return err
			}
			if *req.Nodes[0].Count < 1 || *req.Nodes[0].Count > 10 {
				return fmt.Errorf("--node-count must be between 1 and 10")
			}
			if req.Nodes[0].IsolationGroup != nil && *req.Nodes[0].IsolationGroup != "" && *req.Nodes[0].Count > 8 {
				return fmt.Errorf("--node-count cannot exceed 8 when --node-isolation-group-id is set")
			}
			if err := validateCreateOptionalFields(cmd, req); err != nil {
				return err
			}
			if !strings.ContainsAny(*req.Password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") {
				return fmt.Errorf("--password must contain at least one uppercase letter")
			}
			if cmd.Flags().Changed("charge-type") && !oneOf(chargeType, "Dynamic", "Month", "Year") {
				return fmt.Errorf("--charge-type must be one of Dynamic, Month, or Year")
			}
			if !oneOf(*req.MasterMachineType, "N", "C", "O", "OS") {
				return fmt.Errorf("--master-machine-type must be one of N, C, O, or OS")
			}
			if !oneOf(*req.Nodes[0].MachineType, "N", "C", "G", "O", "OS") {
				return fmt.Errorf("--node-machine-type must be one of N, C, G, O, or OS")
			}
			if *req.Nodes[0].MachineType == "G" {
				if !cmd.Flags().Changed("node-gpu") || !cmd.Flags().Changed("node-gpu-type") {
					return fmt.Errorf("--node-gpu and --node-gpu-type are required when --node-machine-type is G")
				}
			} else if cmd.Flags().Changed("node-gpu") || cmd.Flags().Changed("node-gpu-type") {
				return fmt.Errorf("--node-gpu and --node-gpu-type require --node-machine-type G")
			}
			if err := bindEncodedValue(cmd, "user-data", userData, "user-data-base64", userDataB64, &req.UserData); err != nil {
				return err
			}
			if err := bindEncodedValue(cmd, "init-script", initScript, "init-script-base64", initScriptB64, &req.InitScript); err != nil {
				return err
			}
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
			if req.ImageId == nil || strings.TrimSpace(*req.ImageId) == "" {
				return fmt.Errorf("--image-id is required: choose a compatible image with 'ucloud uk8s image list'")
			}
			if cmd.Flags().Changed("charge-type") {
				req.ChargeType = sdk.String(chargeType)
			}
			if cmd.Flags().Changed("quantity") {
				if chargeType == "Dynamic" {
					return fmt.Errorf("--quantity must not be set when --charge-type is Dynamic")
				}
				req.Quantity = sdk.Int(quantity)
			}
			optional := map[string]func(){
				"k8s-version":              func() { req.K8sVersion = nil },
				"external-api-server":      func() { req.ExternalApiServer = nil },
				"cluster-domain":           func() { req.ClusterDomain = nil },
				"master-boot-disk-type":    func() { req.MasterBootDiskType = nil },
				"master-boot-disk-size-gb": func() { req.MasterBootDiskSize = nil },
				"master-data-disk-type":    func() { req.MasterDataDiskType = nil },
				"master-data-disk-size-gb": func() { req.MasterDataDiskSize = nil },
				"master-cpu-platform":      func() { req.MasterMinimalCpuPlatform = nil },
				"node-boot-disk-type":      func() { req.Nodes[0].BootDiskType = nil },
				"node-boot-disk-size-gb":   func() { req.Nodes[0].BootDiskSIze = nil },
				"node-data-disk-type":      func() { req.Nodes[0].DataDiskType = nil },
				"node-data-disk-size-gb":   func() { req.Nodes[0].DataDiskSize = nil },
				"node-cpu-platform":        func() { req.Nodes[0].MinimalCpuPlatform = nil },
				"node-max-pods":            func() { req.Nodes[0].MaxPods = nil },
				"node-isolation-group-id":  func() { req.Nodes[0].IsolationGroup = nil },
				"node-labels":              func() { req.Nodes[0].Labels = nil },
				"node-taints":              func() { req.Nodes[0].Taints = nil },
				"node-gpu":                 func() { req.Nodes[0].GPU = nil },
				"node-gpu-type":            func() { req.Nodes[0].GpuType = nil },
				"group":                    func() { req.Tag = nil },
			}
			for name, clear := range optional {
				if !cmd.Flags().Changed(name) {
					clear()
				}
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
				IsolationGroup:     req.Nodes[0].IsolationGroup,
				Labels:             req.Nodes[0].Labels,
				Taints:             req.Nodes[0].Taints,
				GPU:                req.Nodes[0].GPU,
				GpuType:            req.Nodes[0].GpuType,
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
	req.K8sVersion = flags.String("k8s-version", "", "Optional. Kubernetes version. If omitted, the service selects the highest supported version.")
	req.Password = flags.String("password", "", "Required. Password for cluster nodes (Master + Node). Must include at least one uppercase letter. Base64-encoded automatically before submission.")

	// --- Required: network ---
	req.VPCId = flags.String("vpc-id", "", "Required. VPC ID. See 'ucloud vpc list'.")
	req.SubnetId = flags.String("subnet-id", "", "Required. Subnet ID where nodes and pods live. See 'ucloud subnet list'.")
	req.ServiceCIDR = flags.String("service-cidr", "", "Required. Service CIDR for ClusterIP allocation (e.g. 172.17.0.0/16). Must not overlap with the VPC CIDR.")

	// --- Required: master ---
	req.MasterCPU = flags.Int("master-cpu", 0, "Required. vCPU cores per Master node. Range [2, 64].")
	req.MasterMem = flags.Int("master-memory-mb", 0, "Required. Memory per Master node. Unit: MB. Range [4096, 262144], multiple of 1024.")
	req.MasterMachineType = flags.String("master-machine-type", "", "Required. Master machine type. One of N, C, O, OS.")
	flags.StringSliceVar(&masterZones, "master-zone", nil, "Required. Availability zone(s) for the 3 Master nodes. Pass 1 zone (replicated 3x) or 3 zones for multi-AZ HA.")

	// --- Required: nodes (first/only group) ---
	req.Nodes[0].CPU = flags.Int("node-cpu", 0, "Required. vCPU cores per Node. Range [2, 64].")
	req.Nodes[0].Count = flags.Int("node-count", 0, "Required. Node count per group. Range [1, 10].")
	req.Nodes[0].Mem = flags.Int("node-memory-mb", 0, "Required. Memory per Node. Unit: MB. Range [4096, 262144], multiple of 1024.")
	req.Nodes[0].MachineType = flags.String("node-machine-type", "", "Required. Node machine type. One of N, C, G, O, OS.")
	flags.StringVar(&nodeZone, "node-zone", "", "Required. Availability zone for the node group.")

	// --- Optional: image / extras ---
	req.ImageId = flags.String("image-id", "", "Required. Compatible UK8S UHost image ID for Master and Node. See 'ucloud uk8s image list'.")
	req.ExternalApiServer = flags.String("external-api-server", "", "Optional. Expose the API server on the public internet. Accept values: Yes, No.")
	req.ClusterDomain = flags.String("cluster-domain", "", "Optional. Custom cluster domain.")
	flags.BoolVar(&async, "async", false, "Optional. Do not wait for the cluster to become RUNNING.")

	// Master disk / platform (optional)
	req.MasterBootDiskType = flags.String("master-boot-disk-type", "", "Optional. Master system disk type. See uhost disk types.")
	req.MasterBootDiskSize = flags.Int("master-boot-disk-size-gb", 0, "Optional. Master system disk size in GB. Range [40, 500].")
	req.MasterDataDiskType = flags.String("master-data-disk-type", "", "Optional. Master data disk type.")
	req.MasterDataDiskSize = flags.Int("master-data-disk-size-gb", 0, "Optional. Master data disk size in GB. Range [20, 1000].")
	req.MasterMinimalCpuPlatform = flags.String("master-cpu-platform", "", "Optional. Minimum CPU platform. E.g. Intel/Cascadelake.")

	// Node disk / platform (optional)
	req.Nodes[0].BootDiskType = flags.String("node-boot-disk-type", "", "Optional. Node system disk type.")
	req.Nodes[0].BootDiskSIze = flags.Int("node-boot-disk-size-gb", 0, "Optional. Node system disk size in GB. Range [40, 500].")
	req.Nodes[0].DataDiskType = flags.String("node-data-disk-type", "", "Optional. Node data disk type.")
	req.Nodes[0].DataDiskSize = flags.Int("node-data-disk-size-gb", 0, "Optional. Node data disk size in GB. Range [20, 1000].")
	req.Nodes[0].MinimalCpuPlatform = flags.String("node-cpu-platform", "", "Optional. Minimum CPU platform.")
	req.Nodes[0].MaxPods = flags.Int("node-max-pods", 0, "Optional. Maximum pods per node.")
	req.Nodes[0].IsolationGroup = flags.String("node-isolation-group-id", "", "Optional. Isolation group ID for Node instances; one group supports at most 8 nodes.")
	req.Nodes[0].Labels = flags.String("node-labels", "", "Optional. Comma-separated node labels in key=value form, at most 5.")
	req.Nodes[0].Taints = flags.String("node-taints", "", "Optional. Comma-separated node taints in key=value:effect form, at most 5.")
	req.Nodes[0].GPU = flags.Int("node-gpu", 0, "Optional. GPU core count; supported only by GPU-capable machine types.")
	req.Nodes[0].GpuType = flags.String("node-gpu-type", "", "Optional. GPU type: K80, P40, or V100.")

	// Cluster customization. Plain values are base64-encoded by the CLI; the
	// *-base64 variants accept already encoded data and conflict with plain input.
	flags.StringVar(&userData, "user-data", "", "Optional. User data; base64-encoded automatically. Maximum decoded size: 16 KiB.")
	flags.StringVar(&userDataB64, "user-data-base64", "", "Optional. Pre-encoded user data. Conflicts with --user-data.")
	flags.StringVar(&initScript, "init-script", "", "Optional. Post-install script; base64-encoded automatically. Maximum decoded size: 16 KiB.")
	flags.StringVar(&initScriptB64, "init-script-base64", "", "Optional. Pre-encoded post-install script. Conflicts with --init-script.")
	req.Tag = flags.String("group", "", "Optional. Business group.")

	// kube-proxy
	flags.StringVar(&kubeProxyMode, "kube-proxy-mode", "", "Optional. kube-proxy mode. Accept values: iptables, ipvs.")

	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)
	flags.StringVar(&chargeType, "charge-type", "", "Optional. Billing mode: Dynamic, Month, or Year.")
	flags.IntVar(&quantity, "quantity", 0, "Optional. Purchase duration.")

	// Static enum candidates.
	command.SetFlagValues(cmd, "master-machine-type", "N", "C", "O", "OS")
	command.SetFlagValues(cmd, "node-machine-type", "N", "C", "G", "O", "OS")
	command.SetFlagValues(cmd, "external-api-server", "Yes", "No")
	command.SetFlagValues(cmd, "kube-proxy-mode", "iptables", "ipvs")
	command.SetFlagValues(cmd, "master-boot-disk-type", "CLOUD_SSD", "CLOUD_NORMAL", "LOCAL_SSD", "LOCAL_NORMAL", "CLOUD_RSSD", "EXCLUSIVE_LOCAL_DISK")
	command.SetFlagValues(cmd, "master-data-disk-type", "CLOUD_SSD", "CLOUD_NORMAL", "LOCAL_SSD", "LOCAL_NORMAL", "CLOUD_RSSD", "EXCLUSIVE_LOCAL_DISK", "")
	command.SetFlagValues(cmd, "node-boot-disk-type", "CLOUD_SSD", "CLOUD_NORMAL", "LOCAL_SSD", "LOCAL_NORMAL", "CLOUD_RSSD", "EXCLUSIVE_LOCAL_DISK")
	command.SetFlagValues(cmd, "node-data-disk-type", "CLOUD_SSD", "CLOUD_NORMAL", "LOCAL_SSD", "LOCAL_NORMAL", "CLOUD_RSSD", "EXCLUSIVE_LOCAL_DISK", "")
	command.SetFlagValues(cmd, "master-cpu-platform", "Intel/Auto", "Intel/IvyBridge", "Intel/Haswell", "Intel/Broadwell", "Intel/Skylake", "Intel/Cascadelake")
	command.SetFlagValues(cmd, "node-cpu-platform", "Intel/Auto", "Intel/IvyBridge", "Intel/Haswell", "Intel/Broadwell", "Intel/Skylake", "Intel/Cascadelake")
	command.SetFlagValues(cmd, "charge-type", "Dynamic", "Month", "Year")
	command.SetFlagValues(cmd, "node-gpu-type", "K80", "P40", "V100")

	// Dynamic completion for resource ids.
	command.SetCompletion(cmd, "vpc-id", func() []string {
		return listVPCIDs(ctx, derefStr(req.ProjectId), derefStr(req.Region))
	})
	command.SetCompletion(cmd, "subnet-id", func() []string {
		return listSubnetIDs(ctx, ctx.PickResourceID(*req.VPCId), derefStr(req.ProjectId), derefStr(req.Region))
	})
	command.SetCompletion(cmd, "k8s-version", func() []string {
		return listUK8SVersions(ctx, derefStr(req.ProjectId), derefStr(req.Region))
	})
	command.SetCompletion(cmd, "image-id", func() []string {
		return listUK8SImageIDs(ctx, derefStr(req.ProjectId), derefStr(req.Region), nodeZone)
	})
	command.SetCompletion(cmd, "node-isolation-group-id", func() []string {
		return listIsolationGroupIDs(ctx, derefStr(req.ProjectId), derefStr(req.Region))
	})

	cmd.MarkFlagRequired("name")
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
	cmd.MarkFlagRequired("image-id")

	return cmd
}

func validateCreateCommon(region, projectID *string, serviceCIDR string) error {
	if region == nil || strings.TrimSpace(*region) == "" {
		return fmt.Errorf("region is required; set --region or configure it in the active profile")
	}
	if projectID == nil || strings.TrimSpace(*projectID) == "" {
		return fmt.Errorf("project ID is required; set --project-id or configure it in the active profile")
	}
	if _, _, err := net.ParseCIDR(serviceCIDR); err != nil {
		return fmt.Errorf("--service-cidr must be a valid CIDR: %w", err)
	}
	return nil
}

func validateCreateShape(prefix string, cpu, memory int) error {
	if cpu < 2 || cpu > 64 {
		return fmt.Errorf("--%s-cpu must be between 2 and 64", prefix)
	}
	if memory < 4096 || memory > 262144 || memory%1024 != 0 {
		return fmt.Errorf("--%s-memory-mb must be between 4096 and 262144 and a multiple of 1024", prefix)
	}
	return nil
}

func validateCreateOptionalFields(cmd *cobra.Command, req *uk8ssdk.CreateUK8SClusterV2Request) error {
	if cmd.Flags().Changed("external-api-server") && !oneOf(*req.ExternalApiServer, "Yes", "No") {
		return fmt.Errorf("--external-api-server must be Yes or No")
	}
	if mode, err := cmd.Flags().GetString("kube-proxy-mode"); err == nil && cmd.Flags().Changed("kube-proxy-mode") && !oneOf(mode, "iptables", "ipvs") {
		return fmt.Errorf("--kube-proxy-mode must be iptables or ipvs")
	}
	if cmd.Flags().Changed("node-gpu-type") && !oneOf(*req.Nodes[0].GpuType, "K80", "P40", "V100") {
		return fmt.Errorf("--node-gpu-type must be one of K80, P40, or V100")
	}
	if cmd.Flags().Changed("node-gpu") && *req.Nodes[0].GPU < 1 {
		return fmt.Errorf("--node-gpu must be greater than 0")
	}
	if cmd.Flags().Changed("node-max-pods") && *req.Nodes[0].MaxPods < 1 {
		return fmt.Errorf("--node-max-pods must be greater than 0")
	}
	for _, item := range []struct {
		name     string
		value    *int
		min, max int
	}{
		{"master-boot-disk-size-gb", req.MasterBootDiskSize, 40, 500},
		{"node-boot-disk-size-gb", req.Nodes[0].BootDiskSIze, 40, 500},
		{"master-data-disk-size-gb", req.MasterDataDiskSize, 20, 1000},
		{"node-data-disk-size-gb", req.Nodes[0].DataDiskSize, 20, 1000},
	} {
		if cmd.Flags().Changed(item.name) && (*item.value < item.min || *item.value > item.max) {
			return fmt.Errorf("--%s must be between %d and %d", item.name, item.min, item.max)
		}
	}
	for _, item := range []struct {
		name  string
		value *string
	}{
		{"node-labels", req.Nodes[0].Labels},
		{"node-taints", req.Nodes[0].Taints},
	} {
		if item.value != nil && *item.value != "" && len(strings.Split(*item.value, ",")) > 5 {
			return fmt.Errorf("--%s accepts at most 5 comma-separated entries", item.name)
		}
	}
	return nil
}

func bindEncodedValue(cmd *cobra.Command, plainName, plain, encodedName, encoded string, target **string) error {
	if plain != "" && encoded != "" {
		return fmt.Errorf("--%s conflicts with --%s", plainName, encodedName)
	}
	if plain != "" {
		if len([]byte(plain)) > 16*1024 {
			return fmt.Errorf("--%s must not exceed 16 KiB", plainName)
		}
		*target = sdk.String(base64.StdEncoding.EncodeToString([]byte(plain)))
		return nil
	}
	if encoded != "" {
		if !common.IsBase64Encoded([]byte(encoded)) {
			return fmt.Errorf("--%s must be base64-encoded", encodedName)
		}
		decoded, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil || len(decoded) > 16*1024 {
			return fmt.Errorf("--%s decoded value must not exceed 16 KiB", encodedName)
		}
		*target = sdk.String(encoded)
		return nil
	}
	if !cmd.Flags().Changed(plainName) && !cmd.Flags().Changed(encodedName) {
		*target = nil
	}
	return nil
}

func oneOf(value string, allowed ...string) bool {
	for _, candidate := range allowed {
		if value == candidate {
			return true
		}
	}
	return false
}
