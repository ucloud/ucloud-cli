package uhost

/*
	UHostDiskSet - DescribeUHostInstance

	this model is auto created by ucloud code generater for open api,
	you can also see https://docs.ucloud.cn for detail.
*/
type UHostDiskSet struct {

	// 磁盘类型。系统盘: Boot，数据盘: Data,网络盘：Udisk
	Type string

	// 磁盘ID
	DiskId string

	// UDisk名字（仅当磁盘是UDisk时返回）
	Name string

	// 磁盘盘符
	Drive string

	// 磁盘大小，单位: GB
	Size int

	// 备份类型，DataArk
	BackupType string
}
