//go:generate go run ../../private/cli/gen-api/main.go ulb AllocateBackend

package ulb

import (
	"github.com/ucloud/ucloud-sdk-go/sdk/request"
	"github.com/ucloud/ucloud-sdk-go/sdk/response"
)

type AllocateBackendRequest struct {
	request.CommonBase

	// Required, 负载均衡实例的ID
	ULBId string

	// Required, VServer实例的ID
	VServerId string

	// Required, 所添加的后端资源的类型
	ResourceType string

	// Required, 所添加的后端资源的资源ID
	ResourceId string

	// Optional, 所添加的后端资源服务端口，取值范围[1-65535]，默认80
	Port int

	// Optional, 后端实例状态开关，枚举值： 1：启用； 0：禁用 默认为启用
	Enabled int
}

type AllocateBackendResponse struct {
	response.CommonBase

	// 所添加的后端资源在ULB中的对象ID，（为ULB系统中使用，与资源自身ID无关），可用于 UpdateBackendAttribute/UpdateBackendAttributeBatch/ReleaseBackend
	BackendId string
}

// NewAllocateBackendRequest will create request of AllocateBackend action.
func (c *ULBClient) NewAllocateBackendRequest() *AllocateBackendRequest {
	cfg := c.client.GetConfig()

	return &AllocateBackendRequest{
		CommonBase: request.CommonBase{
			Region:    cfg.Region,
			ProjectId: cfg.ProjectId,
		},
	}
}

// AllocateBackend - 添加ULB后端资源实例
func (c *ULBClient) AllocateBackend(req *AllocateBackendRequest) (*AllocateBackendResponse, error) {
	var err error
	var res AllocateBackendResponse

	err = c.client.InvokeAction("AllocateBackend", req, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
