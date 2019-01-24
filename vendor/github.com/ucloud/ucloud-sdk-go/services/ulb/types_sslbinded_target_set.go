package ulb

/*
SSLBindedTargetSet - DescribeSSL

this model is auto created by ucloud code generater for open api,
you can also see https://docs.ucloud.cn for detail.
*/
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
