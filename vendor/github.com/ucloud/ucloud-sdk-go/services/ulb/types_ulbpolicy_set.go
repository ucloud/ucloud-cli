package ulb

/*
	ULBPolicySet - 内容转发详细列表

	this model is auto created by ucloud code generater for open api,
	you can also see https://docs.ucloud.cn for detail.
*/
type ULBPolicySet struct {

	// 内容转发Id，默认内容转发类型下为空。
	PolicyId string

	// 内容类型，枚举值：Custom -> 客户自定义；Default -> 默认内容转发
	PolicyType string

	// 内容转发匹配字段的类型，枚举值：Domain -> 域名；Path -> 路径； 默认内容转发类型下为空
	Type string

	// 内容转发匹配字段;默认内容转发类型下为空。
	Match string

	// 内容转发优先级，范围[1,9999]，数字越大优先级越高。默认内容转发规则下为0。
	PolicyPriority int

	// 所属VServerId
	VServerId string

	// 默认内容转发类型下返回当前rs总数
	TotalCount int

	// 内容转发下rs的详细信息，参考PolicyBackendSet
	BackendSet []PolicyBackendSet
}
