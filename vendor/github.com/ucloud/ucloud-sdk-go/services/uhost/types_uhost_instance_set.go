package uhost

/*
	UHostInstanceSet - DescribeUHostInstance

	this model is auto created by ucloud code generater for open api,
	you can also see https://docs.ucloud.cn for detail.
*/
type UHostInstanceSet struct {

	// 可用区。参见 [可用区列表](../summary/regionlist.html)
	Zone string

	// UHost实例ID
	UHostId string

	// UHost类型，枚举为：N1：标准型系列1；N2：标准型系列2 ；I1：高IO型系列1；I2：高IO型系列2；D1：大数据型系列1；G1：GPU型系列1；G2：GPU型系列2；G3：GPU型系列2
	UHostType string

	// 系统盘与数据盘的磁盘类型。 枚举值为：LocalDisk，本地磁盘; UDisk，云硬盘
	StorageType string

	// 镜像ID
	ImageId string

	// 基础镜像ID（指当前自定义镜像的来源镜像）
	BasicImageId string

	// 基础镜像名称（指当前自定义镜像的来源镜像）
	BasicImageName string

	// 业务组名称
	Tag string

	// 备注
	Remark string

	// UHost实例名称
	Name string

	// 实例状态， 初始化: Initializing; 启动中: Starting; 运行中: Running; 关机中: Stopping; 关机: Stopped 安装失败: Install Fail; 重启中: Rebooting
	State string

	// 创建时间，格式为Unix时间戳
	CreateTime int

	// 计费模式，枚举值为： Year，按年付费； Month，按月付费； Dynamic，按需付费（需开启权限）；
	ChargeType string

	// 到期时间，格式为Unix时间戳
	ExpireTime int

	// 虚拟CPU核数，单位: 个
	CPU int

	// 内存大小，单位: MB
	Memory int

	// 是否自动续费，自动续费：“Yes”，不自动续费：“No”
	AutoRenew string

	// 磁盘信息见 UHostDiskSet
	DiskSet []UHostDiskSet

	// 详细信息见 UHostIPSet
	IPSet []UHostIPSet

	// 网络增强。目前仅支持Normal和Super
	NetCapability string

	// 网络状态 连接：Connected， 断开：NotConnected
	NetworkState string

	// yes: 开启方舟； no，未开启方舟
	TimemachineFeature string

	// true: 开启热升级； false，未开启热升级
	HotplugFeature bool

	// 基础网络：Default；子网：Private
	SubnetType string

	// 内网或者子网的IP地址
	IPs []string

	// Os名称
	OsName string

	// "Linux"或者"Windows"
	OsType string

	// 删除时间，格式为Unix时间戳
	DeleteTime int

	// 主机系列：N2，表示系列2；N1，表示系列1
	HostType string

	// 主机的生命周期类型。目前仅支持Normal：普通；
	LifeCycle string

	// 主机的 GPU 数量
	GPU int

	// 系统盘状态 Normal: 已初始化完成
	BootDiskState string

	// 主机的存储空间
	TotalDiskSpace int
}
