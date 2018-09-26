package ulb

/*
	UlbPolicyBackendSet - DescribePolicyGroup

	this model is auto created by ucloud code generater for open api,
	you can also see https://docs.ucloud.cn for detail.
*/
type UlbPolicyBackendSet struct {

	// 后端资源实例的ID
	BackendId string

	// 后端资源实例的内网IP
	PrivateIP string

	// 后端资源实例的服务端口
	Port int
}
