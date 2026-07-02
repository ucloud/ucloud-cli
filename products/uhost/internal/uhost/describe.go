package uhost

import (
	"fmt"

	udisksdk "github.com/ucloud/ucloud-sdk-go/services/udisk"
	uhostsdk "github.com/ucloud/ucloud-sdk-go/services/uhost"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// describeUHostByID mirrors cmd/uhost.go's describeUHostByID (the REGION-aware,
// ERROR-on-not-found variant): it binds projectID/region/zone into the request
// and returns an error (not nil) when the uhost does not exist. The closure
// signature carries a *request.CommonBase only to satisfy ctx.PollerTo's
// describe-func type — it is intentionally ignored, because region/project/zone
// come from the bound args (this is what the sequential pollers and the direct
// resize/reset-password/checkAndCloseUhost/reinstall/leave-isolation callers
// passed at BASE via the 4-arg describe / Poll(id,proj,region,zone)). Returns
// *uhostsdk.UHostInstanceSet.
func describeUHostByID(ctx *cli.Context, projectID, region, zone string) func(uhostID string, commonBase *request.CommonBase) (interface{}, error) {
	return func(uhostID string, _ *request.CommonBase) (interface{}, error) {
		client := cli.NewServiceClient(ctx, uhostsdk.NewClient)
		req := client.NewDescribeUHostInstanceRequest()
		req.UHostIds = []string{uhostID}
		req.ProjectId = &projectID
		req.Region = &region
		req.Zone = &zone
		resp, err := client.DescribeUHostInstance(req)
		if err != nil {
			return nil, err
		}
		if len(resp.UHostSet) < 1 {
			return nil, fmt.Errorf("uhost [%s] does not exist", uhostID)
		}
		return &resp.UHostSet[0], nil
	}
}

// sdescribeUHostByID mirrors cmd/uhost.go's sdescribeUHostByID (the concurrent
// SPOLLER variant): nil-on-not-found and CommonBase-aware (a non-nil commonBase
// carries region/project/zone; nil falls back to the client's default-config
// region, which the SDK marshaler fills when the request region is empty). Used
// by the concurrent create/delete-stop Sspoll path and deleteUHost's lookup —
// the exact sites that used sdescribeUHostByID at BASE. Returns
// *uhostsdk.UHostInstanceSet.
func sdescribeUHostByID(ctx *cli.Context) func(uhostID string, commonBase *request.CommonBase) (interface{}, error) {
	return func(uhostID string, commonBase *request.CommonBase) (interface{}, error) {
		client := cli.NewServiceClient(ctx, uhostsdk.NewClient)
		req := client.NewDescribeUHostInstanceRequest()
		if commonBase != nil {
			req.CommonBase = *commonBase
		}
		req.UHostIds = []string{uhostID}
		resp, err := client.DescribeUHostInstance(req)
		if err != nil {
			return nil, err
		}
		if len(resp.UHostSet) < 1 {
			return nil, nil
		}
		return &resp.UHostSet[0], nil
	}
}

// describeUdiskByID returns the poller's describe func for udisk, used by
// detachUdisk. Copied self-contained from cmd/disk_compat.go (base.BizClient →
// cli.NewServiceClient).
func describeUdiskByID(ctx *cli.Context) func(udiskID string, commonBase *request.CommonBase) (interface{}, error) {
	return func(udiskID string, commonBase *request.CommonBase) (interface{}, error) {
		client := cli.NewServiceClient(ctx, udisksdk.NewClient)
		req := client.NewDescribeUDiskRequest()
		if commonBase != nil {
			req.CommonBase = *commonBase
		}
		req.UDiskId = sdk.String(udiskID)
		req.Limit = sdk.Int(50)
		resp, err := client.DescribeUDisk(req)
		if err != nil {
			return nil, err
		}
		if len(resp.DataSet) < 1 {
			return nil, nil
		}
		return &resp.DataSet[0], nil
	}
}

// describeImageByID returns the image-feature-probe describe func, closing over
// ctx + project/region/zone. Copied self-contained from cmd/image_compat.go;
// used by create to probe an image's HotPlug/CloudInit features and by
// create-image's poller. Returns *uhostsdk.UHostImageSet.
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
