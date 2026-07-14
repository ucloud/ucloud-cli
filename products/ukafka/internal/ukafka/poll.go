package ukafka

import (
	"fmt"

	ukafkasdk "github.com/ucloud/ucloud-sdk-go/services/ukafka"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// describeUKafkaInstanceByID returns the poller's describe func
// It uses commonBase to get region and zone for the request
func describeUKafkaInstanceByID(ctx *cli.Context) func(instanceID string, commonBase *request.CommonBase) (interface{}, error) {
	return func(instanceID string, commonBase *request.CommonBase) (interface{}, error) {
		client := cli.NewServiceClient(ctx, ukafkasdk.NewClient)
		genReq := client.Client.NewGenericRequest()
		genReq.SetAction("DescribeUKafkaInstance")

		// Use region/zone from commonBase if available, otherwise use defaults
		if commonBase != nil {
			if commonBase.Region != nil && *commonBase.Region != "" {
				genReq.SetRegion(*commonBase.Region)
			} else {
				genReq.SetRegion(ctx.DefaultRegion())
			}
			if commonBase.Zone != nil && *commonBase.Zone != "" {
				genReq.SetZone(*commonBase.Zone)
			} else {
				genReq.SetZone(ctx.DefaultZone())
			}
			if commonBase.ProjectId != nil && *commonBase.ProjectId != "" {
				genReq.SetProjectId(*commonBase.ProjectId)
			}
		} else {
			genReq.SetRegion(ctx.DefaultRegion())
			genReq.SetZone(ctx.DefaultZone())
		}

		payload := map[string]interface{}{
			"ClusterInstanceId": instanceID,
		}
		genReq.SetPayload(payload)

		genResp, err := client.Client.GenericInvoke(genReq)
		if err != nil {
			return nil, err
		}

		var resp DescribeUKafkaInstanceResponse
		if err := genResp.Unmarshal(&resp); err != nil {
			return nil, err
		}
		if len(resp.ClusterSet) == 0 {
			return nil, fmt.Errorf("instance not found")
		}
		return &resp.ClusterSet[0], nil
	}
}
