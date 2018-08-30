//go:generate go run ../../private/cli/gen-api/main.go pathx DescribeGlobalSSHInstance

package pathx

import (
	"github.com/ucloud/ucloud-sdk-go/sdk/request"
	"github.com/ucloud/ucloud-sdk-go/sdk/response"
	. "github.com/ucloud/ucloud-sdk-go/service/pathx/types"
)

type DescribeGlobalSSHInstanceRequest struct {
	request.CommonBase

	// Optional, 实例ID，资源唯一标识
	InstanceId string
}

type DescribeGlobalSSHInstanceResponse struct {
	response.CommonBase

	// GlobalSSH实例列表，实例的属性参考GlobalSSHInfo模型
	InstanceSet []GlobalSSHInfo
}

// NewDescribeGlobalSSHInstanceRequest will create request of DescribeGlobalSSHInstance action.
func (c *PathXClient) NewDescribeGlobalSSHInstanceRequest() *DescribeGlobalSSHInstanceRequest {
	cfg := c.client.GetConfig()

	return &DescribeGlobalSSHInstanceRequest{
		CommonBase: request.CommonBase{
			Region:    cfg.Region,
			ProjectId: cfg.ProjectId,
		},
	}
}

// DescribeGlobalSSHInstance - 获取GlobalSSH实例列表（传实例ID获取单个实例信息，不传获取项目下全部实例）
func (c *PathXClient) DescribeGlobalSSHInstance(req *DescribeGlobalSSHInstanceRequest) (*DescribeGlobalSSHInstanceResponse, error) {
	var err error
	var res DescribeGlobalSSHInstanceResponse

	err = c.client.InvokeAction("DescribeGlobalSSHInstance", req, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
