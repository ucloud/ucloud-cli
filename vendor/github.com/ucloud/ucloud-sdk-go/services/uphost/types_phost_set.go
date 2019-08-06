package uphost

/*
PHostSet - DescribePHost

this model is auto created by ucloud code generater for open api,
you can also see https://docs.ucloud.cn for detail.
*/
type PHostSet struct {

	// 可用区，参见 [可用区列表](../summary/regionlist.html)
	Zone string

	// PHost资源ID
	PHostId string

	// 物理机序列号
	SN string

	// 物理云主机状态。枚举值：\\ > 初始化:Initializing; \\ > 启动中：Starting； \\ > 运行中：Running；\\ > 关机中：Stopping； \\ > 安装失败：InstallFailed； \\ > 重启中：Rebooting；\\ > 关机：Stopped；
	PMStatus string

	// 物理机名称
	Name string

	// 物理机备注
	Remark string

	// 业务组
	Tag string

	// 镜像名称
	ImageName string

	// 操作系统名称
	OSname string

	// 创建时间
	CreateTime int

	// 到期时间
	ExpireTime int

	// 计费模式，枚举值为： Year，按年付费； Month，按月付费； Dynamic，按需付费（需开启权限）； Trial，试用（需开启权限）默认为月付
	ChargeType string

	// 电源状态，on 或 off
	PowerState string

	// 物理机类型，参见DescribePHostMachineType返回值
	PHostType string

	// 内存大小，单位：MB
	Memory int

	// CPU信息，见 PHostCPUSet
	CPUSet PHostCPUSet

	// 磁盘信息，见 PHostDiskSet
	DiskSet []PHostDiskSet

	// IP信息，见 PHostIPSet
	IPSet []PHostIPSet

	// 网络环境。枚举值：千兆：1G ，万兆：10G
	Cluster string

	// 自动续费
	AutoRenew string

	// 是否支持紧急登录
	IsSupportKVM string

	// 操作系统类型
	OSType string

	// 组件信息（暂不支持）
	Components string

	// 是否支持Raid。枚举值：Yes：支持；No：不支持。
	RaidSupported string
}
