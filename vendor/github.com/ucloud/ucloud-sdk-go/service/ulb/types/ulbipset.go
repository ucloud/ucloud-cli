package types

// ULBIPSet - DescribeULB
type ULBIPSet struct {

	// 弹性IP的运营商信息，枚举值为： Telecom：电信IP Unicom：联通IP Duplet：双线IP（电信+联通） Bgp：BGP IP International：国际IP
	OperatorName string

	// 弹性IP地址
	EIP string

	// 弹性IP的ID
	EIPId string
}
