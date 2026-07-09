package eip

import (
	"strings"

	"github.com/ucloud/ucloud-sdk-go/services/unet"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// getAllEip returns "EIPId/ip1,ip2" completion candidates filtered by states and
// paymodes (nil filter = no filter). Ported from cmd/eip.go; uses the
// package-local fetchAllEip.
func getAllEip(ctx *cli.Context, projectID, region string, states, paymodes []string) []string {
	list, err := fetchAllEip(ctx, projectID, region)
	if err != nil {
		return nil
	}
	strs := []string{}
	for _, item := range list {
		rightState := false
		if states == nil {
			rightState = true
		} else {
			for _, s := range states {
				if item.Status == s {
					rightState = true
				}
			}
		}

		rightPayMode := false
		if paymodes == nil {
			rightPayMode = true
		} else {
			for _, m := range paymodes {
				if item.PayMode == m {
					rightPayMode = true
				}
			}
		}
		if !rightPayMode || !rightState {
			continue
		}

		ips := []string{}
		for _, ip := range item.EIPAddr {
			ips = append(ips, ip.IP)
		}
		strs = append(strs, item.EIPId+"/"+strings.Join(ips, ","))
	}
	return strs
}

// getAllSharedBW returns "ShareBandwidthId/Name" completion candidates for
// shared bandwidth instances in project/region. Self-contained SDK call COPIED
// from cmd/bandwidth.go getAllSharedBW (base.BizClient → cli.NewServiceClient),
// for the join-shared-bw / leave-shared-bw --shared-bw-id completion.
func getAllSharedBW(ctx *cli.Context, project, region string) ([]string, error) {
	client := cli.NewServiceClient(ctx, unet.NewClient)
	req := client.NewDescribeShareBandwidthRequest()
	req.ProjectId = &project
	req.Region = &region
	resp, err := client.DescribeShareBandwidth(req)
	if err != nil {
		return nil, err
	}
	list := []string{}
	for _, item := range resp.DataSet {
		list = append(list, item.ShareBandwidthId+"/"+item.Name)
	}
	return list, nil
}
