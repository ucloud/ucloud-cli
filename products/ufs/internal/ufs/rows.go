package ufs

// VolumeRow represents a single row in the ufs volume list output.
type VolumeRow struct {
	ResourceID   string
	Name         string
	Group        string
	Size         string
	UsedSize     string
	ProtocolType string
	StorageType  string
	MountPoints  string
	State        string
	CreationTime string
	Expiration   string
}
