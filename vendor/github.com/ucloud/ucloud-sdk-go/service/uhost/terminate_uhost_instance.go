//go:generate go run ../../private/cli/gen-api/main.go uhost TerminateUHostInstance

package uhost

import (
	"github.com/ucloud/ucloud-sdk-go/sdk/request"
	"github.com/ucloud/ucloud-sdk-go/sdk/response"
)

type TerminateUHostInstanceRequest struct {
	request.CommonBase

	// Optional, 可用区。参见 [可用区列表](../summary/regionlist.html)
	Zone string

	// Required, UHost资源Id 参见 [DescribeUHostInstance](describe_uhost_instance.html)
	UHostId string

	// Optional, 是否直接删除，0表示按照原来的逻辑（有回收站权限，则进入回收站），1表示直接删除
	Destroy int
}

type TerminateUHostInstanceResponse struct {
	response.CommonBase

	// UHost 实例 Id
	UHostIds []string

	// 放入回收站:"Yes", 彻底删除：“No”
	InRecycle string
}

// NewTerminateUHostInstanceRequest will create request of TerminateUHostInstance action.
func (c *UHostClient) NewTerminateUHostInstanceRequest() *TerminateUHostInstanceRequest {
	cfg := c.client.GetConfig()

	return &TerminateUHostInstanceRequest{
		CommonBase: request.CommonBase{
			Region:    cfg.Region,
			ProjectId: cfg.ProjectId,
		},
	}
}

// TerminateUHostInstance - 删除指定数据中心的UHost实例。
func (c *UHostClient) TerminateUHostInstance(req *TerminateUHostInstanceRequest) (*TerminateUHostInstanceResponse, error) {
	var err error
	var res TerminateUHostInstanceResponse

	err = c.client.InvokeAction("TerminateUHostInstance", req, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
