//go:generate go run ../../private/cli/gen-api/main.go ulb DeleteVServer

package ulb

import (
	"github.com/ucloud/ucloud-sdk-go/sdk/request"
	"github.com/ucloud/ucloud-sdk-go/sdk/response"
)

type DeleteVServerRequest struct {
	request.CommonBase

	// Required, 负载均衡实例的ID
	ULBId string

	// Required, VServer实例的ID
	VServerId string
}

type DeleteVServerResponse struct {
	response.CommonBase
}

// NewDeleteVServerRequest will create request of DeleteVServer action.
func (c *ULBClient) NewDeleteVServerRequest() *DeleteVServerRequest {
	cfg := c.client.GetConfig()

	return &DeleteVServerRequest{
		CommonBase: request.CommonBase{
			Region:    cfg.Region,
			ProjectId: cfg.ProjectId,
		},
	}
}

// DeleteVServer - 删除VServer实例
func (c *ULBClient) DeleteVServer(req *DeleteVServerRequest) (*DeleteVServerResponse, error) {
	var err error
	var res DeleteVServerResponse

	err = c.client.InvokeAction("DeleteVServer", req, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
