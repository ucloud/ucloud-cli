package ugn

import (
	"fmt"

	ugnsdk "github.com/ucloud/ucloud-sdk-go/services/ugn"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

func getAllUGNIns(ctx *cli.Context, project string) ([]ugnsdk.UGN, error) {
	client := cli.NewServiceClient(ctx, ugnsdk.NewClient)
	req := client.NewListUGNRequest()
	req.ProjectId = sdk.String(cli.PickResourceID(project))
	list := make([]ugnsdk.UGN, 0)
	for offset, limit := 0, 50; offset <= 1000; offset += limit {
		req.Limit = sdk.Int(limit)
		req.Offset = sdk.Int(offset)
		resp, err := client.ListUGN(req)
		if err != nil {
			return nil, err
		}
		list = append(list, resp.UGNs...)
		if offset+limit >= resp.TotalCount {
			break
		}
	}
	return list, nil
}

func getAllUGNIdNames(ctx *cli.Context, project string) []string {
	ugnInsList, err := getAllUGNIns(ctx, project)
	if err != nil {
		return nil
	}
	idNameList := []string{}
	for _, ugn := range ugnInsList {
		idNameList = append(idNameList, fmt.Sprintf("%s/%s", ugn.UGNID, ugn.Name))
	}
	return idNameList
}
