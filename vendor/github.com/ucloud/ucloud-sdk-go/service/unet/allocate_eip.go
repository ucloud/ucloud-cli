//go:generate go run ../../private/cli/gen-api/main.go unet AllocateEIP

package unet

import (
	"github.com/ucloud/ucloud-sdk-go/sdk/request"
	"github.com/ucloud/ucloud-sdk-go/sdk/response"
	. "github.com/ucloud/ucloud-sdk-go/service/unet/types"
)

type AllocateEIPRequest struct {
	request.CommonBase

	// Required, 弹性IP的线路如下: 国际: International BGP: Bgp  各地域允许的线路参数如下:  cn-sh1: Bgp cn-sh2: Bgp cn-gd: Bgp cn-bj1: Bgp cn-bj2: Bgp hk: International us-ca: International th-bkk: International  kr-seoul:International  us-ws:International  ge-fra:International  sg:International  tw-kh:International
	OperatorName string

	// Required, 弹性IP的外网带宽, 单位为Mbps. 共享带宽模式必须指定0M带宽, 非共享带宽模式必须指定非0Mbps带宽. 各地域非共享带宽的带宽范围如下： 流量计费[1-200]，带宽计费[1-800]
	Bandwidth int

	// Optional, 业务组名称, 默认为 "Default"
	Tag string

	// Optional, 付费方式, 枚举值为: Year, 按年付费; Month, 按月付费; Dynamic, 按需付费(需开启权限); Trial, 试用(需开启权限) 默认为按月付费
	ChargeType string

	// Optional, 购买时长, 默认: 1
	Quantity int

	// Optional, 弹性IP的计费模式. 枚举值: "Traffic", 流量计费; "Bandwidth", 带宽计费; "ShareBandwidth",共享带宽模式. 默认为 "Bandwidth".
	PayMode string

	// Optional, 绑定的共享带宽Id，仅当PayMode为ShareBandwidth时有效
	ShareBandwidthId string

	// Optional, 弹性IP的名称, 默认为 "EIP"
	Name string

	// Optional, 弹性IP的备注, 默认为空
	Remark string

	// Optional, 代金券ID, 默认不使用
	CouponId string
}

type AllocateEIPResponse struct {
	response.CommonBase

	// 申请到的EIP资源详情 参见 UnetAllocateEIPSet
	EIPSet []UnetAllocateEIPSet
}

// NewAllocateEIPRequest will create request of AllocateEIP action.
func (c *UNetClient) NewAllocateEIPRequest() *AllocateEIPRequest {
	cfg := c.client.GetConfig()

	return &AllocateEIPRequest{
		CommonBase: request.CommonBase{
			Region:    cfg.Region,
			ProjectId: cfg.ProjectId,
		},
	}
}

// AllocateEIP - 根据提供信息, 申请弹性IP
func (c *UNetClient) AllocateEIP(req *AllocateEIPRequest) (*AllocateEIPResponse, error) {
	var err error
	var res AllocateEIPResponse

	err = c.client.InvokeAction("AllocateEIP", req, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
