//go:generate go run ../../private/cli/gen-api/main.go unet ReleaseEIP

package unet

import (
	"github.com/ucloud/ucloud-sdk-go/sdk/request"
	"github.com/ucloud/ucloud-sdk-go/sdk/response"
)

type ReleaseEIPRequest struct {
	request.CommonBase

	// Required, 弹性IP的资源ID
	EIPId string
}

type ReleaseEIPResponse struct {
	response.CommonBase
}

// NewReleaseEIPRequest will create request of ReleaseEIP action.
func (c *UNetClient) NewReleaseEIPRequest() *ReleaseEIPRequest {
	cfg := c.client.GetConfig()

	return &ReleaseEIPRequest{
		CommonBase: request.CommonBase{
			Region:    cfg.Region,
			ProjectId: cfg.ProjectId,
		},
	}
}

// ReleaseEIP - 释放弹性IP资源, 所释放弹性IP必须为非绑定状态.
func (c *UNetClient) ReleaseEIP(req *ReleaseEIPRequest) (*ReleaseEIPResponse, error) {
	var err error
	var res ReleaseEIPResponse

	err = c.client.InvokeAction("ReleaseEIP", req, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
