//go:generate go run ../../private/cli/gen-api/main.go pathx ModifyGlobalSSHPort

package pathx

import (
	"github.com/ucloud/ucloud-sdk-go/sdk/request"
	"github.com/ucloud/ucloud-sdk-go/sdk/response"
)

type ModifyGlobalSSHPortRequest struct {
	request.CommonBase

	// Required, 实例ID,资源唯一标识
	InstanceId string

	// Required, 调整后的SSH登陆端口
	Port string
}

type ModifyGlobalSSHPortResponse struct {
	response.CommonBase
}

// NewModifyGlobalSSHPortRequest will create request of ModifyGlobalSSHPort action.
func (c *PathXClient) NewModifyGlobalSSHPortRequest() *ModifyGlobalSSHPortRequest {
	cfg := c.client.GetConfig()

	return &ModifyGlobalSSHPortRequest{
		CommonBase: request.CommonBase{
			Region:    cfg.Region,
			ProjectId: cfg.ProjectId,
		},
	}
}

// ModifyGlobalSSHPort - 修改GlobalSSH端口
func (c *PathXClient) ModifyGlobalSSHPort(req *ModifyGlobalSSHPortRequest) (*ModifyGlobalSSHPortResponse, error) {
	var err error
	var res ModifyGlobalSSHPortResponse

	err = c.client.InvokeAction("ModifyGlobalSSHPort", req, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
