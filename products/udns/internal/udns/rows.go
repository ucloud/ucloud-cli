package udns

type zoneRow struct {
	ZoneID     string
	Name       string
	ChargeType string
	Recursion  string
	VPCs       string
	Tag        string
	Remark     string
	CreateTime string
	ExpireTime string
}

type recordRow struct {
	RecordID  string
	Name      string
	Type      string
	TTL       string
	Values    string
	ValueType string
	Remark    string
}
