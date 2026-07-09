package bw

type SharedBWRow struct {
	Name           string
	ResourceID     string
	ChargeType     string
	Bandwidth      string
	EIP            string
	ExpirationTime string
}

type BandwidthPkgRow struct {
	ResourceID string
	EIP        string
	Bandwidth  string
	StartTime  string
	EndTime    string
}
