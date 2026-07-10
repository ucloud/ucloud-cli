package topic

import (
	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-sdk-go/services/urocketmq"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"
)

// TopicList returns the topic name list under the specified service, for --topic-name completion reuse
// by delete/update (exported for use by other packages).
func TopicList(ctx *cli.Context, projectID, region, serviceID string) []string {
	client := cli.NewServiceClient(ctx, urocketmq.NewClient)
	req := client.NewListURocketMQTopicRequest()
	req.ProjectId = sdk.String(projectID)
	req.Region = sdk.String(region)
	req.ServiceId = sdk.String(serviceID)
	req.Limit = sdk.Int(1000)
	req.Offset = sdk.Int(0)
	resp, err := client.ListURocketMQTopic(req)
	if err != nil {
		return nil
	}
	names := make([]string, 0, len(resp.TopicList))
	for _, t := range resp.TopicList {
		names = append(names, t.TopicName)
	}
	return names
}
