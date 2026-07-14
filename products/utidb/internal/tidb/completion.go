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
	seedTypes := []string{"tidb", "tikv", "pd", "tiflash"}
	specs, err := getTiDBClusterUhostSpecs(ctx, region, zone, "", seedTypes)
	if err != nil {
		return nil
	}
	seen := make(map[string]bool)
	var list []string
	for _, s := range specs {
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
	specs, err := getTiDBClusterUhostSpecs(ctx, region, zone, "", []string{nodeType})
	if err != nil {
		return nil
	}
	var list []string
	for _, s := range specs {
		list = append(list, fmt.Sprintf("%s/%s", s.ConfigId, s.ConfigName))
	}
	return list
}

// listServerIDs returns server IDs of the given UTiDB instance for scale-in completion.
// Format: <server-id>/<node-type>@<host-ip>
func listServerIDs(ctx *cli.Context, region, zone, projectID, id string) []string {
	if id == "" {
		return nil
	}
	payload, err := getTiDBClusterPayload(ctx, region, zone, projectID, id)
	if err != nil {
		return nil
	}
	return extractServerIDs(payload)
}
