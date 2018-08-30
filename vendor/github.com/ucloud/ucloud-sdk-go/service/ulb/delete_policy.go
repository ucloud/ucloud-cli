//go:generate go run ../../private/cli/gen-api/main.go ulb DeletePolicy

package ulb

import (
	"github.com/ucloud/ucloud-sdk-go/sdk/request"
	"github.com/ucloud/ucloud-sdk-go/sdk/response"
)

type DeletePolicyRequest struct {
	request.CommonBase

	// Required, 内容转发策略ID
	PolicyId string

	// Optional, 内容转发策略组ID
	GroupId string

	// Optional, VServer 资源ID
	VServerId string
}

type DeletePolicyResponse struct {
	response.CommonBase
}

// NewDeletePolicyRequest will create request of DeletePolicy action.
func (c *ULBClient) NewDeletePolicyRequest() *DeletePolicyRequest {
	cfg := c.client.GetConfig()

	return &DeletePolicyRequest{
		CommonBase: request.CommonBase{
			Region:    cfg.Region,
			ProjectId: cfg.ProjectId,
		},
	}
}

// DeletePolicy - 删除内容转发策略
func (c *ULBClient) DeletePolicy(req *DeletePolicyRequest) (*DeletePolicyResponse, error) {
	var err error
	var res DeletePolicyResponse

	err = c.client.InvokeAction("DeletePolicy", req, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
