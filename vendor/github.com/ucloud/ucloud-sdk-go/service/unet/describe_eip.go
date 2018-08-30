//go:generate go run ../../private/cli/gen-api/main.go unet DescribeEIP

package unet

import (
	"github.com/ucloud/ucloud-sdk-go/sdk/request"
	"github.com/ucloud/ucloud-sdk-go/sdk/response"
	. "github.com/ucloud/ucloud-sdk-go/service/unet/types"
)

type DescribeEIPRequest struct {
	request.CommonBase

	// Optional, 弹性IP的资源ID如果为空, 则返回当前 Region中符合条件的的所有EIP
	EIPIds []string

	// Optional, 数据偏移量, 默认为0
	Offset int

	// Optional, 数据分页值, 默认为20
	Limit int
}

type DescribeEIPResponse struct {
	response.CommonBase

	// 满足条件的弹性IP总数
	TotalCount int

	// 满足条件的弹性IP带宽总和, 单位Mbps
	TotalBandwidth int

	// 弹性IP列表, 每项参数详见 UnetEIPSet
	EIPSet []UnetEIPSet
}

// NewDescribeEIPRequest will create request of DescribeEIP action.
func (c *UNetClient) NewDescribeEIPRequest() *DescribeEIPRequest {
	cfg := c.client.GetConfig()

	return &DescribeEIPRequest{
		CommonBase: request.CommonBase{
			Region:    cfg.Region,
			ProjectId: cfg.ProjectId,
		},
	}
}

// DescribeEIP - 获取弹性IP信息
func (c *UNetClient) DescribeEIP(req *DescribeEIPRequest) (*DescribeEIPResponse, error) {
	var err error
	var res DescribeEIPResponse

	err = c.client.InvokeAction("DescribeEIP", req, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
