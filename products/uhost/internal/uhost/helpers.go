package uhost

import (
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	udisksdk "github.com/ucloud/ucloud-sdk-go/services/udisk"
	uhostsdk "github.com/ucloud/ucloud-sdk-go/services/uhost"
	"github.com/ucloud/ucloud-sdk-go/services/unet"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/model/status"
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

// getEIPByUHostId polls (up to 6 times) for a non-private EIP bound to the uhost.
// Ported verbatim from cmd/uhost.go getEIPByUHostId (base.BizClient →
// cli.NewServiceClient).
func getEIPByUHostId(ctx *cli.Context, uhostId string) (*uhostsdk.UHostIPSet, error) {
	if uhostId == "" {
		return nil, fmt.Errorf("the uhost[%s] is not found", uhostId)
	}
	client := cli.NewServiceClient(ctx, uhostsdk.NewClient)
	for i := 0; i <= 5; i++ {
		req := client.NewDescribeUHostInstanceRequest()
		req.UHostIds = []string{uhostId}

		resp, err := client.DescribeUHostInstance(req)
		if err != nil {
			return nil, err
		}
		if len(resp.UHostSet) < 1 {
			return nil, fmt.Errorf("the uhost[%s] is not found", uhostId)
		}

		if len(resp.UHostSet[0].IPSet) > 0 {
			for _, v := range resp.UHostSet[0].IPSet {
				if v.Type != "Private" && v.IPId != "" {
					return &v, nil
				}
			}
		}

		time.Sleep(1 * time.Second)
	}

	return nil, fmt.Errorf("can not get eip by uhost[%s]", uhostId)
}

// getUhostList returns "UHostId/Name" completion candidates filtered by states
// (nil = all). Copied self-contained from cmd/uhost.go getUhostList
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

// getIsolationGroupList returns "GroupId/Name" completion candidates. Copied
// self-contained from cmd/uhost.go getIsolationGroupList (the original printed
// the fetch error to stdout; that diagnostic is dropped — completion funcs must
// stay silent so they don't corrupt shell completion output).
func getIsolationGroupList(ctx *cli.Context, project, region string) []string {
	client := cli.NewServiceClient(ctx, uhostsdk.NewClient)
	req := client.NewDescribeIsolationGroupRequest()
	req.ProjectId = sdk.String(project)
	req.Region = sdk.String(region)
	req.Limit = sdk.Int(50)
	resp, err := client.DescribeIsolationGroup(req)
	if err != nil {
		return nil
	}
	list := []string{}
	for _, group := range resp.IsolationGroupSet {
		list = append(list, group.GroupId+"/"+strings.Replace(group.GroupName, " ", "-", -1))
	}
	return list
}

// getEIPLine returns the default EIP line for a region. Product-local copy of
// cmd/util.go getEIPLine (domain logic, D-D: COPIED into the product, never
// promoted to platform). "cn" regions default to BGP, others to International.
func getEIPLine(region string) (line string) {
	if strings.HasPrefix(region, "cn") {
		line = "BGP"
	} else {
		line = "International"
	}
	return
}

// getEIPIDbyIP resolves an EIP id from an IP address within project/region.
// Copied self-contained from cmd/eip_compat.go (base.BizClient →
// cli.NewServiceClient) so sbindEIP can accept an IP literal.
func getEIPIDbyIP(ctx *cli.Context, ip net.IP, projectID, region string) (string, error) {
	eipList, err := fetchAllEip(ctx, projectID, region)
	if err != nil {
		return "", err
	}
	for _, eip := range eipList {
		for _, addr := range eip.EIPAddr {
			if addr.IP == ip.String() {
				return eip.EIPId, nil
			}
		}
	}
	return "", fmt.Errorf("IP[%s] not exist", ip.String())
}

// fetchAllEip lists all EIPs in project/region, paging by 100. Copied
// self-contained from cmd/eip_compat.go (base.BizClient → cli.NewServiceClient).
func fetchAllEip(ctx *cli.Context, projectID, region string) ([]unet.UnetEIPSet, error) {
	client := cli.NewServiceClient(ctx, unet.NewClient)
	req := client.NewDescribeEIPRequest()
	list := []unet.UnetEIPSet{}
	req.ProjectId = sdk.String(projectID)
	req.Region = sdk.String(region)
	for offset, step := 0, 100; ; offset += step {
		req.Offset = &offset
		req.Limit = &step
		resp, err := client.DescribeEIP(req)
		if err != nil {
			return nil, err
		}
		for i, size := 0, len(resp.EIPSet); i < size; i++ {
			list = append(list, resp.EIPSet[i])
		}
		if resp.TotalCount <= offset+step {
			break
		}
	}
	return list, nil
}

// sbindEIP binds an EIP to a resource, returning a log trail instead of printing
// (used for the concurrent create flow). Copied self-contained from
// cmd/eip_compat.go; the base.ToQueryMap request-log line is dropped (platform
// SDK handler logs requests now, D-C).
func sbindEIP(ctx *cli.Context, resourceID, resourceType, eipID, projectID, region *string) ([]string, error) {
	logs := make([]string, 0)
	ip := net.ParseIP(*eipID)
	if ip != nil {
		id, err := getEIPIDbyIP(ctx, ip, *projectID, *region)
		if err != nil {
			ctx.HandleError(err)
		} else {
			*eipID = id
		}
	}
	client := cli.NewServiceClient(ctx, unet.NewClient)
	req := client.NewBindEIPRequest()
	req.ResourceId = resourceID
	req.ResourceType = resourceType
	req.EIPId = sdk.String(ctx.PickResourceID(*eipID))
	req.ProjectId = sdk.String(ctx.PickResourceID(*projectID))
	req.Region = region
	_, err := client.BindEIP(req)
	if err != nil {
		logs = append(logs, fmt.Sprintf("bind eip failed: %v", err))
		return logs, err
	}
	logs = append(logs, fmt.Sprintf("bind eip[%s] with %s[%s] successfully", *req.EIPId, *req.ResourceType, *req.ResourceId))
	return logs, nil
}

// detachUdisk detaches a udisk from its uhost, narrating progress to out.
// Copied self-contained from cmd/disk_compat.go (base.BizClient →
// cli.NewServiceClient), used by reinstall-os before reinstalling the OS.
func detachUdisk(ctx *cli.Context, async bool, udiskID string, out io.Writer) error {
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
