package vpc

import (
	"fmt"

	vpcsdk "github.com/ucloud/ucloud-sdk-go/services/vpc"

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
