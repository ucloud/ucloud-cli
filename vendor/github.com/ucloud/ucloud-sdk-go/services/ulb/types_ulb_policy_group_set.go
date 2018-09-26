package ulb

/*
	UlbPolicyGroupSet - DescribePolicyGroup

	this model is auto created by ucloud code generater for open api,
	you can also see https://docs.ucloud.cn for detail.
*/
type UlbPolicyGroupSet struct {

	// 内容转发策略组ID
	GroupId string

	// 内容转发策略组名称
	GroupName string

	// 内容转发策略组详细信息，具体结构见 UlbPolicySet
	PolicySet []UlbPolicySet
}
