//Code is generated by ucloud code generator, don't modify it by hand, it will cause undefined behaviors.
//go:generate ucloud-gen-go-api UAccount TerminateProject

package uaccount

import (
	"github.com/ucloud/ucloud-sdk-go/sdk"
	"github.com/ucloud/ucloud-sdk-go/sdk/request"
	"github.com/ucloud/ucloud-sdk-go/sdk/response"
)

// TerminateProjectRequest is request schema for TerminateProject action
type TerminateProjectRequest struct {
	request.CommonBase
}

// TerminateProjectResponse is response schema for TerminateProject action
type TerminateProjectResponse struct {
	response.CommonBase
}

// NewTerminateProjectRequest will create request of TerminateProject action.
func (c *UAccountClient) NewTerminateProjectRequest() *TerminateProjectRequest {
	cfg := c.client.GetConfig()

	return &TerminateProjectRequest{
		CommonBase: request.CommonBase{
			Region:    sdk.String(cfg.Region),
			ProjectId: sdk.String(cfg.ProjectId),
		},
	}
}

// TerminateProject - 删除项目
func (c *UAccountClient) TerminateProject(req *TerminateProjectRequest) (*TerminateProjectResponse, error) {
	var err error
	var res TerminateProjectResponse

	err = c.client.InvokeAction("TerminateProject", req, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
