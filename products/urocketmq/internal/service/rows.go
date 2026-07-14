package service

type serviceRow struct {
	ServiceId       string
	Name            string
	State           string
	Tps             int
	Storage         int
	TopicLimit      int
	Address         string
	AddressExtranet string
	VpcId           string
	SubnetId        string
	ChargeType      string
	CreateTime      int
	ExpireTime      int
	Remark          string
	Tag             string
	Edition         string
	Mode            string
	AutoRenew       string
	IsExpire        string
	Quantity        int
	Region          string
}

type serviceRowDefault struct {
	Name       string
	ServiceId  string
	State      string
	Config     string
	Address    string
	CreateTime string
	ExpireTime string
}

type serviceRowAllRegion struct {
	Name       string
	ServiceId  string
	State      string
	Config     string
	Address    string
	CreateTime string
	ExpireTime string
	Region     string
}

type priceRowDefault struct {
	ChargeName string
	ChargeType string
	Price      float64
}
