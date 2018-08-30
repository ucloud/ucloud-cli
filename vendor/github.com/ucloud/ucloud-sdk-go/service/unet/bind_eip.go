//go:generate go run ../../private/cli/gen-api/main.go unet BindEIP

package unet

import (
	"github.com/ucloud/ucloud-sdk-go/sdk/request"
	"github.com/ucloud/ucloud-sdk-go/sdk/response"
)

type BindEIPRequest struct {
	request.CommonBase

	// Required, 弹性IP的资源Id
	EIPId string

	// Required, 弹性IP请求绑定的资源类型, 枚举值为: uhost: 云主机; vrouter: 虚拟路由器; ulb, 负载均衡器 upm: 物理机; hadoophost: 大数据集群;fortresshost：堡垒机；udockhost：容器；udhost：私有专区主机；natgw：natgw；udb：udb；vpngw：ipsec vpn；ucdr：云灾备；dbaudit：数据库审计；
	ResourceType string

	// Required, 弹性IP请求绑定的资源ID
	ResourceId string
}

type BindEIPResponse struct {
	response.CommonBase
}

// NewBindEIPRequest will create request of BindEIP action.
func (c *UNetClient) NewBindEIPRequest() *BindEIPRequest {
	cfg := c.client.GetConfig()

	return &BindEIPRequest{
		CommonBase: request.CommonBase{
			Region:    cfg.Region,
			ProjectId: cfg.ProjectId,
		},
	}
}

// BindEIP - 将尚未使用的弹性IP绑定到指定的资源
func (c *UNetClient) BindEIP(req *BindEIPRequest) (*BindEIPResponse, error) {
	var err error
	var res BindEIPResponse

	err = c.client.InvokeAction("BindEIP", req, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
