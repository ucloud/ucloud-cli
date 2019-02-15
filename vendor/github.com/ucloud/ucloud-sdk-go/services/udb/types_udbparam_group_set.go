package udb

/*
UDBParamGroupSet - DescribeUDBParamGroup

this model is auto created by ucloud code generater for open api,
you can also see https://docs.ucloud.cn for detail.
*/
type UDBParamGroupSet struct {

	// 参数组id
	GroupId int

	// 参数组名称
	GroupName string

	// DB类型id，mysql/mongodb按版本细分各有一个id 目前id的取值范围为[1,7],数值对应的版本如下 1：mysql-5.5，2：mysql-5.1，3：percona-5.5 4：mongodb-2.4，5：mongodb-2.6，6：mysql-5.6 7：percona-5.6
	DBTypeId string

	// 参数组描述
	Description string

	// 参数组是否可修改
	Modifiable bool

	// 参数的键值对表 UDBParamMemberSet
	ParamMember []UDBParamMemberSet
}
