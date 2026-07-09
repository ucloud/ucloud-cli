package mysql

import (
	"fmt"

	"github.com/ucloud/ucloud-sdk-go/services/udb"
	"github.com/ucloud/ucloud-sdk-go/services/vpc"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

var dbVersionList = []string{"mysql-5.7", "mysql-8.0", "mysql-8.4", "percona-5.7"}

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

func getUDBIDList(ctx *cli.Context, states []string, dbType, project, region, zone string) []string {
	udbs, err := getUDBList(ctx, states, dbType, project, region, zone)
	if err != nil {
		return nil
	}
	list := []string{}
	for _, db := range udbs {
		list = append(list, fmt.Sprintf("%s/%s", db.DBId, db.Name))
	}
	return list
}

func getUDBList(ctx *cli.Context, states []string, dbType, project, region, zone string) ([]udb.UDBInstanceSet, error) {
	client := cli.NewServiceClient(ctx, udb.NewClient)
	req := client.NewDescribeUDBInstanceRequest()
	if dbType == "" {
		dbType = "sql"
	}
	req.ClassType = &dbType
	req.ProjectId = &project
	req.Region = &region
	req.Zone = &zone
	list := []udb.UDBInstanceSet{}
	for offset, limit := 0, 50; ; offset += limit {
		req.Offset = sdk.Int(offset)
		req.Limit = sdk.Int(limit)
		resp, err := client.DescribeUDBInstance(req)
		if err != nil {
			return nil, err
		}
		for _, ins := range resp.DataSet {
			if states != nil {
				for _, s := range states {
					if s == ins.State {
						list = append(list, ins)
					}
				}
			} else {
				list = append(list, ins)
			}
		}
		if offset+limit >= resp.TotalCount {
			break
		}
	}
	return list, nil
}

func getConfList(ctx *cli.Context, dbType, project, region, zone string) ([]udb.UDBParamGroupSet, error) {
	client := cli.NewServiceClient(ctx, udb.NewClient)
	req := client.NewDescribeUDBParamGroupRequest()
	req.ClassType = &dbType
	req.ProjectId = &project
	req.Region = &region
	req.Zone = &zone
	list := []udb.UDBParamGroupSet{}
	for offset, limit := 0, 50; ; offset += limit {
		req.Offset = sdk.Int(offset)
		req.Limit = sdk.Int(limit)
		resp, err := client.DescribeUDBParamGroup(req)
		if err != nil {
			return nil, err
		}
		for _, conf := range resp.DataSet {
			list = append(list, conf)
		}
		if resp.TotalCount <= offset+limit {
			break
		}
	}
	return list, nil
}

func getModifiableConfIDList(ctx *cli.Context, dbType, project, region, zone string) []string {
	confs, err := getConfList(ctx, dbType, project, region, zone)
	if err != nil {
		return nil
	}
	list := []string{}
	for _, conf := range confs {
		if conf.Modifiable == true {
			list = append(list, fmt.Sprintf("%d/%s", conf.GroupId, conf.GroupName))
		}
	}
	return list
}

func getConfIDList(ctx *cli.Context, dbType, project, region, zone string) []string {
	confs, err := getConfList(ctx, dbType, project, region, zone)
	if err != nil {
		return nil
	}
	list := []string{}
	for _, conf := range confs {
		list = append(list, fmt.Sprintf("%d/%s", conf.GroupId, conf.GroupName))
	}
	return list
}
