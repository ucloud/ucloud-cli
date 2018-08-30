//go:generate go run ../../private/cli/gen-api/main.go uaccount GetProjectList

package uaccount

import (
	"github.com/ucloud/ucloud-sdk-go/sdk/request"
	"github.com/ucloud/ucloud-sdk-go/sdk/response"
	. "github.com/ucloud/ucloud-sdk-go/service/uaccount/types"
)

type GetProjectListRequest struct {
	request.CommonBase

	// Optional, 是否是财务账号(Yes: 是, No: 否)
	IsFinance string
}

type GetProjectListResponse struct {
	response.CommonBase

	// 项目总数
	ProjectCount int

	// JSON格式的项目列表实例
	ProjectSet []ProjectListInfo
}

// NewGetProjectListRequest will create request of GetProjectList action.
func (c *UAccountClient) NewGetProjectListRequest() *GetProjectListRequest {
	cfg := c.client.GetConfig()

	return &GetProjectListRequest{
		CommonBase: request.CommonBase{
			Region:    cfg.Region,
			ProjectId: cfg.ProjectId,
		},
	}
}

// GetProjectList - 获取项目列表
func (c *UAccountClient) GetProjectList(req *GetProjectListRequest) (*GetProjectListResponse, error) {
	var err error
	var res GetProjectListResponse

	err = c.client.InvokeAction("GetProjectList", req, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
