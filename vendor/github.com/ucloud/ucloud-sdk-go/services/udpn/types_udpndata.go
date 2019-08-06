package udpn

/*
UDPNData - UDPN 详细信息

this model is auto created by ucloud code generater for open api,
you can also see https://docs.ucloud.cn for detail.
*/
type UDPNData struct {

	// UDPN 资源短 ID
	UDPNId string

	// 可用区域 1
	Peer1 string

	// 可用区域 2
	Peer2 string

	// 计费类型
	ChargeType string

	// 带宽
	Bandwidth int

	// unix 时间戳 创建时间
	CreateTime int

	// unix 时间戳 到期时间
	ExpireTime int
}
