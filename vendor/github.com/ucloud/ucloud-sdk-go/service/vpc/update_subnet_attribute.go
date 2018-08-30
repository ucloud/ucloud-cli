//go:generate go run ../../private/cli/gen-api/main.go vpc UpdateSubnetAttribute

package vpc

import (
	"github.com/ucloud/ucloud-sdk-go/sdk/request"
	"github.com/ucloud/ucloud-sdk-go/sdk/response"
)

type UpdateSubnetAttributeRequest struct {
	request.CommonBase

	// Required, 子网ID
	SubnetId string

	// Optional, 子网名称(如果Name不填写，Tag必须填写)
	Name string

	// Optional, 业务组名称(如果Tag不填写，Name必须填写)
	Tag string
}

type UpdateSubnetAttributeResponse struct {
	response.CommonBase
}

// NewUpdateSubnetAttributeRequest will create request of UpdateSubnetAttribute action.
func (c *VPCClient) NewUpdateSubnetAttributeRequest() *UpdateSubnetAttributeRequest {
	cfg := c.client.GetConfig()

	return &UpdateSubnetAttributeRequest{
		CommonBase: request.CommonBase{
			Region:    cfg.Region,
			ProjectId: cfg.ProjectId,
		},
	}
}

// UpdateSubnetAttribute - 更新子网信息
func (c *VPCClient) UpdateSubnetAttribute(req *UpdateSubnetAttributeRequest) (*UpdateSubnetAttributeResponse, error) {
	var err error
	var res UpdateSubnetAttributeResponse

	err = c.client.InvokeAction("UpdateSubnetAttribute", req, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
