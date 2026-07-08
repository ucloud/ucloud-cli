package memcache

import (
	"fmt"

	"github.com/ucloud/ucloud-sdk-go/services/umem"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

func describeByID(ctx *cli.Context) func(string, *request.CommonBase) (interface{}, error) {
	return func(memcacheID string, commonBase *request.CommonBase) (interface{}, error) {
		client := cli.NewServiceClient(ctx, umem.NewClient)
		req := client.NewDescribeUMemRequest()
		if commonBase != nil {
			req.CommonBase = *commonBase
		}
		req.Protocol = sdk.String("memcache")
		req.ResourceId = &memcacheID

		resp, err := client.DescribeUMem(req)
		if err != nil {
			return nil, err
		}
		if len(resp.DataSet) < 1 {
			return nil, fmt.Errorf("resource [%s] may not exist", memcacheID)
		}
		return &resp.DataSet[0], nil
	}
}
