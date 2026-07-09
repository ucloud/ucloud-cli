package pathx

import (
	"fmt"

	ppathx "github.com/ucloud/ucloud-sdk-go/private/services/pathx"
	pathxsdk "github.com/ucloud/ucloud-sdk-go/services/pathx"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

func getPathxList(ctx *cli.Context, project, region, zone string) []string {
	client := cli.NewServiceClient(ctx, pathxsdk.NewClient)
	req := client.NewDescribeUGA3InstanceRequest()
	req.ProjectId = sdk.String(cli.PickResourceID(project))
	req.Region = sdk.String(region)
	req.Zone = sdk.String(zone)
	req.Limit = sdk.Int(50)
	resp, err := client.DescribeUGA3Instance(req)
	if err != nil {
		ctx.HandleError(err)
		return nil
	}
	list := make([]string, 0)
	for _, item := range resp.ForwardInstanceInfos {
		list = append(list, item.InstanceId)
	}
	return list
}

func getUGAList(ctx *cli.Context, project string) ([]ppathx.UGAAInfo, error) {
	client := cli.NewServiceClient(ctx, ppathx.NewClient)
	req := client.NewDescribeUGAInstanceRequest()
	req.ProjectId = sdk.String(cli.PickResourceID(project))
	resp, err := client.DescribeUGAInstance(req)
	if err != nil {
		return nil, err
	}
	return resp.UGAList, nil
}

func getUGAIDList(ctx *cli.Context, project string) []string {
	list, err := getUGAList(ctx, project)
	if err != nil {
		ctx.LogError(fmt.Sprintf("getUGAIDList failed:%v", err))
		return nil
	}
	strs := make([]string, 0)
	for _, ins := range list {
		strs = append(strs, fmt.Sprintf("%s/%s", ins.UGAId, ins.UGAName))
	}
	return strs
}

func getUpathList(ctx *cli.Context, project string) ([]ppathx.UPathInfo, error) {
	client := cli.NewServiceClient(ctx, ppathx.NewClient)
	req := client.NewDescribeUPathRequest()
	req.ProjectId = sdk.String(cli.PickResourceID(project))
	resp, err := client.DescribeUPath(req)
	if err != nil {
		return nil, err
	}
	return resp.UPathSet, nil
}

func getUpathIDList(ctx *cli.Context, project string) []string {
	list, err := getUpathList(ctx, project)
	if err != nil {
		ctx.LogError(fmt.Sprintf("getUpathIDList failed:%v", err))
		return nil
	}
	strs := make([]string, 0)
	for _, ins := range list {
		strs = append(strs, fmt.Sprintf("%s/%s", ins.UPathId, ins.Name))
	}
	return strs
}
