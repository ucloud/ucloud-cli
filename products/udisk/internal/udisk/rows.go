package udisk

// DiskRow TableRow
type DiskRow struct {
	ResourceID     string
	Name           string
	Group          string
	Size           string
	Type           string
	MountUHost     string
	MountPoint     string
	EnableDataArk  string
	State          string
	CreationTime   string
	ExpirationTime string
}

// SnapshotRow 表格行
type SnapshotRow struct {
	Name             string
	ResourceID       string
	AvailabilityZone string
	BoundUDisk       string
	Size             string
	State            string
	UDiskType        string
	CreationTime     string
}
