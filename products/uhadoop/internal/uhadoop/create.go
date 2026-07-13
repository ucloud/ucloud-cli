package uhadoop

import (
	"encoding/base64"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/ucloud/request"
	"github.com/ucloud/ucloud-sdk-go/ucloud/response"

	uhadoopsdk "github.com/ucloud/ucloud-sdk-go/services/uhadoop"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// instanceGroupConfig matches the real API's InstanceGroupConfigs struct.
type instanceGroupConfig struct {
	NodeRole     string `json:"NodeRole"`
	NodeType     string `json:"NodeType"`
	Count        int    `json:"Count"`
	DataDiskSize int    `json:"DataDiskSize"`
	DataDiskNum  int    `json:"DataDiskNum"`
	DataDiskType string `json:"DataDiskType"`
	BootDiskSize int    `json:"BootDiskSize"`
	BootDiskType string `json:"BootDiskType"`
}

// createRequest is the real request body for CreateUHadoopInstance.
type createRequest struct {
	request.CommonBase
	InstanceName         string                `json:"InstanceName"`
	Framework            string                `json:"Framework"`
	FrameworkVersion     string                `json:"FrameworkVersion"`
	Password             string                `json:"Password"`
	VPCId                string                `json:"VPCId"`
	SubnetId             string                `json:"SubnetId"`
	ChargeType           string                `json:"ChargeType,omitempty"`
	Quantity             int                   `json:"Quantity,omitempty"`
	BusinessId           string                `json:"BusinessId,omitempty"`
	StorgeClusterId      string                `json:"StorgeClusterId,omitempty"`
	StandAloneMetaStore  string                `json:"StandAloneMetaStore,omitempty"`
	IsSecurityEnabled    string                `json:"IsSecurityEnabled,omitempty"`
	SecGroupIds          string                `json:"SecGroupIds,omitempty"`
	US3Bucket            string                `json:"US3Bucket,omitempty"`
	US3AccessKey         string                `json:"US3AccessKey,omitempty"`
	US3SecretKey         string                `json:"US3SecretKey,omitempty"`
	US3TokenName         string                `json:"US3TokenName,omitempty"`
	AppConfigs           []string              `json:"AppConfigs"`
	InstanceGroupConfigs []instanceGroupConfig `json:"InstanceGroupConfigs"`
}

// createResponse is the response for CreateUHadoopInstance.
type createResponse struct {
	response.CommonBase
	InstanceId string `json:"InstanceId"`
	Message    string `json:"Message"`
}

// clusterCaseApps maps (FrameworkVersion, ClusterCase) to app templates.
// "Hadoop" cluster cases: Spark, Hbase, Core Hadoop.
// For "Hadoop" framework, Hdfs is auto-appended. For "MR", it is not.
var clusterCaseApps = map[string]map[string][]string{
	"3.3.4-udh3.2": {
		"Spark":       {"Spark#3.5.3", "Hive#3.1.3", "Hue#4.11.0", "Zookeeper#3.8.4", "Mysql#8.0.32", "Yarn#3.3.4"},
		"Hbase":       {"Hbase#2.4.18", "Hue#4.11.0", "Zookeeper#3.8.4", "Mysql#8.0.32", "Yarn#3.3.4", "Phoenix#5.2.1"},
		"Core-Hadoop": {"Hive#3.1.3", "Hue#4.11.0", "Zookeeper#3.8.4", "Yarn#3.3.4"},
	},
	"3.3.4-udh3.1": {
		"Spark":       {"Spark#3.5.3", "Hive#3.1.3", "Hue#4.11.0", "Zookeeper#3.8.4", "Mysql#8.0.32", "Yarn#3.3.4"},
		"Hbase":       {"Hbase#2.4.18", "Hue#4.11.0", "Zookeeper#3.8.4", "Mysql#8.0.32", "Yarn#3.3.4", "Phoenix#5.2.1"},
		"Core-Hadoop": {"Hive#3.1.3", "Hue#4.11.0", "Zookeeper#3.8.4", "Yarn#3.3.4"},
	},
	"3.2.1-udh3.0": {
		"Spark":       {"Spark#3.3.0", "Hive#3.1.3", "Hue#4.7.1", "Zookeeper#3.6.3", "Mysql#5.6.47", "Yarn#3.2.1"},
		"Hbase":       {"Hbase#2.2.4", "Hue#4.7.1", "Zookeeper#3.6.3", "Mysql#5.6.47", "Yarn#3.2.1", "Phoenix#5.1.2"},
		"Core-Hadoop": {"Hive#3.1.3", "Hue#4.7.1", "Zookeeper#3.6.3", "Yarn#3.2.1"},
	},
	"2.8.5-udh2.2": {
		"Spark":       {"Spark#2.4.6", "Hive#2.3.6", "Hue#4.7.1", "Zookeeper#3.4.13", "Mysql#5.6.47", "Yarn#2.8.5"},
		"Hbase":       {"Hbase#1.4.10", "Hue#4.7.1", "Zookeeper#3.4.13", "Mysql#5.6.47", "Yarn#2.8.5", "Phoenix#4.14.3"},
		"Core-Hadoop": {"Hive#2.3.6", "Hue#4.7.1", "Zookeeper#3.4.13", "Yarn#2.8.5"},
	},
	"2.6.0-cdh5.13.3": {
		"Spark":       {"Spark#2.4.3", "Hive#2.3.3", "Hue#3.10.0", "Zookeeper#3.4.5", "Mysql#5.1.73", "Yarn#2.6.0"},
		"Hbase":       {"Hbase#1.2.0", "Hue#3.10.0", "Zookeeper#3.4.5", "Mysql#5.1.73", "Yarn#2.6.0", "Phoenix#4.14.0"},
		"Core-Hadoop": {"Hive#2.3.3", "Hue#3.10.0", "Zookeeper#3.4.5", "Yarn#2.6.0"},
	},
	"2.6.0-cdh5.4.9": {
		"Spark":       {"Spark#2.0.1", "Hive#1.2.1", "Hue#3.10.0", "Zookeeper#3.4.5", "Mysql#5.1.73", "Yarn#2.6.0"},
		"Hbase":       {"Hbase#1.0.0", "Hue#3.10.0", "Zookeeper#3.4.5", "Mysql#5.1.73", "Yarn#2.6.0", "Phoenix#4.6.0"},
		"Core-Hadoop": {"Hive#1.2.1", "Hue#3.10.0", "Zookeeper#3.4.5", "Yarn#2.6.0"},
	},
}

// hdfsVersionByFramework maps framework-version to the Hdfs app version for Hadoop framework.
var hdfsVersionByFramework = map[string]string{
	"3.3.4-udh3.2":    "Hdfs#3.3.4",
	"3.3.4-udh3.1":    "Hdfs#3.3.4",
	"3.2.1-udh3.0":    "Hdfs#3.2.1",
	"2.8.5-udh2.2":    "Hdfs#2.8.5",
	"2.6.0-cdh5.13.3": "Hdfs#2.6.0",
	"2.6.0-cdh5.4.9":  "Hdfs#2.6.0",
}

// newCreate ucloud uhadoop create
func newCreate(ctx *cli.Context) *cobra.Command {
	var (
		rawPassword string
		clusterCase string
		master      instanceGroupConfig
		core        instanceGroupConfig
		task        instanceGroupConfig
	)
	client := cli.NewServiceClient(ctx, uhadoopsdk.NewClient)
	sdkReq := client.NewCreateUHadoopInstanceRequest()
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a UHadoop cluster",
		Long:  `Create a UHadoop cluster with specified configuration`,
		Run: func(cmd *cobra.Command, args []string) {
			if sdkReq.Framework != nil && (*sdkReq.Framework == "StarRocks-Shared-Nothing" || *sdkReq.Framework == "StarRocks-Shared-Data") {
				if !cmd.Flags().Changed("master-count") {
					master.Count = 3
				}
			}
			groups := buildGroups(master, core, task)
			if len(groups) == 0 {
				ctx.HandleError(fmt.Errorf("at least one node group (--master-node-type or --core-node-type) is required"))
				return
			}

			// Resolve app configs: --cluster-case overrides --app-config
			appConfigs := sdkReq.AppConfigs
			if clusterCase != "" {
				if sdkReq.FrameworkVersion == nil || *sdkReq.FrameworkVersion == "" {
					ctx.HandleError(fmt.Errorf("--framework-version is required with --cluster-case"))
					return
				}
				versionMap, ok := clusterCaseApps[*sdkReq.FrameworkVersion]
				if !ok {
					ctx.HandleError(fmt.Errorf("unsupported framework-version %q for --cluster-case, supported: %v",
						*sdkReq.FrameworkVersion, supportedVersions()))
					return
				}
				template, ok := versionMap[clusterCase]
				if !ok {
					ctx.HandleError(fmt.Errorf("no cluster-case %q for framework-version %q", clusterCase, *sdkReq.FrameworkVersion))
					return
				}
				appConfigs = append([]string(nil), template...)

				// For Hadoop framework, auto-append Hdfs
				if sdkReq.Framework != nil && *sdkReq.Framework == "Hadoop" {
					if hdfsVer, ok := hdfsVersionByFramework[*sdkReq.FrameworkVersion]; ok {
						appConfigs = append(appConfigs, hdfsVer)
					}
				}
			}

			req := &createRequest{
				InstanceName:         *sdkReq.InstanceName,
				Framework:            *sdkReq.Framework,
				FrameworkVersion:     *sdkReq.FrameworkVersion,
				Password:             base64.StdEncoding.EncodeToString([]byte(rawPassword)),
				VPCId:                *sdkReq.VPCId,
				SubnetId:             *sdkReq.SubnetId,
				AppConfigs:           appConfigs,
				InstanceGroupConfigs: groups,
			}
			req.Region = sdkReq.Region
			req.Zone = sdkReq.Zone
			req.ProjectId = sdkReq.ProjectId
			if sdkReq.ChargeType != nil {
				req.ChargeType = *sdkReq.ChargeType
			}
			if sdkReq.Quantity != nil {
				req.Quantity = *sdkReq.Quantity
			}
			if sdkReq.BusinessId != nil {
				req.BusinessId = *sdkReq.BusinessId
			}
			if sdkReq.StorgeClusterId != nil {
				req.StorgeClusterId = *sdkReq.StorgeClusterId
			}
			if sdkReq.StandAloneMetaStore != nil {
				req.StandAloneMetaStore = *sdkReq.StandAloneMetaStore
			}
			if sdkReq.IsSecurityEnabled != nil {
				req.IsSecurityEnabled = *sdkReq.IsSecurityEnabled
			}
			if sdkReq.SecGroupIds != nil {
				req.SecGroupIds = *sdkReq.SecGroupIds
			}
			if sdkReq.US3Bucket != nil {
				req.US3Bucket = *sdkReq.US3Bucket
			}
			if sdkReq.US3AccessKey != nil {
				req.US3AccessKey = *sdkReq.US3AccessKey
			}
			if sdkReq.US3SecretKey != nil {
				req.US3SecretKey = *sdkReq.US3SecretKey
			}
			if sdkReq.US3TokenName != nil {
				req.US3TokenName = *sdkReq.US3TokenName
			}

			var resp createResponse
			err := client.InvokeAction("CreateUHadoopInstance", req, &resp)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			ctx.PrintJSON(resp)
		},
	}
	cmd.Flags().SortFlags = false

	ctx.BindRegion(cmd, sdkReq)
	ctx.BindZone(cmd, sdkReq)
	ctx.BindProjectID(cmd, sdkReq)
	sdkReq.InstanceName = cmd.Flags().String("name", "", "Required. Instance name")
	sdkReq.Framework = cmd.Flags().String("framework", "", "Required. Framework: Hadoop|HDFS|MR|StarRocks-Shared-Nothing|StarRocks-Shared-Data")
	sdkReq.FrameworkVersion = cmd.Flags().String("framework-version", "", "Required. Framework version, e.g. 3.3.4-udh3.2")
	cmd.Flags().StringVar(&rawPassword, "password", "", "Required. Login password (plain text, auto base64 encoded)")
	sdkReq.VPCId = cmd.Flags().String("vpc-id", "", "Optional. VPC ID, auto-discovered if not set")
	sdkReq.SubnetId = cmd.Flags().String("subnet-id", "", "Optional. Subnet ID, auto-discovered if not set")
	sdkReq.ChargeType = cmd.Flags().String("charge-type", "Month", "Optional. Charge type: Month|Year|Dynamic")
	sdkReq.Quantity = cmd.Flags().Int("quantity", 1, "Optional. Quantity for charge type")
	sdkReq.BusinessId = cmd.Flags().String("business-id", "Default", "Optional. Business group ID")
	sdkReq.StorgeClusterId = cmd.Flags().String("storage-cluster-id", "", "Optional. Storage cluster ID for MR framework")
	sdkReq.StandAloneMetaStore = cmd.Flags().String("meta-store", "", "Optional. Stand alone meta store type, e.g. 'udb'")
	sdkReq.IsSecurityEnabled = cmd.Flags().String("security-enabled", "", "Optional. Enable security group: true|false")
	sdkReq.SecGroupIds = cmd.Flags().String("sec-group-ids", "", "Optional. Security group IDs, comma separated")
	sdkReq.US3Bucket = cmd.Flags().String("us3-bucket", "", "Optional. US3 bucket name for StarRocks-Shared-Data")
	sdkReq.US3AccessKey = cmd.Flags().String("us3-access-key", "", "Optional. US3 access key for StarRocks-Shared-Data")
	sdkReq.US3SecretKey = cmd.Flags().String("us3-secret-key", "", "Optional. US3 secret key for StarRocks-Shared-Data")
	sdkReq.US3TokenName = cmd.Flags().String("us3-token-name", "", "Optional. US3 token name for StarRocks-Shared-Data")

	// Cluster case flag — when set, auto-fills app-config based on the template
	cmd.Flags().StringVar(&clusterCase, "cluster-case", "", "Optional. Cluster use case: Spark|Hbase|Core-Hadoop. When set, --app-config is auto-filled from the template")

	// Master node group
	master.NodeRole = "master"
	cmd.Flags().StringVar(&master.NodeType, "master-node-type", "o.hadoop2m.xlarge", "Master node type, default o.hadoop2m.xlarge")
	cmd.Flags().IntVar(&master.Count, "master-count", 2, "Master node count, default 2")
	cmd.Flags().IntVar(&master.DataDiskSize, "master-data-disk-size", 100, "Master data disk size in GB, default 100")
	cmd.Flags().IntVar(&master.DataDiskNum, "master-data-disk-num", 1, "Master data disk number, default 1")
	cmd.Flags().StringVar(&master.DataDiskType, "master-data-disk-type", "CLOUD_RSSD", "Master data disk type, default CLOUD_RSSD")
	cmd.Flags().IntVar(&master.BootDiskSize, "master-boot-disk-size", 50, "Master boot disk size in GB, default 50")
	cmd.Flags().StringVar(&master.BootDiskType, "master-boot-disk-type", "CLOUD_RSSD", "Master boot disk type, default CLOUD_RSSD")

	// Core node group
	core.NodeRole = "core"
	cmd.Flags().StringVar(&core.NodeType, "core-node-type", "o.hadoop2m.xlarge", "Core node type, default o.hadoop2m.xlarge")
	cmd.Flags().IntVar(&core.Count, "core-count", 3, "Core node count, default 3")
	cmd.Flags().IntVar(&core.DataDiskSize, "core-data-disk-size", 200, "Core data disk size in GB, default 200")
	cmd.Flags().IntVar(&core.DataDiskNum, "core-data-disk-num", 1, "Core data disk number, default 1")
	cmd.Flags().StringVar(&core.DataDiskType, "core-data-disk-type", "CLOUD_RSSD", "Core data disk type, default CLOUD_RSSD")
	cmd.Flags().IntVar(&core.BootDiskSize, "core-boot-disk-size", 50, "Core boot disk size in GB, default 50")
	cmd.Flags().StringVar(&core.BootDiskType, "core-boot-disk-type", "CLOUD_RSSD", "Core boot disk type, default CLOUD_RSSD")

	// Task node group (optional)
	task.NodeRole = "task"
	cmd.Flags().StringVar(&task.NodeType, "task-node-type", "o.hadoop2m.xlarge", "Optional. Task node type, default o.hadoop2m.xlarge")
	cmd.Flags().IntVar(&task.Count, "task-count", 0, "Task node count, default 0 (no task nodes)")
	cmd.Flags().IntVar(&task.DataDiskSize, "task-data-disk-size", 200, "Task data disk size in GB, default 200")
	cmd.Flags().IntVar(&task.DataDiskNum, "task-data-disk-num", 1, "Task data disk number, default 1")
	cmd.Flags().StringVar(&task.DataDiskType, "task-data-disk-type", "CLOUD_RSSD", "Task data disk type, default CLOUD_RSSD")
	cmd.Flags().IntVar(&task.BootDiskSize, "task-boot-disk-size", 50, "Task boot disk size in GB, default 50")
	cmd.Flags().StringVar(&task.BootDiskType, "task-boot-disk-type", "CLOUD_RSSD", "Task boot disk type, default CLOUD_RSSD")

	cmd.Flags().StringSliceVar(&sdkReq.AppConfigs, "app-config", nil, "App configs in format App#Version. Ignored when --cluster-case is set")

	command.SetFlagValues(cmd, "cluster-case", "Spark", "Hbase", "Core-Hadoop")
	command.SetFlagValues(cmd, "charge-type", "Month", "Year", "Dynamic")
	command.SetFlagValues(cmd, "security-enabled", "true", "false")

	return cmd
}

func supportedVersions() []string {
	vs := make([]string, 0, len(clusterCaseApps))
	for v := range clusterCaseApps {
		vs = append(vs, v)
	}
	return vs
}

func buildGroups(master, core, task instanceGroupConfig) []instanceGroupConfig {
	var groups []instanceGroupConfig
	if master.NodeType != "" {
		groups = append(groups, master)
	}
	if core.NodeType != "" {
		groups = append(groups, core)
	}
	if task.NodeType != "" && task.Count > 0 {
		groups = append(groups, task)
	}
	return groups
}
