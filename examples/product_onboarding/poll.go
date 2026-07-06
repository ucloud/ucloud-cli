package onboarding

import (
	"fmt"

	"github.com/ucloud/ucloud-sdk-go/services/udb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// Poller plumbing shared by the long-running verbs — not the "describe" verb (that lives in describe.go).

// describeByID returns the Poller describe func: given a resource id it fetches
// the current resource so the Poller can read its state field. The signature
// (func(string, *request.CommonBase) (interface{}, error)) is exactly what
// ctx.PollerTo expects.
func describeByID(ctx *cli.Context) func(string, *request.CommonBase) (interface{}, error) {
	return func(id string, common *request.CommonBase) (interface{}, error) {
		client := cli.NewServiceClient(ctx, udb.NewClient)
		req := client.NewDescribeUDBInstanceRequest()
		if common != nil {
			req.CommonBase = *common
		}
		req.DBId = sdk.String(id)
		resp, err := client.DescribeUDBInstance(req)
		if err != nil {
			return nil, err
		}
		if len(resp.DataSet) == 0 {
			return nil, fmt.Errorf("instance %q not found", id)
		}
		// Return a *struct whose exported State field the Poller reads by name.
		return &resp.DataSet[0], nil
	}
}
