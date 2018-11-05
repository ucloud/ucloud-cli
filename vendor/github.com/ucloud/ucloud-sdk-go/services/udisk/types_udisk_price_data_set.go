package udisk

/*
	UDiskPriceDataSet - DescribeUDiskPrice

	this model is auto created by ucloud code generater for open api,
	you can also see https://docs.ucloud.cn for detail.
*/
type UDiskPriceDataSet struct {

	// Year， Month， Dynamic，Trial
	ChargeType string

	// 价格 (单位: 分)
	Price float64

	// "UDataArk","UDisk"
	ChargeName string
}
