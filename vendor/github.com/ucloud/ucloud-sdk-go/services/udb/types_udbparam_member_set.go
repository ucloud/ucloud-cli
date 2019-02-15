package udb

/*
UDBParamMemberSet - DescribeUDBParamGroup

this model is auto created by ucloud code generater for open api,
you can also see https://docs.ucloud.cn for detail.
*/
type UDBParamMemberSet struct {

	// 参数名称
	Key string

	// 参数值
	Value string

	// 参数值应用类型，取值范围为{0,10,20,30},各值 代表意义为 0-unknown、10-int、20-string、 30-bool
	ValueType int

	// 允许的值(根据参数类型，用分隔符表示)
	AllowedVal string

	// 参数值应用类型,取值范围为{0,10,20}，各值代表 意义为0-unknown、10-static、20-dynamic
	ApplyType int

	// 是否可更改，默认为false
	Modifiable bool

	// 允许值的格式类型，取值范围为{0,10,20}，意义分 别为PVFT_UNKOWN=0,PVFT_RANGE=10, PVFT_ENUM=20
	FormatType int
}
