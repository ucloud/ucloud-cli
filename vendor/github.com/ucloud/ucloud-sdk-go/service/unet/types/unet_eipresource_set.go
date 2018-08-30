package types

// UnetEIPResourceSet - DescribeEIP
type UnetEIPResourceSet struct {

	// 已绑定的资源类型, 枚举值为: uhost, 云主机；vrouter：虚拟路由器；ulb：负载均衡器
	ResourceType string

	// 已绑定的资源名称
	ResourceName string

	// 已绑定资源的资源ID
	ResourceId string

	// 弹性IP的资源ID
	EIPId string
}
