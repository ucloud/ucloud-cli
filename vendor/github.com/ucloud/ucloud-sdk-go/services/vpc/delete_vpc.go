//Code is generated by ucloud code generator, don't modify it by hand, it will cause undefined behaviors.
//go:generate ucloud-gen-go-api VPC DeleteVPC

package vpc

import (
	"github.com/ucloud/ucloud-sdk-go/sdk"
	"github.com/ucloud/ucloud-sdk-go/sdk/request"
	"github.com/ucloud/ucloud-sdk-go/sdk/response"
)

// DeleteVPCRequest is request schema for DeleteVPC action
type DeleteVPCRequest struct {
	request.CommonBase

	// VPC资源Id
	VPCId *string `required:"true"`
}

// DeleteVPCResponse is response schema for DeleteVPC action
type DeleteVPCResponse struct {
	response.CommonBase
}

// NewDeleteVPCRequest will create request of DeleteVPC action.
func (c *VPCClient) NewDeleteVPCRequest() *DeleteVPCRequest {
	cfg := c.client.GetConfig()

	return &DeleteVPCRequest{
		CommonBase: request.CommonBase{
			Region:    sdk.String(cfg.Region),
			ProjectId: sdk.String(cfg.ProjectId),
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