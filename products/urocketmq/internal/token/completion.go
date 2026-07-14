package token

import (
	"github.com/ucloud/ucloud-cli/pkg/cli"
	urocketmq "github.com/ucloud/ucloud-sdk-go/services/urocketmq"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"
)

// TokenList returns the TokenId list under the specified Service, for same-group --token-id completion reuse.
// ListURocketMQToken Limit max is 100 (different from service/topic's 1000),
// paginates by 100 to fetch all, see group.GroupList.
func TokenList(ctx *cli.Context, projectID, region, serviceID string) []string {
	client := cli.NewServiceClient(ctx, urocketmq.NewClient)
	req := client.NewListURocketMQTokenRequest()
	req.ProjectId = sdk.String(projectID)
	req.Region = sdk.String(region)
	req.ServiceId = sdk.String(serviceID)
	ids := make([]string, 0)
	for limit, offset := 100, 0; ; offset += limit {
		req.Limit = sdk.Int(limit)
		req.Offset = sdk.Int(offset)
		resp, err := client.ListURocketMQToken(req)
		if err != nil {
			return nil
		}
		for _, t := range resp.TokenList {
			ids = append(ids, t.TokenId)
		}
		if offset+limit >= resp.TotalCount {
			break
		}
	}
	return ids
}
