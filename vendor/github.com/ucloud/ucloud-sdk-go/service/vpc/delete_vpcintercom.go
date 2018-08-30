//go:generate go run ../../private/cli/gen-api/main.go vpc DeleteVPCIntercom

package vpc

import (
	"github.com/ucloud/ucloud-sdk-go/sdk/request"
	"github.com/ucloud/ucloud-sdk-go/sdk/response"
)

type DeleteVPCIntercomRequest struct {
	request.CommonBase

	// Required, 源VPC短ID
	VPCId string

	// Required, 目的VPC短ID
	DstVPCId string

	// Optional, 目的所在地域
	DstRegion string

	// Optional, 目的项目ID（如果目的VPC和源VPC不在同一个地域，两个地域需要建立跨域通道，且该字段必选）
	DstProjectId string
}

type DeleteVPCIntercomResponse struct {
	response.CommonBase
}

// NewDeleteVPCIntercomRequest will create request of DeleteVPCIntercom action.
func (c *VPCClient) NewDeleteVPCIntercomRequest() *DeleteVPCIntercomRequest {
	cfg := c.client.GetConfig()

	return &DeleteVPCIntercomRequest{
		CommonBase: request.CommonBase{
			Region:    cfg.Region,
			ProjectId: cfg.ProjectId,
		},
	}
}

// DeleteVPCIntercom - 删除VPC互通关系
func (c *VPCClient) DeleteVPCIntercom(req *DeleteVPCIntercomRequest) (*DeleteVPCIntercomResponse, error) {
	var err error
	var res DeleteVPCIntercomResponse

	err = c.client.InvokeAction("DeleteVPCIntercom", req, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
