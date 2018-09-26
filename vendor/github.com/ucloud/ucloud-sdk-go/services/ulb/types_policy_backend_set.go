package ulb

/*
	PolicyBackendSet - 内容转发下rs详细信息

	this model is auto created by ucloud code generater for open api,
	you can also see https://docs.ucloud.cn for detail.
*/
type PolicyBackendSet struct {

	// 所添加的后端资源在ULB中的对象ID，（为ULB系统中使用，与资源自身ID无关
	BackendId string

	// 后端资源的对象ID
	ObjectId string

	// 所添加的后端资源服务端口
	Port int

	// 后端资源的内网IP
	PrivateIP string

	// 后端资源的实例名称
	ResourceName string
}
