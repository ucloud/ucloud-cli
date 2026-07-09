package subnet

type Row struct {
	SubnetName     string
	ResourceID     string
	Group          string
	AffiliatedVPC  string
	NetworkSegment string
	CreationTime   string
}

type ResourceRow struct {
	ResourceName string
	ResourceID   string
	ResourceType string
	PrivateIP    string
}
