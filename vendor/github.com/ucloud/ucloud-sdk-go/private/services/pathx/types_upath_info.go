package pathx

/*
UPathInfo - 加速线路信息

this model is auto created by ucloud code generater for open api,
you can also see https://docs.ucloud.cn for detail.
*/
type UPathInfo struct {

	// 支付方式
	ChargeType string

	// UPath名字
	Name string

	// UPath ID 号
	UPathId string

	// 带宽
	Bandwidth int

	// 线路ID
	LineId string

	// 与该UPath绑定的UGA列表
	UGAList []UGAAInfo

	// UPath创建的时间
	CreateTime int

	// UPath的过期时间
	ExpireTime int

	// 线路入口名称
	LineFromName string

	// 线路出口名称
	LineToName string

	// 线路出口IP信息
	OutPublicIpList []OutPublicIpInfo
}
