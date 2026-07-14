package clickhouse

// ClusterRow represents a UClickhouse cluster in list output.
type ClusterRow struct {
	ClusterID               string
	ClusterName             string
	Status                  string
	ClickhouseVersion       string
	ShardCount              string
	ReplicateCount          string
	VPCId                   string
	SubnetId                string
	ClickhouseMachineTypeID string
	ClickhouseDataDiskType  string
	ClickhouseDataDiskSize  string
	CreateTime              string
	ExpireTime              string
}

// CreateOptionRow represents an available creation option.
type CreateOptionRow struct {
	OptionType      string
	Version         string
	VersionName     string
	NodeType        string
	MachineTypeID   string
	MachineTypeName string
	MachineType     string
	CPU             string
	MemoryGB        string
	NodeCounts      string
	IsSecGroup      string
	DiskType        string
	MinSizeGB       string
	MaxSizeGB       string
	DefaultSizeGB   string
	StepGB          string
	MaxNodeCount    string
}
