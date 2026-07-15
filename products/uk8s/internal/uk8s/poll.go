package uk8s

import (
	"fmt"

	uk8ssdk "github.com/ucloud/ucloud-sdk-go/services/uk8s"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// describeByID is the Poller's describe closure. The signature
// func(string, *request.CommonBase) (interface{}, error) is exactly what
// ctx.PollerTo expects; the returned interface{} is asserted back to
// *uk8ssdk.DescribeUK8SClusterResponse in the polling logic so the Poller can
// read resp.Status against the target states declared by the create verb.
func describeByID(ctx *cli.Context) func(string, *request.CommonBase) (interface{}, error) {
	client := cli.NewServiceClient(ctx, uk8ssdk.NewClient)
	return func(clusterID string, common *request.CommonBase) (interface{}, error) {
		req := client.NewDescribeUK8SClusterRequest()
		if common != nil {
			req.CommonBase = *common
		}
		req.ClusterId = sdk.String(clusterID)
		resp, err := client.DescribeUK8SCluster(req)
		if err != nil {
			return nil, err
		}
		if resp.ClusterId == "" {
			return nil, fmt.Errorf("cluster %q not found", clusterID)
		}
		return resp, nil
	}
}
