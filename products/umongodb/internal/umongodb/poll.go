package umongodb

import (
	"github.com/ucloud/ucloud-sdk-go/services/umongodb"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// describeByID returns the Poller describe func: given a ClusterId it fetches
// the current cluster state via typed DescribeUMongoDBInstance.
// region/zone are captured at call time because the Poller always passes nil
// for CommonBase (see image/internal/image/describe.go for the same pattern).
func describeByID(ctx *cli.Context, region, zone string) func(string, *request.CommonBase) (interface{}, error) {
	return func(id string, _ *request.CommonBase) (interface{}, error) {
		client := cli.NewServiceClient(ctx, umongodb.NewClient)
		req := client.NewDescribeUMongoDBInstanceRequest()
		req.Region = &region
		if zone != "" {
			req.Zone = &zone
		}
		req.ClusterId = &id
		resp, err := client.DescribeUMongoDBInstance(req)
		if err != nil {
			return nil, err
		}
		return &resp.ClusterInfo, nil
	}
}
