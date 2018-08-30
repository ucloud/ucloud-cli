//go:generate go run ../../private/cli/gen-api/main.go ulb CreatePolicy

package ulb

import (
	"github.com/ucloud/ucloud-sdk-go/sdk/request"
	"github.com/ucloud/ucloud-sdk-go/sdk/response"
)

type CreatePolicyRequest struct {
	request.CommonBase

	// Required, 需要添加内容转发策略的负载均衡实例ID
	ULBId string

	// Required, 需要添加内容转发策略的VServer实例ID
	VServerId string

	// Required, 内容转发策略应用的后端资源实例的ID，来源于 AllocateBackend 返回的 BackendId
	BackendId []string

	// Required, 内容转发匹配字段
	Match string

	// Optional, 内容转发匹配字段的类型
	Type string
}

type CreatePolicyResponse struct {
	response.CommonBase

	// 内容转发策略ID
	PolicyId string
}

// NewCreatePolicyRequest will create request of CreatePolicy action.
func (c *ULBClient) NewCreatePolicyRequest() *CreatePolicyRequest {
	cfg := c.client.GetConfig()

	return &CreatePolicyRequest{
		CommonBase: request.CommonBase{
			Region:    cfg.Region,
			ProjectId: cfg.ProjectId,
		},
	}
}

// CreatePolicy - 创建VServer内容转发策略
func (c *ULBClient) CreatePolicy(req *CreatePolicyRequest) (*CreatePolicyResponse, error) {
	var err error
	var res CreatePolicyResponse

	err = c.client.InvokeAction("CreatePolicy", req, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
