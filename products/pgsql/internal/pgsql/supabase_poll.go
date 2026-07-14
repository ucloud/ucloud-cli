package pgsql

import (
	"fmt"

	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// describeSupabaseByID returns the poller's describe func: DescribeUSupabase → State.
// region/zone/projectID/memoryDB are captured from the originating command's flags so
// the poll uses the same scope (USupabase IAM requires the ProjectId context).
// Signature matches pkg/cli.Poller: func(string, *request.CommonBase) (interface{}, error).
func describeSupabaseByID(ctx *cli.Context, region, zone, projectID string, memoryDB bool) func(instanceID string, _ *request.CommonBase) (interface{}, error) {
	return func(instanceID string, _ *request.CommonBase) (interface{}, error) {
		params := map[string]interface{}{
			"Region":     region,
			"Zone":       zone,
			"ProjectId":  projectID,
			"InstanceID": instanceID,
		}
		if memoryDB {
			params["IsMemoryDB"] = true
		}
		payload, err := invokeSupabase(ctx, "DescribeUSupabase", params)
		if err != nil {
			return nil, err
		}
		ds, ok := payload["DataSet"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("pgsql supabase[%s] may not exist", instanceID)
		}
		return &supabaseState{State: getString(ds, "State")}, nil
	}
}
