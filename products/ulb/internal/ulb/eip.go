package ulb

import (
	"fmt"
	"net"
	"strings"

	"github.com/ucloud/ucloud-sdk-go/services/unet"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

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

func getEIPLine(region string) (line string) {
	if strings.HasPrefix(region, "cn") {
		line = "BGP"
	} else {
		line = "International"
	}
	return
}
