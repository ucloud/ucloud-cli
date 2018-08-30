//go:generate go run ../../private/cli/gen-api/main.go vpc DescribeVPCIntercom

package vpc

import (
	"github.com/ucloud/ucloud-sdk-go/sdk/request"
	"github.com/ucloud/ucloud-sdk-go/sdk/response"
)

type DescribeVPCIntercomRequest struct {
	request.CommonBase

	// Required, VPC短ID
	VPCId string

	// Optional, 目的地域
	DstRegion string

	// Optional, 目的项目ID
	DstProjectId string
}

type DescribeVPCIntercomResponse struct {
	response.CommonBase

	// VPC的地址空间
	Network []string

	// 所属地域
	DstRegion string

	// VPC名字
	Name string

	// 项目Id
	ProjectId string

	// vpc_id
	VPCId string

	// 业务组（未分组显示为 Default）
	Tag string
}

// NewDescribeVPCIntercomRequest will create request of DescribeVPCIntercom action.
func (c *VPCClient) NewDescribeVPCIntercomRequest() *DescribeVPCIntercomRequest {
	cfg := c.client.GetConfig()

	return &DescribeVPCIntercomRequest{
		CommonBase: request.CommonBase{
			Region:    cfg.Region,
			ProjectId: cfg.ProjectId,
		},
	}
}

// DescribeVPCIntercom - 获取VPC互通信息
func (c *VPCClient) DescribeVPCIntercom(req *DescribeVPCIntercomRequest) (*DescribeVPCIntercomResponse, error) {
	var err error
	var res DescribeVPCIntercomResponse

	err = c.client.InvokeAction("DescribeVPCIntercom", req, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
