package umem

/*
URedisConfigSet - 主备Redis配置文件信息

this model is auto created by ucloud code generater for open api,
you can also see https://docs.ucloud.cn for detail.
*/
type URedisConfigSet struct {

	// 配置ID
	ConfigId string

	// 配置名称
	Name string

	// 配置描述
	Description string

	// 配置对应的Redis版本
	Version string

	// 置是否可以修改
	IsModify string

	// 配置所处的状态
	State string

	// 创建时间 (UNIX时间戳)
	CreateTime int

	// 修改时间 (UNIX时间戳)
	ModifyTime int

	// 是否是跨机房URedis(默认false)
	RegionFlag bool

	// 配置文件所在的可用区
	Zone string
}
