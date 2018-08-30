//go:generate go run ../../private/cli/gen-api/main.go uaccount GetRegion

package uaccount

import (
	"github.com/ucloud/ucloud-sdk-go/sdk/request"
	"github.com/ucloud/ucloud-sdk-go/sdk/response"
	. "github.com/ucloud/ucloud-sdk-go/service/uaccount/types"
)

type GetRegionRequest struct {
	request.CommonBase
}

type GetRegionResponse struct {
	response.CommonBase

	// 各数据中心信息
	Regions []RegionInfo
}

// NewGetRegionRequest will create request of GetRegion action.
func (c *UAccountClient) NewGetRegionRequest() *GetRegionRequest {
	cfg := c.client.GetConfig()

	return &GetRegionRequest{
		CommonBase: request.CommonBase{
			Region:    cfg.Region,
			ProjectId: cfg.ProjectId,
		},
	}
}

// GetRegion - 获取用户在各数据中心的权限等信息
func (c *UAccountClient) GetRegion(req *GetRegionRequest) (*GetRegionResponse, error) {
	var err error
	var res GetRegionResponse

	err = c.client.InvokeAction("GetRegion", req, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
