//go:generate go run ../../private/cli/gen-api/main.go pathx CreateGlobalSSHInstance

package pathx

import (
	"github.com/ucloud/ucloud-sdk-go/sdk/request"
	"github.com/ucloud/ucloud-sdk-go/sdk/response"
)

type CreateGlobalSSHInstanceRequest struct {
	request.CommonBase

	// Required, 填写支持SSH访问IP的地区名称，如“洛杉矶”，“新加坡”，“香港”，“东京”，“华盛顿”
	Area string

	// Required, 被SSH访问的IP
	TargetIP string

	// Required, SSH端口，禁止使用80，443等端口
	Port string

	// Optional, 备注信息
	Remark string

	// Optional, 支付方式，如按月、按年、按时
	ChargeType string

	// Optional, 购买数量
	Quantity string

	// Optional, 使用代金券可冲抵部分费用
	CouponId string
}

type CreateGlobalSSHInstanceResponse struct {
	response.CommonBase

	// 实例ID，资源唯一标识
	InstanceId string

	// 加速域名，访问该域名可就近接入
	AcceleratingDomain string
}

// NewCreateGlobalSSHInstanceRequest will create request of CreateGlobalSSHInstance action.
func (c *PathXClient) NewCreateGlobalSSHInstanceRequest() *CreateGlobalSSHInstanceRequest {
	cfg := c.client.GetConfig()

	return &CreateGlobalSSHInstanceRequest{
		CommonBase: request.CommonBase{
			Region:    cfg.Region,
			ProjectId: cfg.ProjectId,
		},
	}
}

// CreateGlobalSSHInstance - 创建GlobalSSH实例
func (c *PathXClient) CreateGlobalSSHInstance(req *CreateGlobalSSHInstanceRequest) (*CreateGlobalSSHInstanceResponse, error) {
	var err error
	var res CreateGlobalSSHInstanceResponse

	err = c.client.InvokeAction("CreateGlobalSSHInstance", req, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
