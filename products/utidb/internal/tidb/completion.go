package tidb

import (
	"fmt"

	"github.com/ucloud/ucloud-sdk-go/services/tidb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// listResourceIDs returns UTiDB instance IDs formatted as id/name.
func listResourceIDs(ctx *cli.Context, states []string, region, zone, projectID string) []string {
	client := cli.NewServiceClient(ctx, tidb.NewClient)
	req := client.NewListTiDBClusterServiceRequest()
	if region != "" {
		req.Region = sdk.String(region)
	}
	if zone != "" {
		req.Zone = sdk.String(zone)
	}
	if projectID != "" {
		req.ProjectId = sdk.String(projectID)
	}
	resp, err := client.ListTiDBClusterService(req)
	if err != nil {
		return nil
	}
	var list []string
	for _, d := range resp.Data {
		if states != nil {
			found := false
			for _, s := range states {
				if s == d.State {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		list = append(list, fmt.Sprintf("%s/%s", d.Id, d.Name))
	}
	return list
}

// listNodeTypes returns all available node types by querying specs.
func listNodeTypes(ctx *cli.Context, region, zone string) []string {
	// The API requires NodeTypes as input, so we seed with known types and
	// collect the actual available set from the response.
	seedTypes := []string{"tidb", "tikv", "pd", "tiflash"}
	client := cli.NewServiceClient(ctx, tidb.NewClient)
	req := client.NewGetTiDBClusterUhostSpecsRequest()
	if region != "" {
		req.Region = sdk.String(region)
	}
	req.NodeTypes = seedTypes
	resp, err := client.GetTiDBClusterUhostSpecs(req)
	if err != nil {
		return nil
	}
	seen := make(map[string]bool)
	var list []string
	for _, s := range resp.Data {
		if !seen[s.NodeType] {
			seen[s.NodeType] = true
			list = append(list, s.NodeType)
		}
	}
	return list
}

// listConfigIDs returns uhost config IDs for the given node type.
func listConfigIDs(ctx *cli.Context, region, zone, nodeType string) []string {
	if nodeType == "" {
		return nil
	}
	client := cli.NewServiceClient(ctx, tidb.NewClient)
	req := client.NewGetTiDBClusterUhostSpecsRequest()
	if region != "" {
		req.Region = sdk.String(region)
	}
	req.NodeTypes = []string{nodeType}
	resp, err := client.GetTiDBClusterUhostSpecs(req)
	if err != nil {
		return nil
	}
	var list []string
	for _, s := range resp.Data {
		list = append(list, fmt.Sprintf("%s/%s", s.ConfigId, s.ConfigName))
	}
	return list
}

// listServerIDs returns the server IDs of the given UTiDB instance.
// The current TiDB SDK does not expose server IDs in GetTiDBClusterService,
// so this function returns an empty list.
func listServerIDs(ctx *cli.Context, id string) []string {
	_ = ctx
	_ = id
	return nil
}
