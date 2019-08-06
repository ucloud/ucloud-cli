package udb

/*
UDBInstanceByParamGroup - DescribeUDBInstanceByParamGroup

this model is auto created by ucloud code generater for open api,
you can also see https://docs.ucloud.cn for detail.
*/
type UDBInstanceByParamGroup struct {

	// DB实例Id
	DBId string

	// 实例名称
	Name string

	// DB实例虚ip
	VirtualIP string

	// 端口号
	Port string

	// DB状态标记
	State string

	// DB实例创建时间
	CreateTime string

	// DB实例过期时间
	ExpiredTime string

	// DB实例角色
	Role string
}
