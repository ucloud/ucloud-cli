package udisk

/*
	UDiskDataSet - DescribeUDisk

	this model is auto created by ucloud code generater for open api,
	you can also see https://docs.ucloud.cn for detail.
*/
type UDiskDataSet struct {

	// UDisk实例Id
	UDiskId string

	// 实例名称
	Name string

	// 容量单位GB
	Size int

	// 状态:Available(可用),Attaching(挂载中), InUse(已挂载), Detaching(卸载中), Initializating(分配中), Failed(创建失败),Cloning(克隆中),Restoring(恢复中),RestoreFailed(恢复失败),
	Status string

	// 创建时间
	CreateTime int

	// 过期时间
	ExpiredTime int

	// 挂载的UHost的Id
	UHostId string

	// 挂载的UHost的Name
	UHostName string

	// 挂载的UHost的IP
	UHostIP string

	// 挂载的设备名称
	DeviceName string

	// Year,Month,Dynamic,Trial
	ChargeType string

	// 业务组名称
	Tag string

	// 资源是否过期，过期:"Yes", 未过期:"No"
	IsExpire string

	// 是否支持数据方舟，支持:"2.0", 不支持:"1.0"
	Version string

	// 是否开启数据方舟，开启:"Yes", 不支持:"No"
	UDataArkMode string

	// 该盘快照个数
	SnapshotCount int

	// 该盘快照上限
	SnapshotLimit int

	// 云硬盘类型: 普通数据盘:DataDisk,普通系统盘:SystemDisk,SSD数据盘:SSDDataDisk
	DiskType string
}
