//go:generate go run ../../private/cli/gen-api/main.go uhost StopUHostInstance

package uhost

import (
	"github.com/ucloud/ucloud-sdk-go/sdk/request"
	"github.com/ucloud/ucloud-sdk-go/sdk/response"
)

type StopUHostInstanceRequest struct {
	request.CommonBase

	// Optional, 可用区。参见 [可用区列表](../summary/regionlist.html)
	Zone string

	// Required, UHost实例ID 参见 [DescribeUHostInstance](describe_uhost_instance.html)
	UHostId string
}

type StopUHostInstanceResponse struct {
	response.CommonBase

	// UHost实例ID
	UhostId string
}

// NewStopUHostInstanceRequest will create request of StopUHostInstance action.
func (c *UHostClient) NewStopUHostInstanceRequest() *StopUHostInstanceRequest {
	cfg := c.client.GetConfig()

	return &StopUHostInstanceRequest{
		CommonBase: request.CommonBase{
			Region:    cfg.Region,
			ProjectId: cfg.ProjectId,
		},
	}
}

// StopUHostInstance - 指停止处于运行状态的UHost实例，需指定数据中心及UhostID。
func (c *UHostClient) StopUHostInstance(req *StopUHostInstanceRequest) (*StopUHostInstanceResponse, error) {
	var err error
	var res StopUHostInstanceResponse

	err = c.client.InvokeAction("StopUHostInstance", req, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
