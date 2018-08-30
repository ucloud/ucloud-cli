//go:generate go run ../../private/cli/gen-api/main.go ulb ReleaseBackend

package ulb

import (
	"github.com/ucloud/ucloud-sdk-go/sdk/request"
	"github.com/ucloud/ucloud-sdk-go/sdk/response"
)

type ReleaseBackendRequest struct {
	request.CommonBase

	// Required, 负载均衡实例的ID
	ULBId string

	// Required, 后端资源实例的ID(ULB后端ID，非资源自身ID)
	BackendId string
}

type ReleaseBackendResponse struct {
	response.CommonBase
}

// NewReleaseBackendRequest will create request of ReleaseBackend action.
func (c *ULBClient) NewReleaseBackendRequest() *ReleaseBackendRequest {
	cfg := c.client.GetConfig()

	return &ReleaseBackendRequest{
		CommonBase: request.CommonBase{
			Region:    cfg.Region,
			ProjectId: cfg.ProjectId,
		},
	}
}

// ReleaseBackend - 从VServer释放后端资源实例
func (c *ULBClient) ReleaseBackend(req *ReleaseBackendRequest) (*ReleaseBackendResponse, error) {
	var err error
	var res ReleaseBackendResponse

	err = c.client.InvokeAction("ReleaseBackend", req, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
