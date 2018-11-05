package uhost

/*
	UHostDisk - the request query for disk of uhost
*/
type UHostDisk struct {
	// 磁盘大小，单位GB。
	Size *int `required:"true"`

	// 磁盘类型。枚举值：LOCAL_NORMAL 普通本地盘 | CLOUD_NORMAL 普通云盘 |LOCAL_SSD SSD本地盘 | CLOUD_SSD SSD云盘，默认为LOCAL_NORMAL。磁盘仅支持有限组合，详情请查询控制台。
	Type *string `required:"true"`

	// 是否是系统盘。枚举值：True|False
	IsBoot *string `required:"true"`

	// NONE|DATAARK
	BackupType *string `required:"false"`
}
