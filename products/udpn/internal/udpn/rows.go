package udpn

// UDPNRow is the table row for `ucloud udpn list`.
type UDPNRow struct {
	ResourceID   string
	Peers        string
	Bandwidth    string
	ChargeType   string
	CreationTime string
}
