package group

import (
	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-sdk-go/services/urocketmq"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"
)

// GroupList returns group names for the given service. Exported for self-use (delete --group-name completion)
// and reused by sibling list; cross-group one-way imports service, same-group calls directly.
func GroupList(ctx *cli.Context, projectID, region, serviceID string) []string {
	client := cli.NewServiceClient(ctx, urocketmq.NewClient)
	req := client.NewListURocketMQGroupRequest()
	req.ProjectId = sdk.String(projectID)
	req.Region = sdk.String(region)
	req.ServiceId = sdk.String(serviceID)
	names := make([]string, 0)
	for limit, offset := 50, 0; ; offset += limit {
		req.Limit = sdk.Int(limit)
		req.Offset = sdk.Int(offset)
		resp, err := client.ListURocketMQGroup(req)
		if err != nil {
			return nil
		}
		for _, g := range resp.GroupList {
			names = append(names, g.GroupName)
		}
		if offset+limit >= resp.TotalCount {
			break
		}
	}
	return names
}
