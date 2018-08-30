//go:generate go run ../../private/cli/gen-api/main.go vpc CreateVPCIntercom

package vpc

import (
	"github.com/ucloud/ucloud-sdk-go/sdk/request"
	"github.com/ucloud/ucloud-sdk-go/sdk/response"
)

type CreateVPCIntercomRequest struct {
	request.CommonBase

	// Required, 源VPC短ID
	VPCId string

	// Required, 目的VPC短ID
	DstVPCId string

	// Optional, 目的所在地域（如果目的VPC和源VPC不在同一个地域，两个地域需要建立跨域通道，且该字段必选）
	DstRegion string

	// Optional, 目的项目ID
	DstProjectId string
}

type CreateVPCIntercomResponse struct {
	response.CommonBase
}

// NewCreateVPCIntercomRequest will create request of CreateVPCIntercom action.
func (c *VPCClient) NewCreateVPCIntercomRequest() *CreateVPCIntercomRequest {
	cfg := c.client.GetConfig()

	return &CreateVPCIntercomRequest{
		CommonBase: request.CommonBase{
			Region:    cfg.Region,
			ProjectId: cfg.ProjectId,
		},
	}
}

// CreateVPCIntercom - 新建VPC互通关系
func (c *VPCClient) CreateVPCIntercom(req *CreateVPCIntercomRequest) (*CreateVPCIntercomResponse, error) {
	var err error
	var res CreateVPCIntercomResponse

	err = c.client.InvokeAction("CreateVPCIntercom", req, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
