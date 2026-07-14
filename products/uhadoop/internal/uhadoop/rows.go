package uhadoop

// listRow is the full row for ListUHadoopInstance output.
type listRow struct {
	InstanceId     string
	InstanceName   string
	Framework      string
	ReleaseVersion string
	HadoopVersion  string
	State          string
	Zone           string
	VPCId          string
	SubnetId       string
	ChargeType     string
	CreateTime     string
	ExpireTime     string
}

// listRowDefault is the default (non-wide) column set for list.
type listRowDefault struct {
	InstanceId     string
	InstanceName   string
	Framework      string
	ReleaseVersion string
	HadoopVersion  string
	State          string
	Zone           string
	CreateTime     string
	ExpireTime     string
}

// instanceTypeRow is the row for GetUHadoopNodeType output.
type instanceTypeRow struct {
	NodeType         string
	HostType         string
	CPU              string
	Memory           string
	CPUToMemoryRatio string
	SuitableRole     string
	IsUsable         string
	GpuType          string
	GpuCount         int
	DiskType         string
	DiskMinSize      string
	DiskMaxSize      string
	DiskMinNum       string
	DiskMaxNum       string
}

// frameworkAppRow is the row for ListUHadoopFrameworkAppByUseCase output.
type frameworkAppRow struct {
	Framework        string
	FrameworkVersion string
	ReleaseVersion   string
	HadoopVersion    string
	UseCase          string
	Apps             string
	MustHas          string
}
