package nlb

// NLBRow is the output row for `nlb list`. Field declaration order is the
// table column order.
type NLBRow struct {
	ResourceID       string
	Name             string
	Status           string
	VPC              string
	Subnet           string
	IPVersion        string
	ForwardingMode   string
	AutoRenewEnabled bool
	PurchaseValue    string
	Group            string
	CreationTime     string
}

// ListenerRow is the output row for `nlb listener list`.
type ListenerRow struct {
	ListenerID         string
	Name               string
	Protocol           string
	Scheduler          string
	PortRange          string
	ForwardSrcIPMethod string
	State              string
	StickinessTimeout  int
	HealthCheckType    string
	HealthCheckPort    int
	HealthCheckReqMsg  string
	HealthCheckResMsg  string
}

// ListenerDetailRow is the output row for the "Listeners" detail table
// appended under `nlb describe`. It surfaces the per-listener fields the
// summary "ListenerCount" attribute row can't. HealthCheckType/HealthCheckPort
// give a health-check summary; TargetCount is an at-a-glance count, with the
// full per-target breakdown following right after in the "Targets of ..."
// tables (see TargetRow) — keeping this row itself flat/one-level-deep.
type ListenerDetailRow struct {
	ListenerID         string
	Name               string
	Protocol           string
	PortRange          string
	Scheduler          string
	ForwardSrcIPMethod string
	StickinessTimeout  int
	State              string
	HealthCheckType    string
	HealthCheckPort    int
	HealthCheckReqMsg  string
	HealthCheckResMsg  string
	TargetCount        int
}

// TargetRow is the output row for the per-listener "Targets of ..." tables
// appended under `nlb listener list` and `nlb describe` (Target sits one
// level below Listener — NLB → Listener → Target — so it is surfaced there,
// grouped by listener, rather than via a standalone `nlb target list`
// command).
type TargetRow struct {
	TargetID     string
	Name         string
	ResourceType string
	ResourceID   string
	Port         int
	Weight       int
	Enabled      bool
	State        string
}
