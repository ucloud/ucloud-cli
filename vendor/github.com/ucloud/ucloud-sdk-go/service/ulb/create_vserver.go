//go:generate go run ../../private/cli/gen-api/main.go ulb CreateVServer

package ulb

import (
	"github.com/ucloud/ucloud-sdk-go/sdk/request"
	"github.com/ucloud/ucloud-sdk-go/sdk/response"
)

type CreateVServerRequest struct {
	request.CommonBase

	// Required, 负载均衡实例ID
	ULBId string

	// Optional, VServer实例名称，默认为"VServer"
	VServerName string

	// Optional, 监听器类型，枚举值为：RequestProxy：请求代理；PacketsTransmit：报文转发；默认为RequestProxy
	ListenType string

	// Optional, VServer实例的协议，请求代理模式下有 HTTP、HTTPS、TCP，报文转发下有 TCP，UDP
	Protocol string

	// Optional, VServer后端端口，取值范围[1-65535]；默认值为80
	FrontendPort int

	// Optional, VServer负载均衡模式， 默认为轮询模式，ConsistentHash，SourcePort，ConsistentHashPort 只在报文转发中使用；Roundrobin和Source在请求代理和报文转发中使用。
	Method string

	// Optional, VServer会话保持方式，默认关闭会话保持。枚举值：None：关闭；ServerInsert：自动生成KEY；UserDefined：用户自定义KEY。
	PersistenceType string

	// Optional, 根据PersistenceType确认； None和ServerInsert：此字段无意义； UserDefined：此字段传入自定义会话保持String
	PersistenceInfo string

	// Optional, ListenType为RequestProxy时表示空闲连接的回收时间，单位：秒，取值范围：时(0，86400]，默认值为60；ListenType为PacketsTransmit时表示连接保持的时间，单位：秒，取值范围：[60，900]，0 表示禁用连接保持
	ClientTimeout int
}

type CreateVServerResponse struct {
	response.CommonBase

	// VServer实例的Id
	VServerId string
}

// NewCreateVServerRequest will create request of CreateVServer action.
func (c *ULBClient) NewCreateVServerRequest() *CreateVServerRequest {
	cfg := c.client.GetConfig()

	return &CreateVServerRequest{
		CommonBase: request.CommonBase{
			Region:    cfg.Region,
			ProjectId: cfg.ProjectId,
		},
	}
}

// CreateVServer - 创建VServer实例，定义监听的协议和端口以及负载均衡算法
func (c *ULBClient) CreateVServer(req *CreateVServerRequest) (*CreateVServerResponse, error) {
	var err error
	var res CreateVServerResponse

	err = c.client.InvokeAction("CreateVServer", req, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
