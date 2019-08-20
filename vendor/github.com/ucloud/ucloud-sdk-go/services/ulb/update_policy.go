//Code is generated by ucloud code generator, don't modify it by hand, it will cause undefined behaviors.
//go:generate ucloud-gen-go-api ULB UpdatePolicy

package ulb

import (
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"
	"github.com/ucloud/ucloud-sdk-go/ucloud/response"
)

// UpdatePolicyRequest is request schema for UpdatePolicy action
type UpdatePolicyRequest struct {
	request.CommonBase

	// [公共参数] 地域。 参见 [地域和可用区列表](../summary/regionlist.html)
	// Region *string `required:"true"`

	// [公共参数] 项目ID。不填写为默认项目，子帐号必须填写。 请参考[GetProjectList接口](../summary/get_project_list.html)
	// ProjectId *string `required:"true"`

	// 需要添加内容转发策略的负载均衡实例ID
	ULBId *string `required:"true"`

	// 需要添加内容转发策略的VServer实例ID
	VServerId *string `required:"true"`

	// 转发规则的ID
	PolicyId *string `required:"true"`

	// 内容转发策略应用的后端资源实例的ID，来源于 AllocateBackend 返回的 BackendId
	BackendId []string `required:"true"`

	// 内容转发匹配字段
	Match *string `required:"true"`

	// 内容转发匹配字段的类型
	Type *string `required:"false"`
}

// UpdatePolicyResponse is response schema for UpdatePolicy action
type UpdatePolicyResponse struct {
	response.CommonBase

	// 转发规则的ID
	PolicyId string
}

// NewUpdatePolicyRequest will create request of UpdatePolicy action.
func (c *ULBClient) NewUpdatePolicyRequest() *UpdatePolicyRequest {
	req := &UpdatePolicyRequest{}

	// setup request with client config
	c.Client.SetupRequest(req)

	// setup retryable with default retry policy (retry for non-create action and common error)
	req.SetRetryable(true)
	return req
}

// UpdatePolicy - 更新内容转发规则，包括转发规则后的服务节点
func (c *ULBClient) UpdatePolicy(req *UpdatePolicyRequest) (*UpdatePolicyResponse, error) {
	var err error
	var res UpdatePolicyResponse

	err = c.Client.InvokeAction("UpdatePolicy", req, &res)
	if err != nil {
		return &res, err
	}

	return &res, nil
}