//go:generate go run ../../private/cli/gen-api/main.go pathx DeleteGlobalSSHInstance

package pathx

import (
	"github.com/ucloud/ucloud-sdk-go/sdk/request"
	"github.com/ucloud/ucloud-sdk-go/sdk/response"
)

type DeleteGlobalSSHInstanceRequest struct {
	request.CommonBase

	// Required, 实例Id,资源的唯一标识
	InstanceId string
}

type DeleteGlobalSSHInstanceResponse struct {
	response.CommonBase
}

// NewDeleteGlobalSSHInstanceRequest will create request of DeleteGlobalSSHInstance action.
func (c *PathXClient) NewDeleteGlobalSSHInstanceRequest() *DeleteGlobalSSHInstanceRequest {
	cfg := c.client.GetConfig()

	return &DeleteGlobalSSHInstanceRequest{
		CommonBase: request.CommonBase{
			Region:    cfg.Region,
			ProjectId: cfg.ProjectId,
		},
	}
}

// DeleteGlobalSSHInstance - 删除GlobalSSH实例
func (c *PathXClient) DeleteGlobalSSHInstance(req *DeleteGlobalSSHInstanceRequest) (*DeleteGlobalSSHInstanceResponse, error) {
	var err error
	var res DeleteGlobalSSHInstanceResponse

	err = c.client.InvokeAction("DeleteGlobalSSHInstance", req, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
