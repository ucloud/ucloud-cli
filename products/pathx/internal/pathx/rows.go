package pathx

type UGA3PriceRow struct {
	AccelerationArea           string
	AccelerationAreaName       string
	AccelerationForwarderPrice string
	AccelerationBandwidthPrice string
}

type Uga3DescribeRow struct {
	ResourceID           string
	CName                string
	Name                 string
	AccelerationArea     string
	AccelerationAreaName string
	EgressIpList         string
	Bandwidth            int
	Remark               string
	OriginArea           string
	OriginAreaCode       string
	CreateTime           string
	ExpireTime           string
	ChargeType           string
	IPList               string
	Domain               string
}

type Uga3PortRow struct {
	Protocol string
	RSPort   int
	Port     int
}

type PathxUpdatePriceRow struct {
	InstanceId  string
	Bandwidth   int
	UpdatePrice float64
}

type PathxOptimizationRow struct {
	AccelerationName string
	AccelerationArea string
	Area             string
	AreaCode         string
	CountryCode      string
	FlagUnicode      string
	FlagEmoji        string
	Latency          string
	LatencyWAN       string
	LatencyPathX     string
	Loss             string
	LossWAN          string
	LossPathx        string
}

type PathxOptionalAreaRow struct {
	AreaCode      string
	Area          string
	CountryCode   string
	FlagUnicode   string
	FlagEmoji     string
	ContinentCode string
}

type EgressIpInfoRow struct {
	IP   string
	Area string
}

type upathRow struct {
	ResourceID      string
	UPathName       string
	AcceleratedPath string
	BoundUGA        string
}

type UGARow struct {
	ResourceID      string
	UGAName         string
	CName           string
	Origin          string
	AcceleratedPath string
}

type describeRow struct {
	Attribute string
	Content   string
}
