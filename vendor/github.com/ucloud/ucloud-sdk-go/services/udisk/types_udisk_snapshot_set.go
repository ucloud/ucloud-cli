package udisk

/*
UDiskSnapshotSet - DescribeUDiskSnapshot

this model is auto created by ucloud code generater for open api,
you can also see https://docs.ucloud.cn for detail.
*/
type UDiskSnapshotSet struct {

	// 快照Id
	SnapshotId string

	// 快照名称
	Name string

	// 快照的源UDisk的Id
	UDiskId string

	// 快照的源UDisk的Name
	UDiskName string

	// 创建时间
	CreateTime int

	// 过期时间
	ExpiredTime int

	// 容量单位GB
	Size int

	// 快照描述
	Comment string

	// 快照状态，Normal:正常,Failed:失败,Creating:制作中
	Status string

	// 对应磁盘是否处于可用状态
	IsUDiskAvailable bool

	// 快照版本
	Version string

	// 对应磁盘制作快照时所挂载的主机
	UHostId string

	// 磁盘类型，0:数据盘，1:系统盘
	DiskType int
}
