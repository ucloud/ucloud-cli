package udpn

import (
	"fmt"

	udpnsdk "github.com/ucloud/ucloud-sdk-go/services/udpn"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

func getAllUDPNIns(ctx *cli.Context, project, region string) ([]udpnsdk.UDPNData, error) {
	client := cli.NewServiceClient(ctx, udpnsdk.NewClient)
	req := client.NewDescribeUDPNRequest()
	req.ProjectId = sdk.String(cli.PickResourceID(project))
	req.Region = sdk.String(region)
	list := make([]udpnsdk.UDPNData, 0)
	for offset, limit := 0, 50; ; offset += limit {
		req.Offset = sdk.Int(offset)
		req.Limit = sdk.Int(limit)
		resp, err := client.DescribeUDPN(req)
		if err != nil {
			return nil, err
		}
		list = append(list, resp.DataSet...)
		if offset+limit > resp.TotalCount {
			break
		}
	}
	return list, nil
}

func getAllUDPNIdNames(ctx *cli.Context, project, region string) []string {
	udpnInsList, err := getAllUDPNIns(ctx, project, region)
	if err != nil {
		return nil
	}
	idNameList := []string{}
	for _, udpn := range udpnInsList {
		idNameList = append(idNameList, fmt.Sprintf("%s/%s:%s", udpn.UDPNId, udpn.Peer1, udpn.Peer2))
	}
	return idNameList
}
