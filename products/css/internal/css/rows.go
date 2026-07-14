package css

// InstanceRow represents a UES instance in list output
type InstanceRow struct {
	InstanceID   string
	InstanceName string
	AppName      string
	AppVersion   string
	Zone         string
	State        string
	NodeCount    string
	VPCId        string
	SubnetId     string
	ChargeType   string
	CreateTime   string
	ExpireTime   string
}

// DiskLimitRow represents a disk size limitation in disk-limit output
type DiskLimitRow struct {
	DiskType  string
	MinSizeGB string
	MaxSizeGB string
}

// NodeConfRow represents a node configuration in node-conf output
type NodeConfRow struct {
	NodeConf   string
	CPU        string
	MemoryGB   string
	DiskSizeGB string
	DiskType   string
	SecGroup   string
}

// AppVersionRow represents an application version in app-version output
type AppVersionRow struct {
	AppName     string
	AppVersion  string
	IsMultiZone string
}
