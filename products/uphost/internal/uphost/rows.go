package uphost

// uphostRow 表格行
type uphostRow struct {
	ResourceID string
	Name       string
	PrivateIP  string
	PublicIP   string
	Config     string
	Image      string
	HostType   string
	Status     string
	Group      string
}
