package ulb

/*
	ULBSet - DescribeULB

	this model is auto created by ucloud code generater for open api,
	you can also see https://docs.ucloud.cn for detail.
*/
type ULBSet struct {

	// 负载均衡的资源ID
	ULBId string

	// 负载均衡的资源名称(内部记载，废弃)
	ULBName string

	// 负载均衡的资源名称（资源系统中），缺省值“ULB”
	Name string

	// 负载均衡的业务组名称，缺省值“Default”
	Tag string

	// 负载均衡的备注，缺省值“”
	Remark string

	// 带宽类型，枚举值为： 0，非共享带宽； 1，共享带宽
	BandwidthType int

	// 带宽
	Bandwidth int

	// ULB的创建时间，格式为Unix Timestamp
	CreateTime int

	// ULB的到期时间，格式为Unix Timestamp
	ExpireTime int

	// ULB的详细信息列表（废弃）
	Resource []string

	// ULB的详细信息列表，具体结构见下方 ULBIPSet
	IPSet []ULBIPSet

	// 负载均衡实例中存在的VServer实例列表，具体结构见下方 ULBVServerSet
	VServerSet []ULBVServerSet

	// ULB 的类型
	ULBType string

	// ULB所在的VPC的ID
	VPCId string

	// ULB 为 InnerMode 时，ULB 所属的子网ID，默认为空
	SubnetId string

	// ULB 所属的业务组ID
	BusinessId string

	// ULB的内网IP，当ULBType为OuterMode时，该值为空
	PrivateIP string
}
