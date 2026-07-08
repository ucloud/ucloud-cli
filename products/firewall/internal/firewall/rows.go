package firewall

// FirewallRow 表格行
type FirewallRow struct {
	ResourceID          string
	FirewallName        string
	Rule                string
	Group               string
	RuleAmount          int
	BoundResourceAmount int
	CreationTime        string
}

// FirewallResourceRow 表格行
type FirewallResourceRow struct {
	ResourceName string
	ResourceID   string
	ResourceType string
	IntranetIP   string
	Group        string
	Remark       string
}
