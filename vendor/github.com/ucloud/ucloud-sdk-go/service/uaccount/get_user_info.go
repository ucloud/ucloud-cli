//go:generate go run ../../private/cli/gen-api/main.go uaccount GetUserInfo

package uaccount

import (
	"github.com/ucloud/ucloud-sdk-go/sdk/request"
	"github.com/ucloud/ucloud-sdk-go/sdk/response"
	. "github.com/ucloud/ucloud-sdk-go/service/uaccount/types"
)

type GetUserInfoRequest struct {
	request.CommonBase
}

type GetUserInfoResponse struct {
	response.CommonBase

	// 用户信息返回数组
	DataSet []UserInfo
}

// NewGetUserInfoRequest will create request of GetUserInfo action.
func (c *UAccountClient) NewGetUserInfoRequest() *GetUserInfoRequest {
	cfg := c.client.GetConfig()

	return &GetUserInfoRequest{
		CommonBase: request.CommonBase{
			Region:    cfg.Region,
			ProjectId: cfg.ProjectId,
		},
	}
}

// GetUserInfo - 获取用户信息
func (c *UAccountClient) GetUserInfo(req *GetUserInfoRequest) (*GetUserInfoResponse, error) {
	var err error
	var res GetUserInfoResponse

	err = c.client.InvokeAction("GetUserInfo", req, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
