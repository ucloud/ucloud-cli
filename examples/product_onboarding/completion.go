package onboarding

import (
	"fmt"

	"github.com/ucloud/ucloud-sdk-go/services/udb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// listResourceIDs returns the resource-id completion candidates for the
// `--<product>-id` flag, in the conventional "resourceID/name" form. A caller
// later runs the picked value through ctx.PickResourceID to strip the "/name"
// suffix back to a bare id.
//
// A real product filters by the current region/zone/project; here we close over
// ctx and read those from the bound request at completion time. The candidates
// double as a worked example of calling cli.NewServiceClient inside a
// completion provider.
//
// states, when non-nil, restricts candidates to those whose State is in the
// set — e.g. start completes only stopped instances, stop only running ones.
func listResourceIDs(ctx *cli.Context, states []string, region, zone, projectID string) []string {
	client := cli.NewServiceClient(ctx, udb.NewClient)

	req := client.NewDescribeUDBInstanceRequest()
	req.Region = sdk.String(region)
	req.Zone = sdk.String(zone)
	req.ProjectId = sdk.String(projectID)
	req.ClassType = sdk.String("sql")
	req.Limit = sdk.Int(100)

	resp, err := client.DescribeUDBInstance(req)
	if err != nil {
		// Completion must never error out the shell; degrade to no candidates.
		return nil
	}

	candidates := make([]string, 0, len(resp.DataSet))
	for _, ins := range resp.DataSet {
		if !stateAllowed(ins.State, states) {
			continue
		}
		candidates = append(candidates, fmt.Sprintf("%s/%s", ins.DBId, ins.Name))
	}
	return candidates
}

// stateAllowed reports whether state passes the optional allow-list. A nil
// allow-list means "any state".
func stateAllowed(state string, states []string) bool {
	if states == nil {
		return true
	}
	for _, s := range states {
		if s == state {
			return true
		}
	}
	return false
}

// derefStr safely dereferences a *string bound by a flag, returning "" for nil.
func derefStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
