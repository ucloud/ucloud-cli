package tidb

import (
	"io"

	"github.com/ucloud/ucloud-sdk-go/services/tidb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

func createPollTargets() []string {
	return []string{stateAvailable, stateRunning, stateCreateFail}
}

func upgradePollTargets() []string {
	return []string{stateAvailable, stateRunning, stateUpgradeFail}
}

func spollCreate(ctx *cli.Context, w io.Writer, region, zone, projectID, id, text string) {
	ctx.PollerTo(w, describeByID(ctx, region, zone, projectID)).
		Spoll(id, text, createPollTargets())
}

func spollUpgrade(ctx *cli.Context, w io.Writer, region, zone, projectID, id, text string) {
	ctx.PollerTo(w, describeByID(ctx, region, zone, projectID)).
		Spoll(id, text, upgradePollTargets())
}

// describeByID returns a poller function that reads a UTiDB instance by ID.
// The returned data is a pointer to UTiDBServiceData so the poller can read its
// State field via reflection.
func describeByID(ctx *cli.Context, region, zone, projectID string) func(string, *request.CommonBase) (interface{}, error) {
	return func(id string, _ *request.CommonBase) (interface{}, error) {
		client := cli.NewServiceClient(ctx, tidb.NewClient)
		req := client.NewGetTiDBClusterServiceRequest()
		if region != "" {
			req.Region = sdk.String(region)
		}
		if zone != "" {
			req.Zone = sdk.String(zone)
		}
		if projectID != "" {
			req.ProjectId = sdk.String(projectID)
		}
		req.Id = sdk.String(id)
		resp, err := client.GetTiDBClusterService(req)
		if err != nil {
			return nil, err
		}
		return &resp.Data, nil
	}
}
