package clickhouse

import (
	"strings"

	uclickhousesdk "github.com/ucloud/ucloud-sdk-go/services/uclickhouse"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// getClusterList returns "ClusterId/Name" completion candidates for clickhouse-id flags.
func getClusterList(ctx *cli.Context, statuses []string, project, region string) []string {
	client := cli.NewServiceClient(ctx, uclickhousesdk.NewClient)
	req := client.NewListUClickhouseClusterRequest()
	req.ProjectId = sdk.String(project)
	req.Region = sdk.String(region)
	resp, err := listUClickhouseCluster(client, req)
	if err != nil {
		return nil
	}
	list := []string{}
	for _, cluster := range resp.Data.Clusters {
		if cluster.ClusterId == "" {
			continue
		}
		if statuses != nil {
			matched := false
			for _, status := range statuses {
				if cluster.Status == status {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}
		list = append(list, cluster.ClusterId+"/"+strings.Replace(cluster.ClusterName, " ", "-", -1))
	}
	return list
}
