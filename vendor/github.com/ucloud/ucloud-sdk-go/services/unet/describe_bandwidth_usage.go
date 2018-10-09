//Code is generated by ucloud code generator, don't modify it by hand, it will cause undefined behaviors.
//go:generate ucloud-gen-go-api UNet DescribeBandwidthUsage

package unet

import (
	"github.com/ucloud/ucloud-sdk-go/sdk"
	"github.com/ucloud/ucloud-sdk-go/sdk/request"
	"github.com/ucloud/ucloud-sdk-go/sdk/response"
)

// DescribeBandwidthUsageRequest is request schema for DescribeBandwidthUsage action
type DescribeBandwidthUsageRequest struct {
	request.CommonBase

	// 返回数据分页值, 取值范围为 [0,10000000] 之间的整数, 默认为20
	Limit *int `required:"false"`

	// 返回数据偏移量, 默认为0
	OffSet *int `required:"false"`

	// 弹性IP的资源Id. 如果为空, 则返回当前 Region中符合条件的所有EIP的带宽用量, n为自然数
	EIPIds []string `required:"false"`
}

// DescribeBandwidthUsageResponse is response schema for DescribeBandwidthUsage action
type DescribeBandwidthUsageResponse struct {
	response.CommonBase

	// EIPSet中的元素个数
	TotalCount int

	// 单个弹性IP的带宽用量详细信息, 详见 UnetBandwidthUsageEIPSet, 如没有弹性IP资源则没有该返回值。
	EIPSet []UnetBandwidthUsageEIPSet
}

// NewDescribeBandwidthUsageRequest will create request of DescribeBandwidthUsage action.
func (c *UNetClient) NewDescribeBandwidthUsageRequest() *DescribeBandwidthUsageRequest {
	cfg := c.client.GetConfig()

	return &DescribeBandwidthUsageRequest{
		CommonBase: request.CommonBase{
			Region:    sdk.String(cfg.Region),
			ProjectId: sdk.String(cfg.ProjectId),
		},
	}
}

// DescribeBandwidthUsage - 获取带宽用量信息
func (c *UNetClient) DescribeBandwidthUsage(req *DescribeBandwidthUsageRequest) (*DescribeBandwidthUsageResponse, error) {
	var err error
	var res DescribeBandwidthUsageResponse

	err = c.client.InvokeAction("DescribeBandwidthUsage", req, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}