//go:generate go run ../../private/cli/gen-api/main.go vpc CreateVPC

package vpc

import (
	"github.com/ucloud/ucloud-sdk-go/sdk/request"
	"github.com/ucloud/ucloud-sdk-go/sdk/response"
)

type CreateVPCRequest struct {
	request.CommonBase

	// Required, VPC名称
	Name string

	// Required, VPC网段
	Network []string

	// Optional, 业务组名称
	Tag string

	// Optional, 备注
	Remark string
}

type CreateVPCResponse struct {
	response.CommonBase

	// VPC资源Id
	VPCId string
}

// NewCreateVPCRequest will create request of CreateVPC action.
func (c *VPCClient) NewCreateVPCRequest() *CreateVPCRequest {
	cfg := c.client.GetConfig()

	return &CreateVPCRequest{
		CommonBase: request.CommonBase{
			Region:    cfg.Region,
			ProjectId: cfg.ProjectId,
		},
	}
}

// CreateVPC - 创建VPC
func (c *VPCClient) CreateVPC(req *CreateVPCRequest) (*CreateVPCResponse, error) {
	var err error
	var res CreateVPCResponse

	err = c.client.InvokeAction("CreateVPC", req, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
