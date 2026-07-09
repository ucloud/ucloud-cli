package memcache

import (
	"fmt"

	"github.com/ucloud/ucloud-sdk-go/services/vpc"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

func fillDefaultVPCAndSubnet(ctx *cli.Context, vpcID, subnetID *string, projectID, region, zone string) error {
	if *vpcID != "" && *subnetID != "" {
		return nil
	}
	vpcs, err := getAllVPCIns(ctx, projectID, region)
	if err != nil {
		return fmt.Errorf("failed to get vpc list: %s", err)
	}
	if len(vpcs) == 0 {
		return fmt.Errorf("no vpc found in region[%s], please specify --vpc-id and --subnet-id", region)
	}

	var defaultVPC *vpc.VPCInfo
	for i := range vpcs {
		if vpcs[i].VPCType == "DefaultVPC" {
			defaultVPC = &vpcs[i]
			break
		}
	}
	if defaultVPC == nil {
		defaultVPC = &vpcs[0]
	}

	if *vpcID == "" {
		*vpcID = defaultVPC.VPCId
	}

	if *subnetID == "" {
		subnets, err := getAllSubnets(ctx, *vpcID, projectID, region)
		if err != nil {
			return fmt.Errorf("failed to get subnet list: %s", err)
		}
		if len(subnets) == 0 {
			return fmt.Errorf("no subnet found in vpc[%s], please specify --subnet-id", *vpcID)
		}
		if zone != "" {
			for _, sn := range subnets {
				if sn.Zone == zone {
					*subnetID = sn.SubnetId
					return nil
				}
			}
		}
		*subnetID = subnets[0].SubnetId
	}

	return nil
}

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
