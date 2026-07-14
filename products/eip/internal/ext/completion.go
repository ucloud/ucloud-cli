package ext

import (
	"strings"

	uhostsdk "github.com/ucloud/ucloud-sdk-go/services/uhost"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

func listUHostIDs(ctx *cli.Context, states []string, project, region, zone string) []string {
	client := cli.NewServiceClient(ctx, uhostsdk.NewClient)
	req := client.NewDescribeUHostInstanceRequest()
	req.ProjectId = sdk.String(project)
	req.Region = sdk.String(region)
	req.Zone = sdk.String(zone)
	req.Limit = sdk.Int(50)
	resp, err := client.DescribeUHostInstance(req)
	if err != nil {
		return nil
	}
	list := []string{}
	for _, host := range resp.UHostSet {
		if !hostStateAllowed(host.State, states) {
			continue
		}
		list = append(list, host.UHostId+"/"+strings.Replace(host.Name, " ", "-", -1))
	}
	return list
}

func hostStateAllowed(state string, states []string) bool {
	if states == nil {
		return true
	}
	for _, s := range states {
		if state == s {
			return true
		}
	}
	return false
}
