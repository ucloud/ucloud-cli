package uhadoop

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/ucloud/request"
	"github.com/ucloud/ucloud-sdk-go/ucloud/response"

	uhadoopsdk "github.com/ucloud/ucloud-sdk-go/services/uhadoop"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

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

type createResponse struct {
	response.CommonBase
	InstanceId string `json:"InstanceId"`
}

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

var hdfsVersionByFramework = map[string]string{
	"3.3.4-udh3.2":    "Hdfs#3.3.4",
	"3.3.4-udh3.1":    "Hdfs#3.3.4",
	"3.2.1-udh3.0":    "Hdfs#3.2.1",
	"2.8.5-udh2.2":    "Hdfs#2.8.5",
	"2.6.0-cdh5.13.3": "Hdfs#2.6.0",
	"2.6.0-cdh5.4.9":  "Hdfs#2.6.0",
}

func newCreate(ctx *cli.Context) *cobra.Command {
	var (
		async       *bool
		rawPassword string
		clusterCase string
		master      instanceGroupConfig
		core        instanceGroupConfig
		task        instanceGroupConfig
	)
	client := cli.NewServiceClient(ctx, uhadoopsdk.NewClient)
	sdkReq := client.NewCreateUHadoopInstanceRequest()
	cmd := &cobra.Command{
		Use:          "create",
		Short:        "Create a UHadoop cluster",
		Long:         `Create a UHadoop cluster with specified configuration`,
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, args []string) {
			w := ctx.ProgressWriter()

			if sdkReq.Framework != nil && (*sdkReq.Framework == "StarRocks-Shared-Nothing" || *sdkReq.Framework == "StarRocks-Shared-Data") {
				if !cmd.Flags().Changed("master-count") {
					master.Count = 3
				}
			}
			groups := buildGroups(master, core, task)
			if len(groups) == 0 {
				ctx.HandleError(fmt.Errorf("at least one node group is required"))
				return
			}

			appConfigs := sdkReq.AppConfigs
			if clusterCase != "" {
				versionMap, ok := clusterCaseApps[*sdkReq.FrameworkVersion]
				if !ok {
					ctx.HandleError(fmt.Errorf("unsupported framework-version %q for --cluster-case", *sdkReq.FrameworkVersion))
					return
				}
				template, ok := versionMap[clusterCase]
				if !ok {
					ctx.HandleError(fmt.Errorf("no cluster-case %q for framework-version %q", clusterCase, *sdkReq.FrameworkVersion))
					return
				}
				appConfigs = append([]string(nil), template...)
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
			if resp.RetCode != 0 {
				ctx.HandleError(fmt.Errorf("[%d] %s", resp.RetCode, resp.Message))
				return
			}
			text := fmt.Sprintf("uhadoop[%s] is creating", resp.InstanceId)
			if *async {
				fmt.Fprintln(w, text)
			} else {
				ctx.PollerTo(w, describeClusterForPoll(ctx, client), cli.WithTimeout(60*time.Minute)).Spoll(resp.InstanceId, text, []string{stateRunning})
			}
			ctx.EmitResult(cli.OpResultRow{ResourceID: resp.InstanceId, Action: "create", Status: "Creating"})
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	sdkReq.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	sdkReq.Zone = flags.String("zone", "", "Optional. Assign availability zone")
	sdkReq.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	sdkReq.InstanceName = flags.String("name", "", "Required. Instance name")
	sdkReq.Framework = flags.String("framework", "", "Required. Framework")
	sdkReq.FrameworkVersion = flags.String("framework-version", "", "Required. Framework version")
	flags.StringVar(&rawPassword, "password", "", "Required. Login password")
	sdkReq.VPCId = flags.String("vpc-id", "", "Optional. VPC ID")
	sdkReq.SubnetId = flags.String("subnet-id", "", "Optional. Subnet ID")
	sdkReq.ChargeType = flags.String("charge-type", "Month", "Optional. Charge type")
	sdkReq.Quantity = flags.Int("quantity", 1, "Optional. Quantity")
	sdkReq.BusinessId = flags.String("business-id", "Default", "Optional. Business group")
	sdkReq.StorgeClusterId = flags.String("storage-cluster-id", "", "Optional. Storage cluster ID (MR framework)")
	sdkReq.StandAloneMetaStore = flags.String("meta-store", "", "Optional. Meta store type")
	sdkReq.IsSecurityEnabled = flags.String("security-enabled", "", "Optional. Enable security group")
	sdkReq.SecGroupIds = flags.String("sec-group-ids", "", "Optional. Security group IDs")
	sdkReq.US3Bucket = flags.String("us3-bucket", "", "Optional. US3 bucket")
	sdkReq.US3AccessKey = flags.String("us3-access-key", "", "Optional. US3 access key")
	sdkReq.US3SecretKey = flags.String("us3-secret-key", "", "Optional. US3 secret key")
	sdkReq.US3TokenName = flags.String("us3-token-name", "", "Optional. US3 token name")
	flags.StringVar(&clusterCase, "cluster-case", "", "Cluster use case: Spark|Hbase|Core-Hadoop")

	master.NodeRole = "master"
	flags.StringVar(&master.NodeType, "master-node-type", "o.hadoop4m.xlarge", "Master node type")
	flags.IntVar(&master.Count, "master-count", 2, "Master node count")
	flags.IntVar(&master.DataDiskSize, "master-data-disk-size", 100, "Master data disk GB")
	flags.IntVar(&master.DataDiskNum, "master-data-disk-num", 1, "Master data disk num")
	flags.StringVar(&master.DataDiskType, "master-data-disk-type", "CLOUD_RSSD", "Master data disk type")
	flags.IntVar(&master.BootDiskSize, "master-boot-disk-size", 50, "Master boot disk GB")
	flags.StringVar(&master.BootDiskType, "master-boot-disk-type", "CLOUD_RSSD", "Master boot disk type")

	core.NodeRole = "core"
	flags.StringVar(&core.NodeType, "core-node-type", "o.hadoop2m.xlarge", "Core node type")
	flags.IntVar(&core.Count, "core-count", 3, "Core node count")
	flags.IntVar(&core.DataDiskSize, "core-data-disk-size", 200, "Core data disk GB")
	flags.IntVar(&core.DataDiskNum, "core-data-disk-num", 1, "Core data disk num")
	flags.StringVar(&core.DataDiskType, "core-data-disk-type", "CLOUD_RSSD", "Core data disk type")
	flags.IntVar(&core.BootDiskSize, "core-boot-disk-size", 50, "Core boot disk GB")
	flags.StringVar(&core.BootDiskType, "core-boot-disk-type", "CLOUD_RSSD", "Core boot disk type")

	task.NodeRole = "task"
	flags.StringVar(&task.NodeType, "task-node-type", "o.hadoop2m.xlarge", "Optional. Task node type")
	flags.IntVar(&task.Count, "task-count", 0, "Task node count")
	flags.IntVar(&task.DataDiskSize, "task-data-disk-size", 200, "Task data disk GB")
	flags.IntVar(&task.DataDiskNum, "task-data-disk-num", 1, "Task data disk num")
	flags.StringVar(&task.DataDiskType, "task-data-disk-type", "CLOUD_RSSD", "Task data disk type")
	flags.IntVar(&task.BootDiskSize, "task-boot-disk-size", 50, "Task boot disk GB")
	flags.StringVar(&task.BootDiskType, "task-boot-disk-type", "CLOUD_RSSD", "Task boot disk type")

	async = flags.Bool("async", false, "Optional. Do not wait for creation to finish")
	flags.StringSliceVar(&sdkReq.AppConfigs, "app-config", nil, "App configs: App#Version")

	command.SetFlagValues(cmd, "cluster-case", "Spark", "Hbase", "Core-Hadoop")
	command.SetFlagValues(cmd, "charge-type", "Month", "Year", "Dynamic")
	command.SetFlagValues(cmd, "security-enabled", "true", "false")

	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("framework")
	cmd.MarkFlagRequired("framework-version")
	cmd.MarkFlagRequired("password")

	cmd.MarkFlagRequired("region")
	cmd.MarkFlagRequired("zone")
	return cmd
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
