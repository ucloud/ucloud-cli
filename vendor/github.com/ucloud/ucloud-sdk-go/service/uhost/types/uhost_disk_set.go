package types

// UHostDiskSet - DescribeUHostInstance
type UHostDiskSet struct {

	// 磁盘类型。系统盘: Boot，数据盘: Data,网络盘：Udisk
	Type string

	// 磁盘长ID
	DiskId string

	// UDisk名字（仅当磁盘是UDisk时返回）
	Name int

	// 磁盘盘符
	Drive string

	// 磁盘大小，单位: GB
	Size int

	// 备份方案，枚举类型：BASIC_SNAPSHOT,普通快照；DATAARK,方舟。无快照则不返回该字段。
	BackupType string

	// 当前主机的IOPS值
	IOPS int

	// 磁盘短ID
	DiskShortId string

	// Yes: 加密  No: 非加密
	Encrypted string
}
