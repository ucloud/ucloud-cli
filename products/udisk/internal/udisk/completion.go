package udisk

import (
	"strings"

	uhostsdk "github.com/ucloud/ucloud-sdk-go/services/uhost"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// getUhostList returns "UHostId/Name" completion candidates for the attach
// command's --uhost-id flag. Copied self-contained from cmd/uhost.go
// (base.BizClient → cli.NewServiceClient on the public uhost SDK).
func getUhostList(ctx *cli.Context, states []string, project, region, zone string) []string {
	client := cli.NewServiceClient(ctx, uhostsdk.NewClient)
	req := client.NewDescribeUHostInstanceRequest()
	req.ProjectId = sdk.String(project)
	req.Region = sdk.String(region)
	req.Zone = sdk.String(zone)
	req.Limit = sdk.Int(50)
	resp, err := client.DescribeUHostInstance(req)
	if err != nil {
		//todo runtime log
		return nil
	}
	list := []string{}
	for _, host := range resp.UHostSet {
		if states != nil {
			for _, s := range states {
				if host.State == s {
					list = append(list, host.UHostId+"/"+strings.Replace(host.Name, " ", "-", -1))
				}
			}
		} else {
			list = append(list, host.UHostId+"/"+strings.Replace(host.Name, " ", "-", -1))
		}
	}
	return list
}
