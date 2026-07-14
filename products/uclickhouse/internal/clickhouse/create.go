package clickhouse

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	uclickhousesdk "github.com/ucloud/ucloud-sdk-go/services/uclickhouse"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newCreate ucloud clickhouse create
func newCreate(ctx *cli.Context) *cobra.Command {
	var adminPassword *string
	var async *bool
	var labels []string
	client := cli.NewServiceClient(ctx, uclickhousesdk.NewClient)
	req := client.NewCreateUClickhouseClusterRequest()
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create UClickhouse cluster",
		Long:  "Create UClickhouse cluster",
		Args:  validateCreateArgs,
		Run: func(cmd *cobra.Command, args []string) {
			parsedLabels, err := parseLabels(labels)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			req.Labels = parsedLabels
			req.AdminPassword = sdk.String(base64.StdEncoding.EncodeToString([]byte(*adminPassword)))

			w := ctx.ProgressWriter()
			resp, err := createUClickhouseCluster(client, req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			id := resp.Data.ClusterId
			text := fmt.Sprintf("clickhouse[%s] is creating", id)
			if *async {
				fmt.Fprintln(w, text)
			} else {
				ctx.PollerTo(w, describeClusterByID(ctx)).Spoll(id, text, []string{STATUS_RUNNING, STATUS_CREATE_FAILED})
			}
			ctx.EmitResult(cli.OpResultRow{ResourceID: id, Action: "create", Status: "Creating"})
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.ClusterName = flags.String("name", "clickhouse", "Optional. Cluster name, default clickhouse")
	req.ClickhouseMachineTypeId = flags.String("clickhouse-machine-type-id", "", "Required. ClickHouse machine type ID")
	req.DataDiskType = flags.String("data-disk-type", "", "Required. Data disk type")
	req.ClickhouseVersion = flags.String("clickhouse-version", "", "Required. ClickHouse version")
	adminPassword = flags.String("admin-password", "", "Required. Admin password; CLI base64-encodes it before sending")
	req.VPCId = flags.String("vpc-id", "", "Optional. VPC ID")
	req.SubnetId = flags.String("subnet-id", "", "Optional. Subnet ID")
	req.ShardCount = flags.Int("shard-count", 1, "Optional. Shard count, default 1")
	req.ReplicateCount = flags.Int("replicate-count", 2, "Optional. Replicate count, 1 or 2, default 2")
	req.DataDiskSize = flags.Int("data-disk-size-gb", 100, "Optional. Data disk size in GB, default 100")
	req.ChargeType = flags.String("charge-type", "Month", "Optional. 'Year', 'Month', or 'Dynamic', default Month")
	req.Quantity = flags.Int("quantity", 1, "Optional. Purchase duration, default 1")
	req.BackupId = flags.String("backup-id", "", "Optional. Backup task ID to restore from")
	req.IsZookeeperHA = flags.Bool("zookeeper-ha", true, "Optional. Enable Zookeeper HA, default true")
	req.ZookeeperMachineTypeId = flags.String("zookeeper-machine-type-id", "", "Required when --zookeeper-ha=true. Zookeeper machine type ID")
	req.ZookeeperDataDiskType = flags.String("zookeeper-data-disk-type", "", "Required when --zookeeper-ha=true. Zookeeper data disk type")
	req.ZookeeperDataDiskSize = flags.String("zookeeper-data-disk-size-gb", "", "Required when --zookeeper-ha=true. Zookeeper data disk size in GB")
	req.IsSecGroup = flags.String("sec-group", "false", "Optional. Enable security group, true or false")
	req.SecGroupIds = flags.String("sec-group-ids", "", "Optional. Security group IDs")
	req.IsMultiZone = flags.String("multi-zone", "false", "Optional. Enable multi-zone, true or false")
	flags.StringSliceVar(&req.MultiZones, "multi-zone-name", nil, "Optional. Availability zone name for multi-zone clusters")
	flags.StringSliceVar(&labels, "label", nil, "Optional. Resource label, format: key=value, repeatable")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	async = flags.Bool("async", false, "Optional. Do not wait for creation to finish")

	command.SetFlagValues(cmd, "charge-type", "Month", "Year", "Dynamic")
	command.SetFlagValues(cmd, "sec-group", "false", "true")
	command.SetFlagValues(cmd, "multi-zone", "false", "true")

	cmd.MarkFlagRequired("clickhouse-machine-type-id")
	cmd.MarkFlagRequired("data-disk-type")
	cmd.MarkFlagRequired("clickhouse-version")
	cmd.MarkFlagRequired("admin-password")

	return cmd
}

func createUClickhouseCluster(client *uclickhousesdk.UClickhouseClient, req *uclickhousesdk.CreateUClickhouseClusterRequest) (*createUClickhouseClusterResponse, error) {
	var resp createUClickhouseClusterResponse
	reqCopier := *req
	err := invokeUClickhouseAction(client, "CreateUClickhouseCluster", &reqCopier, &resp)
	return &resp, err
}

func parseLabels(labels []string) ([]uclickhousesdk.CreateUClickhouseClusterParamLabels, error) {
	parsed := []uclickhousesdk.CreateUClickhouseClusterParamLabels{}
	for _, label := range labels {
		key, value, ok := strings.Cut(label, "=")
		if !ok || key == "" {
			return nil, fmt.Errorf("invalid label %q, want key=value", label)
		}
		parsed = append(parsed, uclickhousesdk.CreateUClickhouseClusterParamLabels{
			Key:   sdk.String(key),
			Value: sdk.String(value),
		})
	}
	return parsed, nil
}

func validateCreateArgs(cmd *cobra.Command, args []string) error {
	if err := noFlagLikeValues(
		cmd,
		"clickhouse-machine-type-id",
		"data-disk-type",
		"clickhouse-version",
		"zookeeper-machine-type-id",
		"zookeeper-data-disk-type",
		"zookeeper-data-disk-size-gb",
		"sec-group-ids",
	); err != nil {
		return err
	}
	if err := noArgs(cmd, args); err != nil {
		return err
	}
	if err := requireFlagsWhenBool(cmd, "zookeeper-ha", true,
		"zookeeper-machine-type-id",
		"zookeeper-data-disk-type",
		"zookeeper-data-disk-size-gb",
	); err != nil {
		return err
	}
	if err := requireFlagsWhenString(cmd, "sec-group", "true", "sec-group-ids"); err != nil {
		return err
	}
	if err := requireFlagsWhenString(cmd, "multi-zone", "true", "multi-zone-name"); err != nil {
		return err
	}
	return nil
}
