package uk8s

// clusterRow is the row struct for `ucloud uk8s list` output. Column order
// (field declaration order) drives the table layout; the platform's
// ctx.PrintList renders the row for any --output.
type clusterRow struct {
	ResourceID string
	Name       string
	K8sVersion string
	VPCID      string
	SubnetID   string
	MasterCnt  int
	NodeCnt    int
	Status     string
	Group      string
	Created    string
}