package uhost

// rows.go holds the table-row structs for uhost list output. The platform
// printer (ctx.PrintList) derives table columns from a struct's exported fields
// in declaration order, so the original cmd/uhost.go listUhost column selection
// (which passed an explicit []string column list to base.PrintTable per output
// mode) is reproduced here as three per-mode structs with byte-identical field
// names+order. JSON output uses the full uhostRow (matching the original, which
// always marshalled the full UHostRow slice in --json mode).

// uhostRow is the full row (wide mode + json). Field set+order is byte-identical
// to cmd/uhost.go UHostRow.
type uhostRow struct {
	UHostName    string
	Remark       string
	ResourceID   string
	Group        string
	PrivateIP    string
	PublicIP     string
	Config       string
	DiskSet      string
	Zone         string
	Image        string
	VPC          string
	Subnet       string
	Type         string
	State        string
	CreationTime string
}

// uhostRowDefault is the default (non-wide, non-all-region) column set:
// UHostName, ResourceID, Group, PrivateIP, PublicIP, Config, Image, Type, State,
// CreationTime — matching cmd/uhost.go listUhost's default cols verbatim.
type uhostRowDefault struct {
	UHostName    string
	ResourceID   string
	Group        string
	PrivateIP    string
	PublicIP     string
	Config       string
	Image        string
	Type         string
	State        string
	CreationTime string
}

// uhostRowAllRegion is the default column set plus a trailing Zone column,
// matching cmd/uhost.go listUhost when listAllRegion is true (cols =
// default cols + "Zone").
type uhostRowAllRegion struct {
	UHostName    string
	ResourceID   string
	Group        string
	PrivateIP    string
	PublicIP     string
	Config       string
	Image        string
	Type         string
	State        string
	CreationTime string
	Zone         string
}

// isolationGroupRow mirrors cmd/uhost.go isolationGroupRow byte-for-byte.
type isolationGroupRow struct {
	ResourceID string
	Name       string
	Remark     string
	UHostCount string
}
