package ext

import (
	"fmt"

	uhostsdk "github.com/ucloud/ucloud-sdk-go/services/uhost"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

func describeUHostByID(ctx *cli.Context, uhostID, projectID, region, zone string) (*uhostsdk.UHostInstanceSet, error) {
	client := cli.NewServiceClient(ctx, uhostsdk.NewClient)
	req := client.NewDescribeUHostInstanceRequest()
	req.UHostIds = []string{uhostID}
	req.ProjectId = &projectID
	req.Region = &region
	req.Zone = &zone

	resp, err := client.DescribeUHostInstance(req)
	if err != nil {
		return nil, err
	}
	if len(resp.UHostSet) < 1 {
		return nil, fmt.Errorf("uhost [%s] does not exist", uhostID)
	}
	return &resp.UHostSet[0], nil
}
