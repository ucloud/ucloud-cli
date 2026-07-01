package eip

import (
	"fmt"
	"net"
	"strings"

	"github.com/ucloud/ucloud-sdk-go/services/unet"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

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
// Ported from cmd/eip.go (base.BizClient → cli.NewServiceClient).
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

// fetchAllEip lists all EIPs in project/region, paging by 100. Ported from
// cmd/eip.go (base.BizClient → cli.NewServiceClient).
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

// getEIP fetches a single EIP by id. Ported from cmd/eip.go
// (base.BizClient → cli.NewServiceClient).
func getEIP(ctx *cli.Context, eipID string) (*unet.UnetEIPSet, error) {
	client := cli.NewServiceClient(ctx, unet.NewClient)
	req := client.NewDescribeEIPRequest()
	req.EIPIds = append(req.EIPIds, eipID)
	resp, err := client.DescribeEIP(req)
	if err != nil {
		return nil, err
	}
	if len(resp.EIPSet) == 1 {
		return &resp.EIPSet[0], nil
	}
	return nil, fmt.Errorf("eip[%s] may not exist", eipID)
}

// bindEIP binds an EIP to a resource. Ported from cmd/eip.go
// (base.BizClient → cli.NewServiceClient; progress→ProgressWriter,
// errors→ctx.HandleError). Returns a non-nil error when the bind fails so the
// caller only emits a structured "Bound" result on success (machine output must
// not report success for a failed operation).
func bindEIP(ctx *cli.Context, resourceID, resourceType, eipID, projectID, region *string) error {
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
		ctx.HandleError(err)
		return err
	}
	fmt.Fprintf(ctx.ProgressWriter(), "bind EIP[%s] with %s[%s]\n", *req.EIPId, *req.ResourceType, *req.ResourceId)
	return nil
}

// sbindEIP binds an EIP to a resource, returning a log trail instead of
// printing (used for concurrent flows). Ported from cmd/eip.go; the
// base.ToQueryMap request-log line is dropped (platform SDK handler logs
// requests now, D-C).
// Retained as the canonical product-local copy for uhost Part 6 (see batch-1
// plan); not yet called within the eip product.
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

// unbindEIP unbinds an EIP from a resource, returning a log trail. Ported from
// cmd/eip.go; the base.ToQueryMap request-log line is dropped (platform SDK
// handler logs requests now, D-C).
// Retained as the canonical product-local copy for uhost Part 6 (see batch-1
// plan); not yet called within the eip product.
func unbindEIP(ctx *cli.Context, resourceID, resourceType, eipID, projectID, region string) ([]string, error) {
	logs := make([]string, 0)
	eipID = ctx.PickResourceID(eipID)
	ip := net.ParseIP(eipID)
	if ip != nil {
		id, err := getEIPIDbyIP(ctx, ip, projectID, region)
		if err != nil {
			ctx.HandleError(err)
		} else {
			eipID = id
		}
	}
	client := cli.NewServiceClient(ctx, unet.NewClient)
	req := client.NewUnBindEIPRequest()
	req.ResourceId = &resourceID
	req.ResourceType = &resourceType
	req.EIPId = &eipID
	req.ProjectId = sdk.String(ctx.PickResourceID(projectID))
	req.Region = &region
	_, err := client.UnBindEIP(req)
	if err != nil {
		logs = append(logs, fmt.Sprintf("unbind eip failed: %v", err))
		return logs, err
	}
	logs = append(logs, fmt.Sprintf("unbind eip[%s] with %s[%s] successfully", *req.EIPId, *req.ResourceType, *req.ResourceId))
	return logs, nil
}
