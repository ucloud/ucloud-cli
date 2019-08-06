package pathx

/*
UPathSet - uga信息中携带的upath信息

this model is auto created by ucloud code generater for open api,
you can also see https://docs.ucloud.cn for detail.
*/
type UPathSet struct {

	// UPath名字
	UPathName string

	// UPath ID号
	UPathId string

	// 带宽
	Bandwidth int

	// 线路ID
	LineId string

	// 线路起点名字
	LineFromName string

	// 线路对端名字
	LineToName string

	// 线路对端IP
	OutPublicIpList []OutPublicIpInfo
}
