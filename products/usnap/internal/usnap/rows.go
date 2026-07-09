package usnap

// SnapshotServiceRow represents a single row in the usnap snapshot service list output.
type SnapshotServiceRow struct {
	ResourceID   string
	VDiskID      string
	VDiskName    string
	VDiskSize    string
	VDiskType    string
	Group        string
	ChargeType   string
	AutoRenew    string
	Status       string
	Zone         string
	CreationTime string
	Expiration   string
}
