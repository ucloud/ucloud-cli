package css

import (
	"strings"

	uessdk "github.com/ucloud/ucloud-sdk-go/services/ues"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// getInstanceList returns "InstanceId/Name" completion candidates for instance-id flags.
func getInstanceList(ctx *cli.Context, states []string, project, region, zone string) []string {
	client := cli.NewServiceClient(ctx, uessdk.NewClient)
	req := client.NewListUESInstanceRequest()
	req.ProjectId = sdk.String(project)
	req.Region = sdk.String(region)
	if zone != "" {
		req.Zone = sdk.String(zone)
	}
	req.Limit = sdk.Int(50)
	resp, err := client.ListUESInstance(req)
	if err != nil {
		// silent fail for completion
		return nil
	}
	list := []string{}
	for _, ins := range resp.ClusterSet {
		if states != nil {
			matched := false
			for _, s := range states {
				if ins.State == s {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}
		list = append(list, ins.InstanceId+"/"+strings.Replace(ins.InstanceName, " ", "-", -1))
	}
	return list
}
