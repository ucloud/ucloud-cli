package types

// SSLBindedTargetSet - DescribeSSL
type SSLBindedTargetSet struct {

	// SSL证书绑定到的VServer的资源ID
	VServerId string

	// 对应的VServer的名字
	VServerName string

	// VServer 所属的ULB实例的资源ID
	ULBId string

	// ULB实例的名称
	ULBName string
}
