//go:generate go run ../../private/cli/gen-api/main.go ulb CreateULB

package ulb

import (
	"github.com/ucloud/ucloud-sdk-go/sdk/request"
	"github.com/ucloud/ucloud-sdk-go/sdk/response"
)

type CreateULBRequest struct {
	request.CommonBase

	// Optional, 负载均衡的名字，默认值为“ULB”
	ULBName string

	// Optional, 业务组
	Tag string

	// Optional, 备注
	Remark string

	// Optional, 创建的ULB是否为外网模式，默认即为外网模式
	OuterMode string

	// Optional, 创建的ULB是否为内网模式
	InnerMode string

	// Optional, 付费方式
	ChargeType string

	// Optional, ULB所在的VPC的ID, 如果不传则使用默认的VPC
	VPCId string

	// Optional, 内网ULB 所属的子网ID，如果不传则使用默认的子网
	SubnetId string

	// Optional, ULB 所属的业务组ID，如果不传则使用默认的业务组
	BusinessId string
}

type CreateULBResponse struct {
	response.CommonBase

	// 负载均衡实例的Id
	ULBId string
}

// NewCreateULBRequest will create request of CreateULB action.
func (c *ULBClient) NewCreateULBRequest() *CreateULBRequest {
	cfg := c.client.GetConfig()

	return &CreateULBRequest{
		CommonBase: request.CommonBase{
			Region:    cfg.Region,
			ProjectId: cfg.ProjectId,
		},
	}
}

// CreateULB - 创建负载均衡实例，可以选择内网或者外网
func (c *ULBClient) CreateULB(req *CreateULBRequest) (*CreateULBResponse, error) {
	var err error
	var res CreateULBResponse

	err = c.client.InvokeAction("CreateULB", req, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
