//go:generate go run ../../private/cli/gen-api/main.go vpc DescribeSubnet

package vpc

import (
	"github.com/ucloud/ucloud-sdk-go/sdk/request"
	"github.com/ucloud/ucloud-sdk-go/sdk/response"
	. "github.com/ucloud/ucloud-sdk-go/service/vpc/types"
)

type DescribeSubnetRequest struct {
	request.CommonBase

	// Optional, 子网id数组，适用于一次查询多个子网信息
	SubnetIds []string

	// Optional, 子网id，适用于一次查询一个子网信息
	SubnetId string

	// Optional, VPC资源id
	VPCId string

	// Optional, 业务组名称，默认为Default
	Tag string

	// Optional, 业务组
	BusinessId string

	// Optional, 默认为0
	Offset int

	// Optional, 默认为20
	Limit int
}

type DescribeSubnetResponse struct {
	response.CommonBase

	// 子网总数量
	TotalCount int

	// 子网信息数组
	DataSet []VPCSubnetInfoSet
}

// NewDescribeSubnetRequest will create request of DescribeSubnet action.
func (c *VPCClient) NewDescribeSubnetRequest() *DescribeSubnetRequest {
	cfg := c.client.GetConfig()

	return &DescribeSubnetRequest{
		CommonBase: request.CommonBase{
			Region:    cfg.Region,
			ProjectId: cfg.ProjectId,
		},
	}
}

// DescribeSubnet - 获取子网信息
func (c *VPCClient) DescribeSubnet(req *DescribeSubnetRequest) (*DescribeSubnetResponse, error) {
	var err error
	var res DescribeSubnetResponse

	err = c.client.InvokeAction("DescribeSubnet", req, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
