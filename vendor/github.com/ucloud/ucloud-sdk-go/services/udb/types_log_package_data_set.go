package udb

/*
LogPackageDataSet - DescribeUDBLogPackage

this model is auto created by ucloud code generater for open api,
you can also see https://docs.ucloud.cn for detail.
*/
type LogPackageDataSet struct {

	// 备份id
	BackupId int

	// 备份名称
	BackupName string

	// 备份时间
	BackupTime int

	// 备份文件大小
	BackupSize int

	// 备份类型，包括2-binlog备份，3-slowlog备份
	BackupType int

	// 备份状态 Backuping // 备份中 Success // 备份成功 Failed // 备份失败 Expired // 备份过期
	State string

	// dbid
	DBId string

	// 对应的db名称
	DBName string

	// 所在可用区
	Zone string

	// 跨可用区高可用备库所在可用区
	BackupZone string
}
