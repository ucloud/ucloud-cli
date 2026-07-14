package clickhouse

import (
	"fmt"

	"github.com/spf13/cobra"

	uclickhousesdk "github.com/ucloud/ucloud-sdk-go/services/uclickhouse"
	"github.com/ucloud/ucloud-sdk-go/ucloud/response"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

type listUClickhouseClusterResponse struct {
	response.CommonBase
	Data    listUClickhouseClusterResponseData
	Message string
}

type listUClickhouseClusterResponseData struct {
	Clusters   []uclickhousesdk.ClickhouseCluster
	TotalCount int
}

// newList ucloud clickhouse list
func newList(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, uclickhousesdk.NewClient)
	req := client.NewListUClickhouseClusterRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List UClickhouse clusters",
		Long:  "List UClickhouse clusters",
		Args:  noArgs,
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := listUClickhouseCluster(client, req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			list := []ClusterRow{}
			for _, cluster := range resp.Data.Clusters {
				list = append(list, clusterRow(cluster))
			}
			ctx.PrintList(list)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	return cmd
}

func listUClickhouseCluster(client *uclickhousesdk.UClickhouseClient, req *uclickhousesdk.ListUClickhouseClusterRequest) (*listUClickhouseClusterResponse, error) {
	var resp listUClickhouseClusterResponse
	reqCopier := *req
	err := invokeUClickhouseAction(client, "ListUClickhouseCluster", &reqCopier, &resp)
	return &resp, err
}

func clusterRow(cluster uclickhousesdk.ClickhouseCluster) ClusterRow {
	return ClusterRow{
		ClusterID:               cluster.ClusterId,
		ClusterName:             cluster.ClusterName,
		Status:                  cluster.Status,
		ClickhouseVersion:       cluster.ClickhouseVersion,
		ShardCount:              fmt.Sprintf("%d", cluster.ShardCount),
		ReplicateCount:          fmt.Sprintf("%d", cluster.ReplicateCount),
		VPCId:                   cluster.VPCId,
		SubnetId:                cluster.SubnetId,
		ClickhouseMachineTypeID: cluster.ClickhouseMachineTypeId,
		ClickhouseDataDiskType:  cluster.ClickhouseDataDiskType,
		ClickhouseDataDiskSize:  fmt.Sprintf("%d", cluster.ClickhouseDataDiskSize),
		CreateTime:              formatUnixDate(cluster.CreateTimestamp),
		ExpireTime:              formatUnixDate(int(cluster.ExpireTimestamp)),
	}
}
