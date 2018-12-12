package unet

/*
ResourceSet - 资源信息

this model is auto created by ucloud code generater for open api,
you can also see https://docs.ucloud.cn for detail.
*/
type ResourceSet struct {

	// 名称
	Name string

	// 内网IP
	PrivateIP string

	// 备注
	Remark string

	// 绑定该防火墙的资源id
	ResourceID string

	// 绑定资源的资源类型
	ResourceType string

	// 状态
	Status int

	// 业务组
	Tag string

	// 可用区
	Zone int
}
