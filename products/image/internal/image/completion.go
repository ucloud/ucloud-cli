package image

import (
	uhostsdk "github.com/ucloud/ucloud-sdk-go/services/uhost"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	cliconst "github.com/ucloud/ucloud-cli/model/cli"
	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// getImageList returns "ImageId/ImageName" completion candidates filtered by
// states and image type. Self-contained SDK call COPIED from cmd/image.go
// (base.BizClient → cli.NewServiceClient on the public uhost SDK).
func getImageList(ctx *cli.Context, states []string, imageType, project, region, zone string) []string {
	client := cli.NewServiceClient(ctx, uhostsdk.NewClient)
	req := client.NewDescribeImageRequest()
	req.ProjectId = &project
	req.Region = &region
	req.Zone = &zone
	req.Limit = sdk.Int(1000)
	if imageType != cliconst.IMAGE_ALL {
		req.ImageType = sdk.String(imageType)
	}
	resp, err := client.DescribeImage(req)
	if err != nil {
		return nil
	}
	list := []string{}
	for _, image := range resp.ImageSet {
		for _, s := range states {
			if image.State == s {
				list = append(list, image.ImageId+"/"+image.ImageName)
			}
		}
	}
	return list
}
