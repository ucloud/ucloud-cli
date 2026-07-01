package image

import (
	"strings"

	uhostsdk "github.com/ucloud/ucloud-sdk-go/services/uhost"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// describeImageByID returns the poller's describe func, closing over ctx (for an
// authed uhost client) and the project/region/zone the image lives in. The
// Spoll loop calls this with a nil commonBase, so the project/region/zone MUST
// be bound here — this is what makes cross-region copy (TargetRegion/
// TargetProjectId) converge. Ported from cmd/image.go's describeImageByID, whose
// (project, region, zone) args the legacy base.Poller.Poll passed through.
func describeImageByID(ctx *cli.Context, project, region, zone string) func(imageID string, commonBase *request.CommonBase) (interface{}, error) {
	return func(imageID string, commonBase *request.CommonBase) (interface{}, error) {
		client := cli.NewServiceClient(ctx, uhostsdk.NewClient)
		req := client.NewDescribeImageRequest()
		if commonBase != nil {
			req.CommonBase = *commonBase
		}
		req.ImageId = sdk.String(imageID)
		req.ProjectId = sdk.String(project)
		req.Region = sdk.String(region)
		req.Zone = sdk.String(zone)
		req.Limit = sdk.Int(50)
		resp, err := client.DescribeImage(req)
		if err != nil {
			return nil, err
		}
		if len(resp.ImageSet) < 1 {
			return nil, nil
		}
		return &resp.ImageSet[0], nil
	}
}

// getUhostList returns "UHostId/Name" completion candidates for the create
// command's --uhost-id flag. Copied self-contained from cmd/uhost.go
// (base.BizClient → cli.NewServiceClient on the public uhost SDK); image's
// create-image is its own copy and must not import the uhost product or cmd.
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
