//go:generate go run ../../private/cli/gen-api/main.go unet UnBindEIP

package unet

import (
	"github.com/ucloud/ucloud-sdk-go/sdk/request"
	"github.com/ucloud/ucloud-sdk-go/sdk/response"
)

type UnBindEIPRequest struct {
	request.CommonBase

	// Required, 弹性IP的资源Id
	EIPId string

	// Required, 弹性IP请求解绑的资源类型, 枚举值为: uhost: 云主机; vrouter: 虚拟路由器; ulb, 负载均衡器 upm: 物理机; hadoophost: 大数据集群;fortresshost：堡垒机；udockhost：容器；udhost：私有专区主机；natgw：natgw；udb：udb；vpngw：ipsec vpn；ucdr：云灾备；dbaudit：数据库审计；
	ResourceType string

	// Required, 弹性IP请求解绑的资源ID
	ResourceId string
}

type UnBindEIPResponse struct {
	response.CommonBase
}

// NewUnBindEIPRequest will create request of UnBindEIP action.
func (c *UNetClient) NewUnBindEIPRequest() *UnBindEIPRequest {
	cfg := c.client.GetConfig()

	return &UnBindEIPRequest{
		CommonBase: request.CommonBase{
			Region:    cfg.Region,
			ProjectId: cfg.ProjectId,
		},
	}
}

// UnBindEIP - 将弹性IP从资源上解绑
func (c *UNetClient) UnBindEIP(req *UnBindEIPRequest) (*UnBindEIPResponse, error) {
	var err error
	var res UnBindEIPResponse

	err = c.client.InvokeAction("UnBindEIP", req, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
