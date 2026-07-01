package firewall

import (
	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// getFirewallIDNames returns "FWId/Name" completion candidates for firewalls in
// project/region. Self-contained SDK call (ported from cmd/firewall.go, which
// shared getAllFirewallIns; here it calls the package-local getAllFirewallIns).
func getFirewallIDNames(ctx *cli.Context, project, region string) (idNames []string) {
	list, err := getAllFirewallIns(ctx, project, region)
	if err != nil {
		return
	}
	for _, f := range list {
		idNames = append(idNames, f.FWId+"/"+f.Name)
	}
	return
}
