//go:generate go run ../../private/cli/gen-api/main.go ulb DeleteULB

package ulb

import (
	"github.com/ucloud/ucloud-sdk-go/sdk/request"
	"github.com/ucloud/ucloud-sdk-go/sdk/response"
)

type DeleteULBRequest struct {
	request.CommonBase

	// Required, 负载均衡实例的ID
	ULBId string
}

type DeleteULBResponse struct {
	response.CommonBase
}

// NewDeleteULBRequest will create request of DeleteULB action.
func (c *ULBClient) NewDeleteULBRequest() *DeleteULBRequest {
	cfg := c.client.GetConfig()

	return &DeleteULBRequest{
		CommonBase: request.CommonBase{
			Region:    cfg.Region,
			ProjectId: cfg.ProjectId,
		},
	}
}

// DeleteULB - 删除负载均衡实例
func (c *ULBClient) DeleteULB(req *DeleteULBRequest) (*DeleteULBResponse, error) {
	var err error
	var res DeleteULBResponse

	err = c.client.InvokeAction("DeleteULB", req, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
