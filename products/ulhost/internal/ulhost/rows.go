package ulhost

// rows.go holds the table-row structs for ulhost list output. The platform
// printer (ctx.PrintList) derives table columns from a struct's exported fields
// in declaration order. JSON output uses the full ulhostRow (matching the
// original convention of marshalling the full row slice in --json mode).

// ulhostRow is the full row (wide mode + json). Field set+order matches the
// ucompshare SDK ULHostInstanceSet fields relevant for CLI display.
type ulhostRow struct {
	Name         string
	ResourceID   string
	Remark       string
	Group        string
	PrivateIP    string
	PublicIP     string
	Config       string
	DiskSet      string
	Zone         string
	Image        string
	State        string
	ChargeType   string
	AutoRenew    string
	ExpireTime   string
	CreationTime string
}

// ulhostRowDefault is the default (non-all-region) column set:
// Name, ResourceID, Group, PublicIP, Config, Image, State, ChargeType, CreationTime
type ulhostRowDefault struct {
	Name         string
	ResourceID   string
	Group        string
	PublicIP     string
	Config       string
	Image        string
	State        string
	ChargeType   string
	CreationTime string
}

// ulhostRowAllRegion is the default column set plus a trailing Zone column.
type ulhostRowAllRegion struct {
	Name         string
	ResourceID   string
	Group        string
	PublicIP     string
	Config       string
	Image        string
	State        string
	ChargeType   string
	CreationTime string
	Zone         string
}

// bundleRow mirrors the Bundle struct for table display.
type bundleRow struct {
	BundleID      string
	CPU           string
	Memory        string
	SysDiskSpace  string
	Bandwidth     string
	TrafficPacket string
}

// priceRow mirrors the ULHostPriceSet for table display.
type priceRow struct {
	ChargeType    string
	Price         string
	OriginalPrice string
}
