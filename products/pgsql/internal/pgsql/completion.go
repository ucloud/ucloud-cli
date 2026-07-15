package pgsql

import (
	"fmt"

	"github.com/ucloud/ucloud-sdk-go/services/upgsql"
	"github.com/ucloud/ucloud-sdk-go/services/vpc"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

var pgsqlVersionList = []string{"postgresql-10.4", "postgresql-13.4"}

// getAllVPCIns mirrors products/mysql/internal/mysql/completion.go getAllVPCIns,
// copied here (not imported) so the product stays self-contained per the
// boundary rules (hack/check-product rule1).
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

// getAllVPCIdNames mirrors products/mysql/internal/mysql/completion.go getAllVPCIdNames.
func getAllVPCIdNames(ctx *cli.Context, project, region string) []string {
	vpcInsList, err := getAllVPCIns(ctx, project, region)
	list := []string{}
	if err != nil {
		return nil
	}
	for _, v := range vpcInsList {
		list = append(list, fmt.Sprintf("%s/%s", v.VPCId, v.Name))
	}
	return list
}

// getAllSubnets mirrors products/mysql/internal/mysql/completion.go getAllSubnets.
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

// getAllSubnetIDNames mirrors products/mysql/internal/mysql/completion.go getAllSubnetIDNames.
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

// getUPgSQLList returns all UPgSQL instances for the given project/region/zone.
// ListUPgSQLInstance has no Limit/Offset, so a single call returns the full set.
func getUPgSQLList(ctx *cli.Context, project, region, zone string) ([]upgsql.UDBInstanceSet, error) {
	client := newUPgSQLClient(ctx)
	req := client.NewListUPgSQLInstanceRequest()
	req.ProjectId = sdk.String(project)
	req.Region = sdk.String(region)
	req.Zone = sdk.String(zone)
	resp, err := client.ListUPgSQLInstance(req)
	if err != nil {
		return nil, err
	}
	return resp.DataSet, nil
}

// getUPgSQLIDList returns "InstanceID/Name" completion candidates for --instance-id.
func getUPgSQLIDList(ctx *cli.Context, project, region, zone string) []string {
	instances, err := getUPgSQLList(ctx, project, region, zone)
	if err != nil {
		return nil
	}
	list := []string{}
	for _, ins := range instances {
		list = append(list, fmt.Sprintf("%s/%s", ins.InstanceID, ins.Name))
	}
	return list
}

// listParamTemplates returns the available UPgSQL param templates for the given
// project/region/zone, paginating via item-count (the API exposes no TotalCount).
func listParamTemplates(ctx *cli.Context, project, region, zone string) ([]upgsql.TemplateGroup, error) {
	client := newUPgSQLClient(ctx)
	req := client.NewListUPgSQLParamTemplateRequest()
	req.ProjectId = sdk.String(project)
	req.Region = sdk.String(region)
	req.Zone = sdk.String(zone)
	list := []upgsql.TemplateGroup{}

	resp, err := client.ListUPgSQLParamTemplate(req)
	if err != nil {
		return nil, err
	}
	list = resp.Data
	return list, nil
}

// getDefaultParamGroupID picks a param template for the given DB version via
// ListUPgSQLParamTemplate. The API has no DBVersion filter, so we match by the
// DBVersion field in the returned templates and fall back to the first one.
func getDefaultParamGroupID(ctx *cli.Context, dbVersion, project, region, zone string) (int, error) {
	templates, err := listParamTemplates(ctx, project, region, zone)
	if err != nil {
		return 0, fmt.Errorf("call ListUPgSQLParamTemplate: %w", err)
	}
	if len(templates) == 0 {
		return 0, fmt.Errorf("no param template found in %s/%s", region, zone)
	}
	for _, t := range templates {
		if t.DBVersion == dbVersion {
			return t.GroupID, nil
		}
	}
	return templates[0].GroupID, nil
}

// listParamTemplateIDNames returns "GroupID/GroupName" candidates for --param-group-id.
func listParamTemplateIDNames(ctx *cli.Context, project, region, zone string) []string {
	templates, err := listParamTemplates(ctx, project, region, zone)
	if err != nil {
		return nil
	}
	list := []string{}
	for _, t := range templates {
		list = append(list, fmt.Sprintf("%d/%s", t.GroupID, t.GroupName))
	}
	return list
}

// listMachineTypeIDNames returns "ID/Description" candidates for --machine-type.
func listMachineTypeIDNames(ctx *cli.Context, project, region, zone string) []string {
	client := newUPgSQLClient(ctx)
	req := client.NewListUPgSQLMachineTypeRequest()
	req.ProjectId = sdk.String(project)
	req.Region = sdk.String(region)
	req.Zone = sdk.String(zone)
	resp, err := client.ListUPgSQLMachineType(req)
	if err != nil {
		return nil
	}
	list := []string{}
	for _, mt := range resp.DataSet {
		list = append(list, fmt.Sprintf("%s/%s", mt.ID, mt.Description))
	}
	return list
}

// getBackupIDList returns "BackupID/BackupName" candidates for --backup-id,
// paginating via TotalCount (ListUPgSQLBackup exposes TotalCount).
func getBackupIDList(ctx *cli.Context, instanceID, project, region, zone string) []string {
	client := newUPgSQLClient(ctx)
	req := client.NewListUPgSQLBackupRequest()
	req.ProjectId = sdk.String(project)
	req.Region = sdk.String(region)
	req.Zone = sdk.String(zone)
	req.InstanceID = sdk.String(instanceID)
	list := []string{}
	for limit, offset := 100, 0; ; offset += limit {
		req.Limit = sdk.Int(limit)
		req.Offset = sdk.Int(offset)
		resp, err := client.ListUPgSQLBackup(req)
		if err != nil {
			return nil
		}
		for _, b := range resp.DataSet {
			list = append(list, fmt.Sprintf("%s/%s", b.BackupID, b.BackupName))
		}
		if offset+limit >= resp.TotalCount {
			break
		}
	}
	return list
}
