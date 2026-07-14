package upfs

// VolumeRow represents a single row in the upfs volume list output.
type VolumeRow struct {
	ResourceID   string
	Name         string
	Group        string
	Size         string
	ProtocolType string
	MountAddress string
	ChargeType   string
	State        string
	CreationTime string
	Expiration   string
}
