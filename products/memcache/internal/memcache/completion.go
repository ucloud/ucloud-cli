package memcache

import (
	"fmt"

	"github.com/ucloud/ucloud-sdk-go/services/umem"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

func getIDList(ctx *cli.Context, project, region string) []string {
	client := cli.NewServiceClient(ctx, umem.NewClient)
	req := client.NewDescribeUMemcacheGroupRequest()
	req.ProjectId = &project
	req.Region = &region
	list := []string{}

	for limit, offset := 50, 0; ; offset += limit {
		req.Limit = sdk.Int(limit)
		req.Offset = sdk.Int(offset)
		resp, err := client.DescribeUMemcacheGroup(req)
		if err != nil {
			fmt.Fprintln(ctx.ProgressWriter(), err)
			return nil
		}
		for _, ins := range resp.DataSet {
			list = append(list, fmt.Sprintf("%s/%s", ins.GroupId, ins.Name))
		}
		if offset+limit >= resp.TotalCount {
			break
		}
	}
	return list
}
