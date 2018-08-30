package types

// UnetEIPAddrSet - DescribeEIP
type UnetEIPAddrSet struct {

	// 运营商信息如: 电信: Telecom, 联通: Unicom, 国际: International, Duplet: 双线IP（电信+联通), BGP: Bgp
	OperatorName string

	// IP地址
	IP string
}
