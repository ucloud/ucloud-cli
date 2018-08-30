//go:generate go run ../../private/cli/gen-api/main.go vpc DescribeVPC

package vpc

import (
	"github.com/ucloud/ucloud-sdk-go/sdk/request"
	"github.com/ucloud/ucloud-sdk-go/sdk/response"
	. "github.com/ucloud/ucloud-sdk-go/service/vpc/types"
)

type DescribeVPCRequest struct {
	request.CommonBase

	// Optional, VPCId
	VPCIds []string

	// Optional, 业务组名称
	Tag string
}

type DescribeVPCResponse struct {
	response.CommonBase

	//
	DataSet []VPCInfoSet
}

// NewDescribeVPCRequest will create request of DescribeVPC action.
func (c *VPCClient) NewDescribeVPCRequest() *DescribeVPCRequest {
	cfg := c.client.GetConfig()

	return &DescribeVPCRequest{
		CommonBase: request.CommonBase{
			Region:    cfg.Region,
			ProjectId: cfg.ProjectId,
		},
	}
}

// DescribeVPC - 获取VPC信息
func (c *VPCClient) DescribeVPC(req *DescribeVPCRequest) (*DescribeVPCResponse, error) {
	var err error
	var res DescribeVPCResponse

	err = c.client.InvokeAction("DescribeVPC", req, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
