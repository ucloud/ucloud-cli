package uk8s

// clusterRow is the row struct for `ucloud uk8s list` output. Column order
// (field declaration order) drives the table layout; the platform's
// ctx.PrintList renders the row for any --output.
type clusterRow struct {
	ResourceID string
	Name       string
	K8sVersion string
	VPCID      string
	SubnetID   string
	MasterCnt  int
	NodeCnt    int
	Status     string
	Created    string
}

type nodeGroupRow struct {
	ResourceID  string
	Name        string
	MachineType string
	CPU         int
	MemoryMB    int
	NodeCount   int
	ChargeType  string
	ImageID     string
}

type nodeRow struct {
	ResourceID  string
	InstanceID  string
	Name        string
	Role        string
	Zone        string
	MachineType string
	CPU         int
	MemoryMB    int
	Status      string
	OS          string
}

type imageRow struct {
	ResourceID    string
	Name          string
	ZoneID        int
	ProductType   string
	NotSupportGPU bool
}

type versionRow struct {
	K8sVersion        string
	ContainerdVersion string
}
