package uhost

/*
SnapshotSet - DescribeSnapshot

this model is auto created by ucloud code generater for open api,
you can also see https://docs.ucloud.cn for detail.
*/
type SnapshotSet struct {

	// 快照Id
	SnapshotId string

	// 磁盘Id。仅当为网络盘时返回此id。
	DiskId string

	// 主机Id。若udisk没有挂载，则不返回。
	UHostId string

	// 磁盘类型，枚举值为：LocalBoot,本地系统盘；LocalData,本地数据盘；UDiskBoot,云系统盘；UDiskData，云数据盘
	DiskType string

	// 大小
	Size int

	// 快照状态，枚举值为：Normal,可用；Creating,制作中；Failed,制作失败
	State string

	// 快照名称
	SnapshotName string

	// 快照描述
	SnapshotDescription string

	// 创建成功时间，unix时间
	CreateTime int

	// 指定的制作快照时间，unix时间
	SnapshotTime int

	// 资源名字。本地盘对应主机名字，网络盘对应udisk名字
	ResourceName string

	// 配置文件所在的可用区
	Zone string
}
