//go:generate go run ../../private/cli/gen-api/main.go ulb DescribeULB

package ulb

import (
	"github.com/ucloud/ucloud-sdk-go/sdk/request"
	"github.com/ucloud/ucloud-sdk-go/sdk/response"
	. "github.com/ucloud/ucloud-sdk-go/service/ulb/types"
)

type DescribeULBRequest struct {
	request.CommonBase

	// Optional, 数据偏移量，默认为0
	Offset int

	// Optional, 数据分页值，默认为20
	Limit int

	// Optional, 负载均衡实例的Id。 若指定则返回指定的负载均衡实例的信息； 若不指定则返回当前数据中心中所有的负载均衡实例的信息
	ULBId string

	// Optional, ULB所属的VPC
	VPCId string

	// Optional, ULB所属的子网ID
	SubnetId string

	// Optional, ULB所属的业务组ID
	BusinessId string
}

type DescribeULBResponse struct {
	response.CommonBase

	// 满足条件的ULB总数
	TotalCount int

	// ULB列表，每项参数详见 ULBSet
	DataSet []ULBSet
}

// NewDescribeULBRequest will create request of DescribeULB action.
func (c *ULBClient) NewDescribeULBRequest() *DescribeULBRequest {
	cfg := c.client.GetConfig()

	return &DescribeULBRequest{
		CommonBase: request.CommonBase{
			Region:    cfg.Region,
			ProjectId: cfg.ProjectId,
		},
	}
}

// DescribeULB - 获取ULB详细信息
func (c *ULBClient) DescribeULB(req *DescribeULBRequest) (*DescribeULBResponse, error) {
	var err error
	var res DescribeULBResponse

	err = c.client.InvokeAction("DescribeULB", req, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
