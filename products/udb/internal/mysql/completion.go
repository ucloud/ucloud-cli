package mysql

import (
	"fmt"

	"github.com/ucloud/ucloud-sdk-go/services/vpc"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// getAllVPCIns mirrors cmd/vpc.go getAllVPCIns, copied here (not imported) so
// the product stays self-contained per the boundary rules.
func getAllVPCIns(ctx *cli.Context, project, region string) ([]vpc.VPCInfo, error) {
	client := cli.NewServiceClient(ctx, vpc.NewClient)
	req := client.NewDescribeVPCRequest()
	req.ProjectId = &project
	req.Region = &region
	resp, err := client.DescribeVPC(req)
	if err != nil {
		return nil, err
	}
	return resp.DataSet, nil
}

// getAllVPCIdNames mirrors cmd/vpc.go getAllVPCIdNames.
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

// getAllSubnets mirrors cmd/vpc.go getAllSubnets.
func getAllSubnets(ctx *cli.Context, vpcID, project, region string) ([]vpc.SubnetInfo, error) {
	client := cli.NewServiceClient(ctx, vpc.NewClient)
	req := client.NewDescribeSubnetRequest()
	req.ProjectId = sdk.String(cli.PickResourceID(project))
	req.Region = sdk.String(region)
	if vpcID != "" {
		req.VPCId = sdk.String(cli.PickResourceID(vpcID))
	}
	subnets := []vpc.SubnetInfo{}
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

// getAllSubnetIDNames mirrors cmd/vpc.go getAllSubnetIDNames.
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
