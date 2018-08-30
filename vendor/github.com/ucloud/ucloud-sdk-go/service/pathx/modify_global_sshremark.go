//go:generate go run ../../private/cli/gen-api/main.go pathx ModifyGlobalSSHRemark

package pathx

import (
	"github.com/ucloud/ucloud-sdk-go/sdk/request"
	"github.com/ucloud/ucloud-sdk-go/sdk/response"
)

type ModifyGlobalSSHRemarkRequest struct {
	request.CommonBase

	// Required, 实例ID,资源唯一标识
	InstanceId string

	// Optional, 备注信息，不填默认为空字符串
	Remark string
}

type ModifyGlobalSSHRemarkResponse struct {
	response.CommonBase
}

// NewModifyGlobalSSHRemarkRequest will create request of ModifyGlobalSSHRemark action.
func (c *PathXClient) NewModifyGlobalSSHRemarkRequest() *ModifyGlobalSSHRemarkRequest {
	cfg := c.client.GetConfig()

	return &ModifyGlobalSSHRemarkRequest{
		CommonBase: request.CommonBase{
			Region:    cfg.Region,
			ProjectId: cfg.ProjectId,
		},
	}
}

// ModifyGlobalSSHRemark - 修改GlobalSSH备注
func (c *PathXClient) ModifyGlobalSSHRemark(req *ModifyGlobalSSHRemarkRequest) (*ModifyGlobalSSHRemarkResponse, error) {
	var err error
	var res ModifyGlobalSSHRemarkResponse

	err = c.client.InvokeAction("ModifyGlobalSSHRemark", req, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
