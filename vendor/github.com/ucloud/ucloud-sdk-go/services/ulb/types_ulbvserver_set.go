package ulb

/*
ULBVServerSet - DescribeULB

this model is auto created by ucloud code generater for open api,
you can also see https://docs.ucloud.cn for detail.
*/
type ULBVServerSet struct {

	// 健康检查类型，枚举值：Port -> 端口检查；Path -> 路径检查；
	MonitorType string

	// 根据MonitorType确认； 当MonitorType为Port时，此字段无意义。当MonitorType为Path时，代表HTTP检查路径
	Domain string

	// 根据MonitorType确认； 当MonitorType为Port时，此字段无意义。当MonitorType为Path时，代表HTTP检查域名
	Path string

	// VServer实例的Id
	VServerId string

	// VServer实例的名字
	VServerName string

	// VServer实例的协议。 枚举值为：HTTP，TCP，UDP，HTTPS。
	Protocol string

	// VServer服务端口
	FrontendPort int

	// VServer负载均衡的模式，枚举值：Roundrobin -> 轮询;Source -> 源地址；ConsistentHash -> 一致性哈希；SourcePort -> 源地址（计算端口）；ConsistentHashPort -> 一致性哈希（计算端口）。
	Method string

	// VServer会话保持方式。枚举值为： None -> 关闭会话保持； ServerInsert -> 自动生成； UserDefined -> 用户自定义。
	PersistenceType string

	// 根据PersistenceType确定： None或ServerInsert，此字段为空； UserDefined，此字段展示用户自定义会话string。
	PersistenceInfo string

	// 空闲连接的回收时间，单位：秒。
	ClientTimeout int

	// VServer的运行状态。枚举值： 0 -> rs全部运行正常;1 -> rs部分运行正常；2 -> rs全部运行异常。
	Status int

	// VServer绑定的SSL证书信息，具体结构见下方 ULBSSLSet
	SSLSet []ULBSSLSet

	// 后端资源信息列表，具体结构见下方 ULBBackendSet
	BackendSet []ULBBackendSet

	// 监听器类型，枚举值为: RequestProxy -> 请求代理；PacketsTransmit -> 报文转发
	ListenType string

	// 内容转发信息列表，具体结构见下方 ULBPolicySet
	PolicySet []ULBPolicySet
}
