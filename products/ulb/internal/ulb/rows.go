package ulb

type Row struct {
	Name         string
	ResourceID   string
	Group        string
	Network      string
	VserverCount int
	VPC          string
	CreationTime string
}

type VServerRow struct {
	VServerName         string
	ResourceID          string
	ListenType          string
	Protocol            string
	Port                int
	LBMethod            string
	SessionMaintainMode string
	SessionMaintainKey  string
	ClientTimeout       string
	HealthCheckMode     string
	HealthCheckDomain   string
	HealthCheckPath     string
}

type BackendRow struct {
	Name        string
	ResourceID  string
	BackendID   string
	PrivateIP   string
	Port        int
	HealthCheck string
	NodeMode    string
	Weight      int
}

type PolicyRow struct {
	ForwardMethod string
	Expression    string
	PolicyID      string
	PolicyType    string
	Backends      string
}

type SSLCertificate struct {
	Name         string
	ResourceID   string
	MD5          string
	BindResource string
	UploadTime   string
}
