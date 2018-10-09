package uhost

// UHostPriceSet is a set to describe the price of uhost
type UHostPriceSet struct {
	// 付费方式, 预付费:Year 按年,Month 按月,Dynamic 按需;后付费:Postpay(按月)
	ChargeType string

	// 费用（元）
	Price int
}
