package udb

/*
UDBInstanceBinlogSet - DescribeUDBInstanceBinlog

this model is auto created by ucloud code generater for open api,
you can also see https://docs.ucloud.cn for detail.
*/
type UDBInstanceBinlogSet struct {

	// Binlog文件名
	Name string

	// Binlog文件大小
	Size int

	// Binlog文件生成时间(时间戳)
	BeginTime int

	// Binlog文件结束时间(时间戳)
	EndTime int
}
