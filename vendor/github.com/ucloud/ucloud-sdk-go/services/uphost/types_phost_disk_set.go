package uphost

/*
PHostDiskSet - GetPHostTypeInfo

this model is auto created by ucloud code generater for open api,
you can also see https://docs.ucloud.cn for detail.
*/
type PHostDiskSet struct {

	// 单盘大小，单位GB
	Space int

	// 磁盘数量
	Count int

	// 磁盘属性
	Type string

	// 磁盘名称，sys/data
	Name string

	// 磁盘IO性能，单位MB/s（待废弃）
	IOCap int
}
