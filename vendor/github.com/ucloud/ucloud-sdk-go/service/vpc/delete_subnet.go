//go:generate go run ../../private/cli/gen-api/main.go vpc DeleteSubnet

package vpc

import (
	"github.com/ucloud/ucloud-sdk-go/sdk/request"
	"github.com/ucloud/ucloud-sdk-go/sdk/response"
)

type DeleteSubnetRequest struct {
	request.CommonBase

	// Required, 子网ID
	SubnetId string
}

type DeleteSubnetResponse struct {
	response.CommonBase
}

// NewDeleteSubnetRequest will create request of DeleteSubnet action.
func (c *VPCClient) NewDeleteSubnetRequest() *DeleteSubnetRequest {
	cfg := c.client.GetConfig()

	return &DeleteSubnetRequest{
		CommonBase: request.CommonBase{
			Region:    cfg.Region,
			ProjectId: cfg.ProjectId,
		},
	}
}

// DeleteSubnet - 删除子网
func (c *VPCClient) DeleteSubnet(req *DeleteSubnetRequest) (*DeleteSubnetResponse, error) {
	var err error
	var res DeleteSubnetResponse

	err = c.client.InvokeAction("DeleteSubnet", req, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
