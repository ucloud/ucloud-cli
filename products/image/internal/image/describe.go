package image

import (
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
