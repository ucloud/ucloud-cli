package uk8s

// clusterRow is the row struct for `ucloud uk8s list` output. Column order
// (field declaration order) drives the table layout; the platform's
// ctx.PrintList renders the row for any --output.
type clusterRow struct {
	ResourceID        string
	Name              string
	ApiServer         string
	ClusterLogInfo    string
	ExternalApiServer string
	K8sVersion        string
	VPCID             string
	SubnetID          string
	MasterCnt         int
	NodeCnt           int
	PodCIDR           string
	ServiceCIDR       string
	Status            string
	Created           string
}

type nodeGroupRow struct {
	ResourceID         string
	Name               string
	MachineType        string
	CPU                int
	MemoryMB           int
	NodeCount          int
	NodeIDs            string
	ChargeType         string
	ImageID            string
	BootDiskType       string
	BootDiskSize       int
	DataDiskType       string
	DataDiskSize       int
	GPU                int
	GPUType            string
	MinimalCPUPlatform string
	Tag                string
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
