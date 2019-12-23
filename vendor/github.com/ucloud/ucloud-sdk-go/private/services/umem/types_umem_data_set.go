package umem

/*
UMemDataSet - DescribeUMem

this model is auto created by ucloud code generater for open api,
you can also see https://docs.ucloud.cn for detail.
*/
type UMemDataSet struct {

	// 实例所在可用区，或者master redis所在可用区，参见 [可用区列表](../summary/regionlist.html)
	Zone string

	// 表示实例是主库还是从库,master,slave
	Role string

	// UMEM实例列表 UMemSlaveDataSet 如果没有slave，则没有该字段
	DataSet []UMemSlaveDataSet

	// 是否拥有只读Slave
	OwnSlave string

	// vpc
	VPCId string

	// 子网
	SubnetId string

	// 资源ID
	ResourceId string

	// 资源名称
	Name string

	// 创建时间
	CreateTime int

	// 到期时间
	ExpireTime int

	// 空间类型:single(无热备),double(热备)
	Type string

	// 协议类型: memcache, redis
	Protocol string

	// 容量单位GB
	Size int

	// 使用量单位MB
	UsedSize int

	// 实例状态                                  Starting                  // 创建中       Creating                  // 初始化中     CreateFail                // 创建失败     Fail                      // 创建失败     Deleting                  // 删除中       DeleteFail                // 删除失败     Running                   // 运行         Resizing                  // 容量调整中   ResizeFail                // 容量调整失败 Configing                 // 配置中       ConfigFail                // 配置失败Restarting                // 重启中SetPasswordFail    //设置密码失败
	State string

	// 计费模式，Year, Month, Dynamic, Trial
	ChargeType string

	// IP端口信息请，参见UMemSpaceAddressSet
	Address []UMemSpaceAddressSet

	// 业务组名称
	Tag string

	// distributed: 分布式版Redis,或者分布式Memcache；single：主备版Redis,或者单机Memcache；performance：高性能版
	ResourceType string

	// 节点的配置ID
	ConfigId string

	// 是否需要自动备份,enable,disable
	AutoBackup string

	// 自动备份开始时间,单位小时计,范围[0-23]
	BackupTime int

	// 是否开启高可用,enable,disable
	HighAvailability string

	// Redis版本信息
	Version string

	// 主备Redis，提供两种类型：同机房高可用Redis，和同地域跨机房高可用Redis
	URedisType string

	// 跨机房URedis，slave redis所在可用区，参见 [可用区列表](../summary/regionlist.html)
	SlaveZone string
}
