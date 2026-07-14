package uhadoop

import (
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	uhadoopsdk "github.com/ucloud/ucloud-sdk-go/services/uhadoop"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

func describeClusterForPoll(ctx *cli.Context, client *uhadoopsdk.UHadoopClient) func(string, *request.CommonBase) (interface{}, error) {
	return func(id string, _ *request.CommonBase) (interface{}, error) {
		req := client.NewDescribeUHadoopInstanceRequest()
		req.InstanceId = sdk.String(id)
		var resp describeClusterResponse
		err := client.InvokeAction("DescribeUHadoopInstance", req, &resp)
		if err != nil {
			return nil, err
		}
		if len(resp.ClusterSet) == 0 {
			return nil, nil
		}
		return resp.ClusterSet[0], nil
	}
}
