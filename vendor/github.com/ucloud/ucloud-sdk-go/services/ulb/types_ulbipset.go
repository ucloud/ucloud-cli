package ulb

/*
	ULBIPSet - DescribeULB

	this model is auto created by ucloud code generater for open api,
	you can also see https://docs.ucloud.cn for detail.
*/
type ULBIPSet struct {

	// 弹性IP的运营商信息，枚举值为：  Bgp：BGP IP International：国际IP
	OperatorName string

	// 弹性IP地址
	EIP string

	// 弹性IP的ID
	EIPId string
}
