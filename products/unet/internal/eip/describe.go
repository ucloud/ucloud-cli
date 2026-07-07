package eip

import (
	"fmt"
	"net"

	"github.com/ucloud/ucloud-sdk-go/services/unet"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

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
