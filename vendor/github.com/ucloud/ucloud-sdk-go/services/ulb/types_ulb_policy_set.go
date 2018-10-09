package ulb

/*
	UlbPolicySet - DescribePolicyGroup

	this model is auto created by ucloud code generater for open api,
	you can also see https://docs.ucloud.cn for detail.
*/
type UlbPolicySet struct {

	// 内容转发策略组ID
	PolicyId string

	// 内容转发匹配字段的类型，当前只支持按域名转发。枚举值为： Domain，按域名转发
	Type string

	// 内容转发匹配字段
	Match string

	// 内容转发策略组ID应用的VServer实例的ID
	VServerId string

	// 内容转发策略组ID所应用的后端资源列表，具体结构见 UlbPolicyBackendSet
	BackendSet []UlbPolicyBackendSet
}
