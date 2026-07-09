package ulb

import (
	"fmt"

	ulbsdk "github.com/ucloud/ucloud-sdk-go/services/ulb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

func getAllULB(ctx *cli.Context, project, region string) ([]ulbsdk.ULBSet, error) {
	list := []ulbsdk.ULBSet{}
	client := cli.NewServiceClient(ctx, ulbsdk.NewClient)
	req := client.NewDescribeULBRequest()
	req.ProjectId = &project
	req.Region = &region

	for offset, limit := 0, 50; ; offset += limit {
		req.Offset = sdk.Int(offset)
		req.Limit = sdk.Int(limit)
		resp, err := client.DescribeULB(req)
		if err != nil {
			return nil, err
		}
		list = append(list, resp.DataSet...)

		if resp.TotalCount < offset+limit {
			break
		}
	}
	return list, nil
}

func getAllULBIDNames(ctx *cli.Context, project, region string) []string {
	list := []string{}
	ulbList, err := getAllULB(ctx, project, region)
	if err != nil {
		return nil
	}
	for _, ulb := range ulbList {
		list = append(list, fmt.Sprintf("%s/%s", ulb.ULBId, ulb.Name))
	}
	return list
}

func getAllVServers(ctx *cli.Context, ulbID, vserverID, project, region string) ([]ulbsdk.ULBVServerSet, error) {
	client := cli.NewServiceClient(ctx, ulbsdk.NewClient)
	req := client.NewDescribeVServerRequest()
	req.ULBId = sdk.String(cli.PickResourceID(ulbID))
	req.ProjectId = sdk.String(cli.PickResourceID(project))
	req.Region = &region
	if vserverID != "" {
		req.VServerId = sdk.String(cli.PickResourceID(vserverID))
	}
	resp, err := client.DescribeVServer(req)
	if err != nil {
		return nil, err
	}
	if vserverID != "" {
		if len(resp.DataSet) < 1 {
			return nil, fmt.Errorf("VServer[%s] may not exist", vserverID)
		} else if len(resp.DataSet) > 1 {
			return nil, fmt.Errorf("Internal Error, too many vserver:%#v", resp.DataSet)
		}
	}
	return resp.DataSet, nil
}

func getAllVServerIDNames(ctx *cli.Context, ulbID, project, region string) []string {
	vservers, err := getAllVServers(ctx, ulbID, "", project, region)
	if err != nil {
		return nil
	}
	idNames := []string{}
	for _, vs := range vservers {
		idNames = append(idNames, fmt.Sprintf("%s/%s", vs.VServerId, vs.VServerName))
	}
	return idNames
}

func getAllBackendNodes(ctx *cli.Context, ulbID, vserverID, project, region string) ([]ulbsdk.ULBBackendSet, error) {
	vsList, err := getAllVServers(ctx, ulbID, vserverID, project, region)
	if err != nil {
		return nil, err
	}
	nodeList := []ulbsdk.ULBBackendSet{}
	for _, vs := range vsList {
		nodeList = append(nodeList, vs.BackendSet...)
	}
	return nodeList, nil
}

func getAllBackendNodeIDNames(ctx *cli.Context, ulbID, vserverID, project, region string) []string {
	nodeList, err := getAllBackendNodes(ctx, ulbID, vserverID, project, region)
	if err != nil {
		return nil
	}
	idNames := []string{}
	for _, node := range nodeList {
		idNames = append(idNames, fmt.Sprintf("%s/%s", node.BackendId, node.ResourceName))
	}
	return idNames
}

func getAllSSLCertIDNames(ctx *cli.Context, project, region string) []string {
	sslcs, err := getAllSSLCerts(ctx, project, region)
	if err != nil {
		return nil
	}
	idNames := []string{}
	for _, ssl := range sslcs {
		idNames = append(idNames, fmt.Sprintf("%s/%s", ssl.SSLId, ssl.SSLName))
	}
	return idNames
}

func getAllSSLCerts(ctx *cli.Context, project, region string) ([]ulbsdk.ULBSSLSet, error) {
	client := cli.NewServiceClient(ctx, ulbsdk.NewClient)
	req := client.NewDescribeSSLRequest()
	req.ProjectId = sdk.String(cli.PickResourceID(project))
	req.Region = sdk.String(region)
	list := []ulbsdk.ULBSSLSet{}
	for offset, limit := 0, 50; ; offset += limit {
		req.Offset = sdk.Int(offset)
		req.Limit = sdk.Int(limit)
		resp, err := client.DescribeSSL(req)
		if err != nil {
			return nil, err
		}
		list = append(list, resp.DataSet...)
		if resp.TotalCount <= offset+limit {
			break
		}
	}
	return list, nil
}

func getSSLCertByID(ctx *cli.Context, sslID, project, region string) (*ulbsdk.ULBSSLSet, error) {
	if sslID == "" {
		return nil, fmt.Errorf("ssl certificate resource id can't be empty")
	}
	client := cli.NewServiceClient(ctx, ulbsdk.NewClient)
	req := client.NewDescribeSSLRequest()
	req.ProjectId = sdk.String(cli.PickResourceID(project))
	req.Region = sdk.String(region)
	req.SSLId = sdk.String(cli.PickResourceID(sslID))
	resp, err := client.DescribeSSL(req)
	if err != nil {
		return nil, err
	}
	if len(resp.DataSet) <= 0 {
		return nil, fmt.Errorf("ssl certificate[%s] is not exists", sslID)
	}
	return &resp.DataSet[0], nil
}
