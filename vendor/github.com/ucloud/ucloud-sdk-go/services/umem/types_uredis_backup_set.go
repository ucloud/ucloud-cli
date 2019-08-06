package umem

/*
URedisBackupSet - DescribeURedisBackup

this model is auto created by ucloud code generater for open api,
you can also see https://docs.ucloud.cn for detail.
*/
type URedisBackupSet struct {

	// 备份ID
	BackupId string

	// 可用区，参见[可用区列表](../summary/regionlist.html)
	Zone string

	// 对应的实例ID
	GroupId string

	// 组名称
	GroupName string

	// 备份的名称
	BackupName string

	// 备份时间 (UNIX时间戳)
	BackupTime int

	// 备份文件大小, 以字节为单位
	BackupSize int

	// 备份类型: Manual 手动 Auto 自动
	BackupType string

	// 备份的状态: Backuping 备份中 Success 备份成功 Error 备份失败 Expired 备份过期
	State string
}
