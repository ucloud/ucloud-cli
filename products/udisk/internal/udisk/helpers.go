package udisk

import (
	"fmt"
	"io"
	"strings"

	puhost "github.com/ucloud/ucloud-sdk-go/private/services/uhost"
	udisksdk "github.com/ucloud/ucloud-sdk-go/services/udisk"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/model/status"
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

func getDiskList(ctx *cli.Context, states []string, project, region, zone string) []string {
	client := cli.NewServiceClient(ctx, udisksdk.NewClient)
	req := client.NewDescribeUDiskRequest()
	req.ProjectId = sdk.String(project)
	req.Region = sdk.String(region)
	req.Zone = sdk.String(zone)
	req.Limit = sdk.Int(50)
	resp, err := client.DescribeUDisk(req)
	if err != nil {
		//todo runtime log
		return nil
	}
	list := []string{}
	for _, disk := range resp.DataSet {
		for _, s := range states {
			if disk.Status == s {
				list = append(list, disk.UDiskId+"/"+strings.Replace(disk.Name, " ", "-", -1))
			}
		}
	}
	return list
}

func getSnapshotList(ctx *cli.Context, states []string, project, region, zone string) []string {
	client := cli.NewServiceClient(ctx, udisksdk.NewClient)
	req := client.NewDescribeUDiskSnapshotRequest()
	req.Limit = sdk.Int(50)
	req.ProjectId = &project
	req.Region = &region
	req.Zone = &zone
	resp, err := client.DescribeUDiskSnapshot(req)
	if err != nil {
		return nil
	}
	list := []string{}
	for _, snapshot := range resp.DataSet {
		for _, s := range states {
			if snapshot.Status == s {
				list = append(list, snapshot.SnapshotId+"/"+strings.Replace(snapshot.Name, " ", "-", -1))
			}
		}
	}
	return list
}

// DetachUdisk detaches a udisk from its uhost, narrating progress to out.
// Ported from cmd/disk.go's detachUdisk (base.BizClient → cli.NewServiceClient);
// exported so the restore command and (legacy) callers share one copy.
func DetachUdisk(ctx *cli.Context, async bool, udiskID string, out io.Writer) error {
	any, err := describeUdiskByID(ctx)(udiskID, nil)
	if err != nil {
		return err
	}
	if any == nil {
		return fmt.Errorf("udisk[%v] is not exist", any)
	}
	ins, ok := any.(*udisksdk.UDiskDataSet)
	if !ok {
		return fmt.Errorf("%#v convert to udisk failed", any)
	}
	client := cli.NewServiceClient(ctx, udisksdk.NewClient)
	req := client.NewDetachUDiskRequest()
	req.UHostId = sdk.String(ins.UHostId)
	req.UDiskId = sdk.String(udiskID)
	resp, err := client.DetachUDisk(req)
	if err != nil {
		return err
	}
	text := fmt.Sprintf("udisk[%s] is detaching from uhost[%s]", resp.UDiskId, resp.UHostId)
	if async {
		fmt.Fprintln(out, text)
	} else {
		ctx.PollerTo(out, describeUdiskByID(ctx)).Spoll(udiskID, text, []string{status.DISK_AVAILABLE, status.DISK_FAILED})
	}
	return nil
}
