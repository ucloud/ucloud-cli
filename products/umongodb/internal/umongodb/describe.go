package umongodb

import (
	"strconv"
	"time"

	"github.com/spf13/cobra"

	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-sdk-go/services/umongodb"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newDescribe implements `umongodb describe`.
func newDescribe(ctx *cli.Context) *cobra.Command {
	var clusterID string

	client := cli.NewServiceClient(ctx, umongodb.NewClient)
	req := client.NewDescribeUMongoDBInstanceRequest()

	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Show details of one MongoDB instance",
		Long:  "Show the full attribute/value detail of a single MongoDB instance.",
		Run: func(c *cobra.Command, args []string) {
			req.ClusterId = sdk.String(ctx.PickResourceID(clusterID))

			resp, err := client.DescribeUMongoDBInstance(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}

			ci := resp.ClusterInfo
			rows := []cli.DescribeRow{
				{Attribute: "ClusterId", Content: ci.ClusterId},
				{Attribute: "Name", Content: ci.InstanceName},
				{Attribute: "ClusterType", Content: ci.ClusterType},
				{Attribute: "State", Content: ci.State},
				{Attribute: "DBVersion", Content: ci.DBVersion},
				{Attribute: "ConnectURL", Content: ci.ConnectURL},
				{Attribute: "VPCId", Content: ci.VPCId},
				{Attribute: "SubnetId", Content: ci.SubnetId},
				{Attribute: "Tag", Content: ci.Tag},
				{Attribute: "DiskSpace(GB)", Content: strconv.Itoa(ci.DiskSpace)},
				{Attribute: "MachineType", Content: ci.MachineTypeId},
			}

			// Shard info (for sharded clusters)
			if ci.ShardCount > 0 {
				rows = append(rows, cli.DescribeRow{Attribute: "ShardCount", Content: strconv.Itoa(ci.ShardCount)})
			}
			if ci.ShardNodeCount > 0 {
				rows = append(rows, cli.DescribeRow{Attribute: "ShardNodeCount", Content: strconv.Itoa(ci.ShardNodeCount)})
			}
			if ci.MongosCount > 0 {
				rows = append(rows, cli.DescribeRow{Attribute: "MongosCount", Content: strconv.Itoa(ci.MongosCount)})
			}

			// CreateTime
			if ci.CreateTime > 0 {
				rows = append(rows, cli.DescribeRow{Attribute: "CreateTime", Content: time.Unix(int64(ci.CreateTime), 0).Format("2006-01-02 15:04:05")})
			}

			ctx.PrintList(rows)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringVar(&clusterID, resourceIDFlag, "", "Required. Cluster ID of the MongoDB instance to describe.")

	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)

	cmd.MarkFlagRequired(resourceIDFlag)
	command.SetCompletion(cmd, resourceIDFlag, func() []string {
		return getMongoDBIDList(ctx, nil, *req.Region, *req.Zone, *req.ProjectId)
	})

	return cmd
}
