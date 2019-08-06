package umem

/*
UMemcacheGroupSet - DescribeUMemcacheGroup

this model is auto created by ucloud code generater for open api,
you can also see https://docs.ucloud.cn for detail.
*/
type UMemcacheGroupSet struct {

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

	// Memcache版本信息,默认为1.4.31
	Version string

	// 状态标记 Creating // 初始化中 CreateFail // 创建失败 Deleting // 删除中 DeleteFail // 删除失败 Running // 运行 Resizing // 容量调整中 ResizeFail // 容量调整失败 Configing // 配置中 ConfigFail // 配置失败Restarting // 重启中
	State string

	// 创建时间 (UNIX时间戳)
	CreateTime int

	// 修改时间 (UNIX时间戳)
	ModifyTime int

	// 过期时间 (UNIX时间戳)
	ExpireTime int

	// 计费类型:Year,Month,Dynamic 默认Dynamic
	ChargeType string

	// 业务组名称
	Tag string

	// VPC ID
	VPCId string

	// Subnet ID
	SubnetId string
}
