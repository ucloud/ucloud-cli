// Code is generated by ucloud-model, DO NOT EDIT IT.

package uphost

/*
PHostComponentSet - GetPHostTypeInfo
*/
type PHostComponentSet struct {

	// 组件数量
	Count int

	// 组件名称
	Name string
}

/*
PHostGpuInfoV2 - 裸金属Gpu信息V2版本
*/
type PHostGpuInfoV2 struct {

	// GPU数量
	Count int

	// GPU显存大小
	Memory string

	// GPU名称，例如：NVIDIA_V100S
	Name string

	// GPU性能指标
	Performance string
}

/*
PHostDiskSetV2 - 裸金属磁盘信息V2版本
*/
type PHostDiskSetV2 struct {

	// 磁盘类型
	DiskType int

	// IO性能
	IoCap int

	// 磁盘名
	Name string

	// 数量
	Number int

	// Raid级别
	RaidLevel int

	// 空间大小
	Space int

	// 转换单位
	UnitSize int
}

/*
PHostCPUSetV2 - 裸金属磁盘信息V2版本
*/
type PHostCPUSetV2 struct {

	// CPU核数
	CoreCount int

	// CPU个数
	Count int

	// CPU主频
	Frequency string

	// CPU型号
	Model string
}

/*
PHostCloudMachineTypeSetV2 - 裸金属云盘的MachineTypeSet V2版本
*/
type PHostCloudMachineTypeSetV2 struct {

	// CPU信息
	CPU PHostCPUSetV2

	// 集群名。枚举值：千兆网络集群：1G；万兆网络集群：10G；智能网卡网络：25G；
	Cluster string

	// 组件信息
	Components []PHostComponentSet

	// 磁盘信息
	Disks []PHostDiskSetV2

	// GPU信息
	GpuInfo PHostGpuInfoV2

	// 是否是裸金属机型
	IsBaremetal bool

	// 是否是GPU机型
	IsGpu bool

	// 是否需要加新机型标记
	IsNew bool

	// 内存大小，单位MB
	Memory int

	// 通常获取到的都是可售卖的
	OnSale bool

	// 参考价格。字典类型，default:为默认价格；cn-wlcb-01:乌兰察布A可用区价格
	Price string

	// 是否支持做Raid。枚举值：可以：Yes；不可以：No
	RaidSupported string

	// 适用场景。例如：ai表示AI学习场景；
	Scene []string

	// 库存数量
	Stock int

	// 库存状态。枚举值：有库存：Available；无库存：SoldOut
	StockStatus string

	// 物理云主机机型别名
	Type string

	// 机型所在可用区
	Zone string
}

/*
PHostCPUSet - DescribePHost
*/
type PHostCPUSet struct {

	// CPU核数
	CoreCount int

	// CPU个数
	Count int

	// CPU主频
	Frequence float64

	// CPU型号
	Model string
}

/*
PHostDescDiskSet - DescribePHost（包括传统和裸金属1、裸金属2）
*/
type PHostDescDiskSet struct {

	// 磁盘数量
	Count int

	// 裸金属机型参数：磁盘ID
	DiskId string

	// 裸金属机型参数：磁盘盘符
	Drive string

	// 磁盘IO性能，单位MB/s（待废弃）
	IOCap int

	// 裸金属机型参数：是否是启动盘。True/False
	IsBoot string

	// 磁盘名称，sys/data
	Name string

	// 单盘大小，单位GB
	Space int

	// 磁盘属性
	Type string
}

/*
PHostIPSet - DescribePHost
*/
type PHostIPSet struct {

	// IP对应带宽，单位Mb，内网IP不显示带宽信息
	Bandwidth int

	// IP地址，
	IPAddr string

	// IP资源ID(内网IP无资源ID)（待废弃）
	IPId string

	// MAC地址
	MACAddr string

	// 国际: Internation， BGP: BGP， 内网: Private
	OperatorName string

	// 子网ID
	SubnetId string

	// VPC ID
	VPCId string
}

/*
PHostSet - DescribePHost
*/
type PHostSet struct {

	// 自动续费
	AutoRenew string

	// 裸金属机型字段。枚举值：Normal=>正常、ImageMaking=>镜像制作中。
	BootDiskState string

	// CPU信息，见 PHostCPUSet
	CPUSet PHostCPUSet

	// 计费模式，枚举值为： Year，按年付费； Month，按月付费；默认为月付
	ChargeType string

	// 网络环境。枚举值：千兆：1G ，万兆：10G
	Cluster string

	// 组件信息（暂不支持）
	Components string

	// 创建时间
	CreateTime int

	// 磁盘信息，见 PHostDescDiskSet
	DiskSet []PHostDescDiskSet

	// 到期时间
	ExpireTime int

	// IP信息，见 PHostIPSet
	IPSet []PHostIPSet

	// 镜像名称
	ImageName string

	// 是否支持紧急登录
	IsSupportKVM string

	// 内存大小，单位：MB
	Memory int

	// 物理机名称
	Name string

	// 操作系统类型
	OSType string

	// 操作系统名称
	OSname string

	// PHost资源ID
	PHostId string

	// 物理机类型，参见DescribePHostMachineType返回值
	PHostType string

	// 物理云主机状态。枚举值：\\ > 初始化:Initializing; \\ > 启动中：Starting； \\ > 运行中：Running；\\ > 关机中：Stopping； \\ > 安装失败：InstallFailed； \\ > 重启中：Rebooting；\\ > 关机：Stopped； \\ > 迁移中(裸金属云盘)：Migrating
	PMStatus string

	// 物理云产品类型，枚举值：LocalDisk=>代表传统本地盘机型， CloudDisk=>云盘裸金属机型
	PhostClass string

	// 电源状态，on 或 off
	PowerState string

	// 是否支持Raid。枚举值：Yes：支持；No：不支持。
	RaidSupported string

	// RDMA集群id，仅云盘裸金属返回该值；其他类型物理云主机返回""。当物理机的此值与RSSD云盘的RdmaClusterId相同时，RSSD可以挂载到这台物理机。
	RdmaClusterId string

	// 物理机备注
	Remark string

	// 物理机序列号
	SN string

	// 业务组
	Tag string

	// 可用区，参见 [可用区列表](../summary/regionlist.html)
	Zone string
}

/*
PHostImageSet - DescribePHostImage
*/
type PHostImageSet struct {

	// 裸金属2.0参数。镜像创建时间。
	CreateTime string

	// 镜像描述
	ImageDescription string

	// 镜像ID
	ImageId string

	// 镜像名称
	ImageName string

	// 裸金属2.0参数。镜像大小。
	ImageSize int

	// 枚举值：Base=>基础镜像，Custom=>自制镜像。
	ImageType string

	// 操作系统名称
	OsName string

	// 操作系统类型
	OsType string

	// 裸金属2.0参数。镜像当前状态。
	State string

	// 支持的机型
	Support []string

	// 当前版本
	Version string
}

/*
PHostTagSet - DescribePHostTags
*/
type PHostTagSet struct {

	// 业务组名称
	Tag string

	// 该业务组中包含的主机个数
	TotalCount int
}

/*
PHostPriceSet - GetPHostPrice
*/
type PHostPriceSet struct {

	// Year/Month
	ChargeType string

	// 原价格, 单位:元, 保留小数点后两位有效数字
	OriginalPrice float64

	// 价格, 单位:元, 保留小数点后两位有效数字
	Price float64

	// 枚举值：phost=>为主机价格，如果是云盘包括了系统盘价格。cloudDisk=>所有数据盘价格，只是裸金属机型才返回此参数。
	Product string
}
