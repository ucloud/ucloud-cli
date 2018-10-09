//Code is generated by ucloud code generator, don't modify it by hand, it will cause undefined behaviors.
//go:generate ucloud-gen-go-api UNet CreateBandwidthPackage

package unet

import (
	"github.com/ucloud/ucloud-sdk-go/sdk"
	"github.com/ucloud/ucloud-sdk-go/sdk/request"
	"github.com/ucloud/ucloud-sdk-go/sdk/response"
)

// CreateBandwidthPackageRequest is request schema for CreateBandwidthPackage action
type CreateBandwidthPackageRequest struct {
	request.CommonBase

	// 带宽大小(单位Mbps), 取值范围[2,800] (最大值受地域限制)
	Bandwidth *int `required:"true"`

	// 所绑定弹性IP的资源ID
	EIPId *string `required:"true"`

	// 带宽包有效时长, 取值范围为大于0的整数, 即该带宽包在EnableTime到 EnableTime+TimeRange时间段内生效
	TimeRange *int `required:"true"`

	// 生效时间, 格式为 Unix timestamp, 默认为立即开通
	EnableTime *int `required:"false"`

	// 代金券ID
	CouponId *string `required:"false"`
}

// CreateBandwidthPackageResponse is response schema for CreateBandwidthPackage action
type CreateBandwidthPackageResponse struct {
	response.CommonBase

	// 所创建带宽包的资源ID
	BandwidthPackageId string
}

// NewCreateBandwidthPackageRequest will create request of CreateBandwidthPackage action.
func (c *UNetClient) NewCreateBandwidthPackageRequest() *CreateBandwidthPackageRequest {
	cfg := c.client.GetConfig()

	return &CreateBandwidthPackageRequest{
		CommonBase: request.CommonBase{
			Region:    sdk.String(cfg.Region),
			ProjectId: sdk.String(cfg.ProjectId),
		},
	}
}

// CreateBandwidthPackage - 为非共享带宽模式下, 已绑定资源实例的带宽计费弹性IP附加临时带宽包
func (c *UNetClient) CreateBandwidthPackage(req *CreateBandwidthPackageRequest) (*CreateBandwidthPackageResponse, error) {
	var err error
	var res CreateBandwidthPackageResponse

	err = c.client.InvokeAction("CreateBandwidthPackage", req, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}