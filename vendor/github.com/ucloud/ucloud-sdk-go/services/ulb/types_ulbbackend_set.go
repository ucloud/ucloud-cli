package ulb

/*
	ULBBackendSet - DescribeULB

	this model is auto created by ucloud code generater for open api,
	you can also see https://docs.ucloud.cn for detail.
*/
type ULBBackendSet struct {

	// 后端资源实例的Id
	BackendId string

	// 后端资源实例的类型
	ResourceType string

	// 后端资源实例的资源Id
	ResourceId string

	// 后端资源实例的资源名字
	ResourceName string

	// 后端资源实例的内网IP
	PrivateIP string

	// 后端资源实例服务的端口
	Port int

	// 后端资源实例的启用与否
	Enabled int

	// 后端资源实例的运行状态
	Status int

	// 后端资源实例的资源所在的子网的ID
	SubnetId string
}
