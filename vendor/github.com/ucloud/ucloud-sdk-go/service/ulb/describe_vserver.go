//go:generate go run ../../private/cli/gen-api/main.go ulb DescribeVServer

package ulb

import (
	"github.com/ucloud/ucloud-sdk-go/sdk/request"
	"github.com/ucloud/ucloud-sdk-go/sdk/response"
	. "github.com/ucloud/ucloud-sdk-go/service/ulb/types"
)

type DescribeVServerRequest struct {
	request.CommonBase

	// Required, 负载均衡实例的Id
	ULBId string

	// Optional, VServer实例的Id；若指定则返回指定的VServer实例的信息； 若不指定则返回当前负载均衡实例下所有VServer的信息
	VServerId string
}

type DescribeVServerResponse struct {
	response.CommonBase

	// 满足条件的VServer总数
	TotalCount int

	// VServer列表，每项参数详见 ULBVServerSet
	DataSet []ULBVServerSet
}

// NewDescribeVServerRequest will create request of DescribeVServer action.
func (c *ULBClient) NewDescribeVServerRequest() *DescribeVServerRequest {
	cfg := c.client.GetConfig()

	return &DescribeVServerRequest{
		CommonBase: request.CommonBase{
			Region:    cfg.Region,
			ProjectId: cfg.ProjectId,
		},
	}
}

// DescribeVServer - 获取ULB下的VServer的详细信息
func (c *ULBClient) DescribeVServer(req *DescribeVServerRequest) (*DescribeVServerResponse, error) {
	var err error
	var res DescribeVServerResponse

	err = c.client.InvokeAction("DescribeVServer", req, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
