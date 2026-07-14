package umongodb

// instanceRow is the output struct for `umongodb list`. When passed to
// ctx.PrintList in table mode, the exported field NAMES become the column
// headers, in declaration order.
type instanceRow struct {
	ResourceID  string
	Name        string
	ClusterType string
	Version     string
	MachineType string
	DiskGB      int
	ConnectURL  string
	Status      string
}
