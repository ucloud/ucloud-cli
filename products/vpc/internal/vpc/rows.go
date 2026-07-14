package vpc

type Row struct {
	VPCName        string
	ResourceID     string
	Group          string
	NetworkSegment string
	SubnetCount    int
	CreationTime   string
}

type IntercomRow struct {
	VPCName    string
	ResourceID string
	Segments   string
	ProjectID  string
	DstRegion  string
	Group      string
}
