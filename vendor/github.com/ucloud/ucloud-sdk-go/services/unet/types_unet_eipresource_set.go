package unet

/*
	UnetEIPResourceSet - DescribeEIP

	this model is auto created by ucloud code generater for open api,
	you can also see https://docs.ucloud.cn for detail.
*/
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
