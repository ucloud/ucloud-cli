//Code is generated by ucloud code generator, don't modify it by hand, it will cause undefined behaviors.
//go:generate ucloud-gen-go-api VPC CreateSubnet

package vpc

import (
	"github.com/ucloud/ucloud-sdk-go/sdk"
	"github.com/ucloud/ucloud-sdk-go/sdk/request"
	"github.com/ucloud/ucloud-sdk-go/sdk/response"
)

// CreateSubnetRequest is request schema for CreateSubnet action
type CreateSubnetRequest struct {
	request.CommonBase

	// VPC资源ID
	VPCId *string `required:"true"`

	// 子网网络地址，例如192.168.0.0
	Subnet *string `required:"true"`

	// 子网网络号位数，默认为24
	Netmask *int `required:"false"`

	// 子网名称，默认为Subnet
	SubnetName *string `required:"false"`

	// 业务组名称，默认为Default
	Tag *string `required:"false"`

	// 备注
	Remark *string `required:"false"`
}

// CreateSubnetResponse is response schema for CreateSubnet action
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
			Region:    sdk.String(cfg.Region),
			ProjectId: sdk.String(cfg.ProjectId),
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