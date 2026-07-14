package udns

// ZoneRow is one table row for `udns list`.
type ZoneRow struct {
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

// RecordRow is one table row for `udns record list`.
type RecordRow struct {
	RecordID  string
	Name      string
	Type      string
	TTL       string
	Values    string
	ValueType string
	Remark    string
}
