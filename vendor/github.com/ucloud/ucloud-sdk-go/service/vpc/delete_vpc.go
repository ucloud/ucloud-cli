//go:generate go run ../../private/cli/gen-api/main.go vpc DeleteVPC

package vpc

import (
	"github.com/ucloud/ucloud-sdk-go/sdk/request"
	"github.com/ucloud/ucloud-sdk-go/sdk/response"
)

type DeleteVPCRequest struct {
	request.CommonBase

	// Required, VPC资源Id
	VPCId string
}

type DeleteVPCResponse struct {
	response.CommonBase
}

// NewDeleteVPCRequest will create request of DeleteVPC action.
func (c *VPCClient) NewDeleteVPCRequest() *DeleteVPCRequest {
	cfg := c.client.GetConfig()

	return &DeleteVPCRequest{
		CommonBase: request.CommonBase{
			Region:    cfg.Region,
			ProjectId: cfg.ProjectId,
		},
	}
}

// DeleteVPC - 删除VPC
func (c *VPCClient) DeleteVPC(req *DeleteVPCRequest) (*DeleteVPCResponse, error) {
	var err error
	var res DeleteVPCResponse

	err = c.client.InvokeAction("DeleteVPC", req, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
