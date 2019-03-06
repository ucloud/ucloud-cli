package udb

/*
UDBInstancePriceSet - DescribeUDBInstancePrice

this model is auto created by ucloud code generater for open api,
you can also see https://docs.ucloud.cn for detail.
*/
type UDBInstancePriceSet struct {

	// Year， Month， Dynamic，Trial
	ChargeType string

	// 价格，单位为分，保留小数点后两位
	Price float64
}
