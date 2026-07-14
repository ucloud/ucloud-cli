package gssh

type GSSHRow struct {
	ResourceID         string
	SSHServerIP        string
	AcceleratingDomain string
	SSHServerLocation  string
	SSHPort            int
	GlobalSSHPort      int
	Remark             string
	InstanceType       string
}

type GsshLocation struct {
	AirportCode       string
	SSHServerLocation string
	CoveredArea       string
}
