package uphost

/*
PHostIPSet - DescribePHost

this model is auto created by ucloud code generater for open api,
you can also see https://docs.ucloud.cn for detail.
*/
type PHostIPSet struct {

	//  国际: Internation， BGP: BGP， 内网: Private
	OperatorName string

	// IP资源ID(内网IP无资源ID)（待废弃）
	IPId string

	// IP地址，
	IPAddr string

	// MAC地址
	MACAddr string

	// IP对应带宽，单位Mb，内网IP不显示带宽信息
	Bandwidth int

	// 子网ID
	SubnetId string

	// VPC ID
	VPCId string
}
