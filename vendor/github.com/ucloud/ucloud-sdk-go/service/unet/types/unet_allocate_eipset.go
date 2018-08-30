package types

// UnetAllocateEIPSet - AllocateEIP
type UnetAllocateEIPSet struct {

	// 申请到的EIP资源ID
	EIPId string

	// 申请到的IPv4地址. 如果在请求参数中OperatorName为Duplet, 则EIPAddr中会含有两个IP地址, 一个为电信IP, 一个为联通IP. 其余情况下, EIPAddr只含有一个IP. 参见 UnetEIPAddrSet
	EIPAddr []UnetEIPAddrSet
}
