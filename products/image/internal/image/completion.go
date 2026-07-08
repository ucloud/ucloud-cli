package image

import (
	"strings"

	uhostsdk "github.com/ucloud/ucloud-sdk-go/services/uhost"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

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
	if imageType != IMAGE_ALL {
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
