package pathx

/*
UGAAInfo - 全球加速实例信息

this model is auto created by ucloud code generater for open api,
you can also see https://docs.ucloud.cn for detail.
*/
type UGAAInfo struct {

	// 全球加速ID
	UGAId string

	// 流量转发方式，包括L4、L7
	ForwardType string

	// 全球加速cname
	CName string

	// 加速源IP列表
	IPList []string

	// 加速实例名称
	UGAName string

	// 加速源域名
	Domain string

	// 加速地区
	Location string

	// 绑定的加速线路
	UPathSet []UPathSet

	// 端口配置信息
	TaskSet []UGAATask

	// 线路出口IP地址
	OutPublicIpList []OutPublicIpInfo
}
