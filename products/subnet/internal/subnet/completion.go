package subnet

import (
	"fmt"

	vpcsdk "github.com/ucloud/ucloud-sdk-go/services/vpc"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

func getAllVPCIns(ctx *cli.Context, project, region string) ([]vpcsdk.VPCInfo, error) {
	client := cli.NewServiceClient(ctx, vpcsdk.NewClient)
	req := client.NewDescribeVPCRequest()
	req.ProjectId = &project
	req.Region = &region
	resp, err := client.DescribeVPC(req)
	if err != nil {
		return nil, err
	}
	return resp.DataSet, nil
}

func getAllVPCIdNames(ctx *cli.Context, project, region string) []string {
	vpcInsList, err := getAllVPCIns(ctx, project, region)
	list := []string{}
	if err != nil {
		return nil
	}
	for _, vpc := range vpcInsList {
		list = append(list, fmt.Sprintf("%s/%s", vpc.VPCId, vpc.Name))
	}
	return list
}

func getAllSubnets(ctx *cli.Context, vpcID, project, region string) ([]vpcsdk.SubnetInfo, error) {
	client := cli.NewServiceClient(ctx, vpcsdk.NewClient)
	req := client.NewDescribeSubnetRequest()
	req.ProjectId = sdk.String(cli.PickResourceID(project))
	req.Region = sdk.String(region)
	if vpcID != "" {
		req.VPCId = sdk.String(cli.PickResourceID(vpcID))
	}
	subnets := []vpcsdk.SubnetInfo{}
	for limit, offset := 50, 0; ; offset += limit {
		req.Limit = sdk.Int(limit)
		req.Offset = sdk.Int(offset)
		resp, err := client.DescribeSubnet(req)
		if err != nil {
			ctx.HandleError(err)
			return nil, err
		}
		subnets = append(subnets, resp.DataSet...)
		if limit+offset >= resp.TotalCount {
			break
		}
	}
	return subnets, nil
}

func getAllSubnetIDNames(ctx *cli.Context, vpcID, project, region string) []string {
	subnets, err := getAllSubnets(ctx, vpcID, project, region)
	if err != nil {
		return nil
	}
	list := []string{}
	for _, s := range subnets {
		list = append(list, fmt.Sprintf("%s/%s", s.SubnetId, s.SubnetName))
	}
	return list
}
