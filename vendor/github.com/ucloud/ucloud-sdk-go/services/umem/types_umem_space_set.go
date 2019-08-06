package umem

/*
UMemSpaceSet - DescribeUMemSpace

this model is auto created by ucloud code generater for open api,
you can also see https://docs.ucloud.cn for detail.
*/
type UMemSpaceSet struct {

	// 内存空间ID
	SpaceId string

	// 内存空间名称
	Name string

	// 可用区，参见[可用区列表](../summary/regionlist.html)
	Zone string

	// 创建时间
	CreateTime int

	// 到期时间
	ExpireTime int

	// 空间类型:single(无热备),double(热备)
	Type string

	// 协议类型: memcache, redis
	Protocol string

	// 容量单位GB
	Size int

	// 使用量单位MB
	UsedSize int

	// Starting:创建中 Running:运行中 Fail:失败
	State string

	// Year, Month, Dynamic, Trial
	ChargeType string

	// IP端口信息请参见 UMemSpaceAddressSet
	Address []UMemSpaceAddressSet

	// VPC ID
	VPCId string

	// Subnet ID
	SubnetId string

	// 业务组
	Tag string
}
