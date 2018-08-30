package types

// ULBVServerSet - DescribeULB
type ULBVServerSet struct {

	// VServer实例的Id
	VServerId string

	// VServer实例的名字
	VServerName string

	// VServer实例的协议。 枚举值为：HTTP，TCP，UDP，HTTPS。
	Protocol string

	// VServer服务端口
	FrontendPort int

	// VServer负载均衡的模式，具体的值参见 CreateVServer
	Method string

	// VServer会话保持方式。枚举值为： None，关闭会话保持； ServerInsert，自动生成； UserDefined，用户自定义。
	PersistenceType string

	// 根据PersistenceType确定： None或ServerInsert，此字段为空； UserDefined，此字段展示用户自定义会话。
	PersistenceInfo string

	// 空闲连接的回收时间，单位：秒。
	ClientTimeout int

	// VServer的运行状态。枚举值： 0：运行正常;1：运行异常。
	Status int

	// VServer绑定的SSL证书信息，具体结构见下方 ULBSSLSet
	SSLSet []ULBSSLSet

	// 后端资源信息列表，具体结构见下方 ULBBackendSet
	VServerSet []ULBBackendSet

	// 监听器类型，枚举值为：RequestProxy：请求代理；PacketsTransmit：报文转发；默认为RequestProxy
	ListenType string
}
