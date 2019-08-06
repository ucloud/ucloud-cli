package umem

/*
URedisGroupSet - DescribeURedisGroup

this model is auto created by ucloud code generater for open api,
you can also see https://docs.ucloud.cn for detail.
*/
type URedisGroupSet struct {

	// 组ID
	GroupId string

	// 组名称
	Name string

	// 节点的配置ID
	ConfigId string

	// 节点的虚拟IP地址
	VirtualIP string

	// 节点分配的服务端口
	Port int

	// 容量单位GB
	Size int

	// 使用量单位MB
	UsedSize int

	// 是否需要自动备份,enable,disable
	AutoBackup string

	// 组自动备份开始时间,单位小时计,范围[0-23]
	BackupTime int

	// 是否开启高可用,enable,disable
	HighAvailability string

	// Redis版本信息
	Version string

	// 过期时间 (UNIX时间戳)
	ExpireTime int

	// 计费类型:Year,Month,Dynamic 默认Dynamic
	ChargeType string

	// 状态标记 Creating // 初始化中 CreateFail // 创建失败 Deleting // 删除中 DeleteFail // 删除失败 Running // 运行 Resizing // 容量调整中 ResizeFail // 容量调整失败 Configing // 配置中 ConfigFail // 配置失败
	State string

	// 创建时间 (UNIX时间戳)
	CreateTime int

	// 修改时间 (UNIX时间戳)
	ModifyTime int

	// 业务组名称
	Tag string

	// 实例所在可用区，或者master redis所在可用区，参见 [可用区列表](../summary/regionlist.html)
	Zone string

	// 跨机房URedis，slave redis所在可用区，参见 [可用区列表](../summary/regionlist.html)
	SlaveZone string

	// VPC ID
	VPCId string

	// Subnet ID
	SubnetId string
}
