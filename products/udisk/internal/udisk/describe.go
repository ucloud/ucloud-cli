package udisk

import (
	puhost "github.com/ucloud/ucloud-sdk-go/private/services/uhost"
	udisksdk "github.com/ucloud/ucloud-sdk-go/services/udisk"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// describeUdiskByID returns the poller's describe func, closing over ctx so it
// can build an authed udisk client. Mirrors cmd/disk.go's describeUdiskByID.
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

// describeSnapshotByID returns the poller's describe func for udisk snapshots.
// Mirrors cmd/disk.go's describeSnapshotByID (private uhost DescribeSnapshot).
func describeSnapshotByID(ctx *cli.Context) func(snapshotID string, commonBase *request.CommonBase) (interface{}, error) {
	return func(snapshotID string, commonBase *request.CommonBase) (interface{}, error) {
		client := cli.NewServiceClient(ctx, puhost.NewClient)
		req := client.NewDescribeSnapshotRequest()
		if commonBase != nil {
			req.CommonBase = *commonBase
		}
		req.SnapshotIds = append(req.SnapshotIds, snapshotID)
		req.Limit = sdk.Int(50)
		resp, err := client.DescribeSnapshot(req)
		if err != nil {
			return nil, err
		}
		if len(resp.UHostSnapshotSet) != 1 {
			return nil, nil
		}
		return &resp.UHostSnapshotSet[0], nil
	}
}
