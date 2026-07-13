package ulhost

import (
	"fmt"

	ucompsharesdk "github.com/ucloud/ucloud-sdk-go/services/ucompshare"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/base"
	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// SdescribeULHostByID is the concurrent SPOLLER variant of the ulhost describe
// function: nil-on-not-found and CommonBase-aware (a non-nil commonBase carries
// region/project/zone; nil falls back to the client's default-config region,
// which the SDK marshaler fills when the request region is empty). Used by
// cmd/api.go's RepeatsSupportedAPI for CreateULHostInstance polling — the exact
// site that used sdescribeULHostByID at BASE. Returns *ucompsharesdk.ULHostInstanceSet.
//
// This is the legacy path that uses base.BizClient (no cli.Context), so that
// cmd/api.go can continue to reference it at package init time. The ctx-based
// variant below (sdescribeULHostByID) is the new-platform path.
func SdescribeULHostByID(ulhostID string, common *request.CommonBase) (interface{}, error) {
	req := base.BizClient.UCompShareClient.NewDescribeULHostInstanceRequest()
	req.ULHostIds = []string{ulhostID}
	if common != nil {
		req.CommonBase = *common
	}
	resp, err := base.BizClient.UCompShareClient.DescribeULHostInstance(req)
	if err != nil {
		return nil, err
	}
	if len(resp.ULHostInstanceSets) < 1 {
		return nil, nil
	}

	return &resp.ULHostInstanceSets[0], nil
}

// sdescribeULHostByID is the new-platform ctx-based variant (mirrors uhost's
// sdescribeUHostByID pattern): it uses cli.NewServiceClient instead of
// base.BizClient. Nil-on-not-found and CommonBase-aware. For use by ulhost
// product commands. Returns *ucompsharesdk.ULHostInstanceSet.
func sdescribeULHostByID(ctx *cli.Context) func(ulhostID string, commonBase *request.CommonBase) (interface{}, error) {
	return func(ulhostID string, commonBase *request.CommonBase) (interface{}, error) {
		client := cli.NewServiceClient(ctx, ucompsharesdk.NewClient)
		req := client.NewDescribeULHostInstanceRequest()
		if commonBase != nil {
			req.CommonBase = *commonBase
		}
		req.ULHostIds = []string{ulhostID}
		resp, err := client.DescribeULHostInstance(req)
		if err != nil {
			return nil, err
		}
		if len(resp.ULHostInstanceSets) < 1 {
			return nil, nil
		}
		return &resp.ULHostInstanceSets[0], nil
	}
}

// describeULHostByID mirrors uhost's describeUHostByID (the ERROR-on-not-found
// variant): it binds projectID/region into the request and returns an error
// (not nil) when the ulhost does not exist. Used by sequential pollers.
// Returns *ucompsharesdk.ULHostInstanceSet.
func describeULHostByID(ctx *cli.Context, projectID, region string) func(ulhostID string, commonBase *request.CommonBase) (interface{}, error) {
	return func(ulhostID string, _ *request.CommonBase) (interface{}, error) {
		client := cli.NewServiceClient(ctx, ucompsharesdk.NewClient)
		req := client.NewDescribeULHostInstanceRequest()
		req.ULHostIds = []string{ulhostID}
		req.ProjectId = sdk.String(projectID)
		req.Region = sdk.String(region)
		resp, err := client.DescribeULHostInstance(req)
		if err != nil {
			return nil, err
		}
		if len(resp.ULHostInstanceSets) < 1 {
			return nil, fmt.Errorf("ulhost [%s] does not exist", ulhostID)
		}
		return &resp.ULHostInstanceSets[0], nil
	}
}
