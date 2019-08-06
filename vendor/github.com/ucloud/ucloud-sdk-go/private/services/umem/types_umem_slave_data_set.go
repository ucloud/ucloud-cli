package umem

/*
UMemSlaveDataSet - DescribeUMem

this model is auto created by ucloud code generater for open api,
you can also see https://docs.ucloud.cn for detail.
*/
type UMemSlaveDataSet struct {

	// 实例所在可用区，或者master redis所在可用区，参见 [可用区列表](../summary/regionlist.html)
	Zone string

	// 子网
	SubnetId string

	// vpc
	VPCId string

	//
	VirtualIP string

	// 主实例id
	MasterGroupId string

	// 资源id
	GroupId string

	// 端口
	Port int

	// 实力大小
	MemorySize int

	// 资源名称
	GroupName string

	// 表示实例是主库还是从库,master,slave
	Role string

	// 修改时间
	ModifyTime int

	// 资源名称
	Name string

	// 创建时间
	CreateTime int

	// 到期时间
	ExpireTime int

	// 容量单位GB
	Size int

	// 使用量单位MB
	UsedSize int

	// 实例状态                                  Starting                  // 创建中       Creating                  // 初始化中     CreateFail                // 创建失败     Fail                      // 创建失败     Deleting                  // 删除中       DeleteFail                // 删除失败     Running                   // 运行         Resizing                  // 容量调整中   ResizeFail                // 容量调整失败 Configing                 // 配置中       ConfigFail                // 配置失败Restarting                // 重启中SetPasswordFail  //设置密码失败
	State string

	// 计费模式，Year, Month, Dynamic, Trial
	ChargeType string

	// 业务组名称
	Tag string

	// distributed: 分布式版Redis,或者分布式Memcache；single：主备版Redis,或者单机Memcache；performance：高性能版
	ResourceType string

	// 节点的配置ID
	ConfigId string

	// Redis版本信息
	Version string
}
