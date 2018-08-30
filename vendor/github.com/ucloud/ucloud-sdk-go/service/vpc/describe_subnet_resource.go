//go:generate go run ../../private/cli/gen-api/main.go vpc DescribeSubnetResource

package vpc

import (
	"github.com/ucloud/ucloud-sdk-go/sdk/request"
	"github.com/ucloud/ucloud-sdk-go/sdk/response"
	. "github.com/ucloud/ucloud-sdk-go/service/vpc/types"
)

type DescribeSubnetResourceRequest struct {
	request.CommonBase

	// Required, 子网id
	SubnetId string

	// Optional, 资源类型
	ResourceType string

	// Optional, 分页号
	Offset int

	// Optional, 单页limit
	Limit int
}

type DescribeSubnetResourceResponse struct {
	response.CommonBase

	// 总数
	TotalCount int

	// 返回数据集
	DataSet []ResourceInfo
}

// NewDescribeSubnetResourceRequest will create request of DescribeSubnetResource action.
func (c *VPCClient) NewDescribeSubnetResourceRequest() *DescribeSubnetResourceRequest {
	cfg := c.client.GetConfig()

	return &DescribeSubnetResourceRequest{
		CommonBase: request.CommonBase{
			Region:    cfg.Region,
			ProjectId: cfg.ProjectId,
		},
	}
}

// DescribeSubnetResource - 展示子网资源
func (c *VPCClient) DescribeSubnetResource(req *DescribeSubnetResourceRequest) (*DescribeSubnetResourceResponse, error) {
	var err error
	var res DescribeSubnetResourceResponse

	err = c.client.InvokeAction("DescribeSubnetResource", req, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
