package uhost

/*
	UHostIPSet - DescribeUHostInstance

	this model is auto created by ucloud code generater for open api,
	you can also see https://docs.ucloud.cn for detail.
*/
type UHostIPSet struct {

	// 电信: China-telecom，联通: China-unicom， 国际: Internation，BGP: Bgp，内网: Private 双线: Duplet
	Type string

	// IP资源ID (内网IP无对应的资源ID)
	IPId string

	// IP地址
	IP string

	// IP对应的带宽, 单位: Mb (内网IP不显示带宽信息)
	Bandwidth int
}
