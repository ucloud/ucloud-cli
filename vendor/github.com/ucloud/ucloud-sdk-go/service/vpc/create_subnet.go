//go:generate go run ../../private/cli/gen-api/main.go vpc CreateSubnet

package vpc

import (
	"github.com/ucloud/ucloud-sdk-go/sdk/request"
	"github.com/ucloud/ucloud-sdk-go/sdk/response"
)

type CreateSubnetRequest struct {
	request.CommonBase

	// Required, VPC资源ID
	VPCId string

	// Required, 子网网络地址，例如192.168.0.0
	Subnet string

	// Optional, 子网网络号位数，默认为24
	Netmask int

	// Optional, 子网名称，默认为Subnet
	SubnetName string

	// Optional, 业务组名称，默认为Default
	Tag string

	// Optional, 备注
	Remark string
}

type CreateSubnetResponse struct {
	response.CommonBase

	// 子网ID
	SubnetId string
}

// NewCreateSubnetRequest will create request of CreateSubnet action.
func (c *VPCClient) NewCreateSubnetRequest() *CreateSubnetRequest {
	cfg := c.client.GetConfig()

	return &CreateSubnetRequest{
		CommonBase: request.CommonBase{
			Region:    cfg.Region,
			ProjectId: cfg.ProjectId,
		},
	}
}

// CreateSubnet - 创建子网
func (c *VPCClient) CreateSubnet(req *CreateSubnetRequest) (*CreateSubnetResponse, error) {
	var err error
	var res CreateSubnetResponse

	err = c.client.InvokeAction("CreateSubnet", req, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
