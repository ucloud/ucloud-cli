package service

import (
	"fmt"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-sdk-go/services/urocketmq"
	"github.com/ucloud/ucloud-sdk-go/services/vpc"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"
)

// ServiceList returns ServiceIds for the given project/region. Exported for --service-id completion
// reuse by topic/group/token/message (one-way import of service package to avoid circular deps).
func ServiceList(ctx *cli.Context, projectID, region string) []string {
	client := cli.NewServiceClient(ctx, urocketmq.NewClient)
	req := client.NewListURocketMQServiceRequest()
	req.ProjectId = sdk.String(projectID)
	req.Region = sdk.String(region)
	req.Limit = sdk.Int(1000)
	req.Offset = sdk.Int(0)
	resp, err := client.ListURocketMQService(req)
	if err != nil {
		return nil
	}
	ids := make([]string, 0, len(resp.ServiceList))
	for _, s := range resp.ServiceList {
		ids = append(ids, s.ServiceId)
	}
	return ids
}

// getAllVPCIdNames returns "VPCId/Name" completion candidates (--vpc-id completion). See uhost completion.go.
func getAllVPCIdNames(ctx *cli.Context, project, region string) []string {
	client := cli.NewServiceClient(ctx, vpc.NewClient)
	req := client.NewDescribeVPCRequest()
	req.ProjectId = &project
	req.Region = &region
	resp, err := client.DescribeVPC(req)
	if err != nil {
		return nil
	}
	list := make([]string, 0, len(resp.DataSet))
	for _, v := range resp.DataSet {
		list = append(list, fmt.Sprintf("%s/%s", v.VPCId, v.Name))
	}
	return list
}

// getAllSubnetIDNames returns "SubnetId/Name" completion candidates (--subnet-id completion). See uhost completion.go.
func getAllSubnetIDNames(ctx *cli.Context, vpcID, project, region string) []string {
	client := cli.NewServiceClient(ctx, vpc.NewClient)
	req := client.NewDescribeSubnetRequest()
	req.ProjectId = sdk.String(project)
	req.Region = sdk.String(region)
	if vpcID != "" {
		req.VPCId = sdk.String(cli.PickResourceID(vpcID))
	}
	subnets := make([]vpc.SubnetInfo, 0)
	for limit, offset := 50, 0; ; offset += limit {
		req.Limit = sdk.Int(limit)
		req.Offset = sdk.Int(offset)
		resp, err := client.DescribeSubnet(req)
		if err != nil {
			return nil
		}
		subnets = append(subnets, resp.DataSet...)
		if limit+offset >= resp.TotalCount {
			break
		}
	}
	list := make([]string, 0, len(subnets))
	for _, s := range subnets {
		list = append(list, fmt.Sprintf("%s/%s", s.SubnetId, s.SubnetName))
	}
	return list
}
