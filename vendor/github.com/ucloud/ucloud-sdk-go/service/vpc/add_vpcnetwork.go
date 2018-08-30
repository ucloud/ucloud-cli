//go:generate go run ../../private/cli/gen-api/main.go vpc AddVPCNetwork

package vpc

import (
	"github.com/ucloud/ucloud-sdk-go/sdk/request"
	"github.com/ucloud/ucloud-sdk-go/sdk/response"
)

type AddVPCNetworkRequest struct {
	request.CommonBase

	// Required, 源VPC短ID
	VPCId string

	// Required, 增加网段
	Network []string
}

type AddVPCNetworkResponse struct {
	response.CommonBase
}

// NewAddVPCNetworkRequest will create request of AddVPCNetwork action.
func (c *VPCClient) NewAddVPCNetworkRequest() *AddVPCNetworkRequest {
	cfg := c.client.GetConfig()

	return &AddVPCNetworkRequest{
		CommonBase: request.CommonBase{
			Region:    cfg.Region,
			ProjectId: cfg.ProjectId,
		},
	}
}

// AddVPCNetwork - 添加VPC网段
func (c *VPCClient) AddVPCNetwork(req *AddVPCNetworkRequest) (*AddVPCNetworkResponse, error) {
	var err error
	var res AddVPCNetworkResponse

	err = c.client.InvokeAction("AddVPCNetwork", req, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
