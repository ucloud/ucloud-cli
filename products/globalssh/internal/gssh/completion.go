package gssh

import (
	"fmt"
	"strings"

	pathxsdk "github.com/ucloud/ucloud-sdk-go/services/pathx"
	"github.com/ucloud/ucloud-sdk-go/services/unet"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

func getAllGssh(ctx *cli.Context, project string) ([]pathxsdk.GlobalSSHInfo, error) {
	client := cli.NewServiceClient(ctx, pathxsdk.NewClient)
	req := client.NewDescribeGlobalSSHInstanceRequest()
	req.ProjectId = sdk.String(cli.PickResourceID(project))
	resp, err := client.DescribeGlobalSSHInstance(req)
	if err != nil {
		return nil, err
	}
	return resp.InstanceSet, nil
}

func getAllGsshIDNames(ctx *cli.Context, project string) []string {
	gsshs, err := getAllGssh(ctx, project)
	if err != nil {
		return nil
	}
	list := []string{}
	for _, gssh := range gsshs {
		list = append(list, fmt.Sprintf("%s/%s", gssh.InstanceId, gssh.TargetIP))
	}
	return list
}

func getAllEip(ctx *cli.Context, projectID, region string, states, paymodes []string) []string {
	list, err := fetchAllEip(ctx, projectID, region)
	if err != nil {
		return nil
	}
	strs := []string{}
	for _, item := range list {
		rightState := states == nil
		for _, s := range states {
			if item.Status == s {
				rightState = true
			}
		}
		rightPayMode := paymodes == nil
		for _, m := range paymodes {
			if item.PayMode == m {
				rightPayMode = true
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

func fetchAllEip(ctx *cli.Context, projectID, region string) ([]unet.UnetEIPSet, error) {
	client := cli.NewServiceClient(ctx, unet.NewClient)
	req := client.NewDescribeEIPRequest()
	list := []unet.UnetEIPSet{}
	req.ProjectId = sdk.String(cli.PickResourceID(projectID))
	req.Region = sdk.String(region)
	for offset, step := 0, 100; ; offset += step {
		req.Offset = sdk.Int(offset)
		req.Limit = sdk.Int(step)
		resp, err := client.DescribeEIP(req)
		if err != nil {
			return nil, err
		}
		list = append(list, resp.EIPSet...)
		if resp.TotalCount <= offset+step {
			break
		}
	}
	return list, nil
}
