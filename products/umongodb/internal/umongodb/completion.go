package umongodb

import (
	"fmt"
	"strings"

	"github.com/ucloud/ucloud-sdk-go/services/vpc"
	"github.com/ucloud/ucloud-sdk-go/services/umongodb"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// ---------------------------------------------------------------------------
// Generic helpers
// ---------------------------------------------------------------------------

// derefStr safely dereferences a *string bound by a flag, returning "" for nil.
func derefStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// stateAllowed reports whether state passes the optional allow-list. A nil
// allow-list means "any state".
func stateAllowed(state string, states []string) bool {
	if states == nil {
		return true
	}
	for _, s := range states {
		if s == state {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// MongoDB instance ID completion (ListUMongoDBInstances) — GenericInvoke
// SDK has no typed method for ListUMongoDBInstances.
// ---------------------------------------------------------------------------

// getMongoDBIDList returns the resource-id completion candidates for the
// --umongodb-id flag, in the conventional "id/name" form. states restricts
// candidates to those whose State is in the set (nil = any state).
func getMongoDBIDList(ctx *cli.Context, states []string, region, zone, projectID string) []string {
	params := map[string]interface{}{
		"Action": "ListUMongoDBInstances",
		"Region": region,
	}
	if zone != "" {
		params["Zone"] = zone
	}
	if projectID != "" {
		params["ProjectId"] = projectID
	}
	payload, err := genericCall(ctx, "ListUMongoDBInstances", params)
	if err != nil {
		return nil
	}
	dataSet, ok := payload["DataSet"].([]interface{})
	if !ok {
		return nil
	}
	candidates := make([]string, 0, len(dataSet))
	for _, item := range dataSet {
		ins, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		if states != nil {
			// ListUMongoDBInstances has no state filter param;
			// client-side filter if states provided.
			st, _ := ins["State"].(string)
			if !stateAllowed(st, states) {
				continue
			}
		}
		id, _ := ins["ClusterId"].(string)
		name, _ := ins["Name"].(string)
		candidates = append(candidates, fmt.Sprintf("%s/%s", id, name))
	}
	return candidates
}

// ---------------------------------------------------------------------------
// MongoDB version completion (ListUMongoDBVersion) — GenericInvoke
// SDK typed response lacks EngineType and DefaultDBVersion fields.
// ---------------------------------------------------------------------------

// getMongoDBVersionList returns available MongoDB version strings via
// ListUMongoDBVersion.
func getMongoDBVersionList(ctx *cli.Context, region, zone string) []string {
	params := map[string]interface{}{
		"Action": "ListUMongoDBVersion",
		"Region": region,
		"Zone":   zone,
	}
	payload, err := genericCall(ctx, "ListUMongoDBVersion", params)
	if err != nil {
		return nil
	}
	dataSet, ok := payload["DataSet"].([]interface{})
	if !ok {
		return nil
	}
	list := make([]string, 0, len(dataSet))
	for _, item := range dataSet {
		v, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		ver, _ := v["DBVersion"].(string)
		if ver != "" {
			list = append(list, ver)
		}
	}
	return list
}

// ---------------------------------------------------------------------------
// Config template completion and auto-default (ListUMongoDBConfigTemplate)
// Migrated to typed SDK.
// ---------------------------------------------------------------------------

// getDefaultTemplateID fetches the default config template ID for a given
// MongoDB version and cluster type via typed ListUMongoDBConfigTemplate.
func getDefaultTemplateID(ctx *cli.Context, dbVersion, clusterType, project, region string) (string, error) {
	client := cli.NewServiceClient(ctx, umongodb.NewClient)
	req := client.NewListUMongoDBConfigTemplateRequest()
	req.Region = &region
	if project != "" {
		req.ProjectId = &project
	}
	resp, err := client.ListUMongoDBConfigTemplate(req)
	if err != nil {
		return "", err
	}
	if len(resp.DataSet) == 0 {
		return "", fmt.Errorf("no config template found for version %s / type %s in %s", dbVersion, clusterType, region)
	}
	for _, t := range resp.DataSet {
		if t.MongodbVersion == dbVersion && t.ClusterType == clusterType && t.TemplateType == "DefaultTemplate" {
			if t.TemplateId != "" {
				return t.TemplateId, nil
			}
		}
	}
	return "", fmt.Errorf("no default config template found for version %s / type %s in %s", dbVersion, clusterType, region)
}

// getMongoDBTemplateList returns config template candidates as "id/name" strings
// via typed ListUMongoDBConfigTemplate.
func getMongoDBTemplateList(ctx *cli.Context, dbVersion, project, region string) []string {
	client := cli.NewServiceClient(ctx, umongodb.NewClient)
	req := client.NewListUMongoDBConfigTemplateRequest()
	req.Region = &region
	if project != "" {
		req.ProjectId = &project
	}
	resp, err := client.ListUMongoDBConfigTemplate(req)
	if err != nil {
		return nil
	}
	var list []string
	for _, t := range resp.DataSet {
		// Filter by version if specified
		if dbVersion != "" {
			if !strings.EqualFold(t.MongodbVersion, dbVersion) {
				continue
			}
		}
		if t.TemplateId != "" {
			list = append(list, fmt.Sprintf("%s/%s", t.TemplateId, t.TemplateName))
		}
	}
	return list
}

// ---------------------------------------------------------------------------
// Machine spec completion (ListUMongoDBMachineSpec) — GenericInvoke
// SDK has no typed method for ListUMongoDBMachineSpec.
// ---------------------------------------------------------------------------

// getMongoDBMachineSpecList returns machine type candidates from
// ListUMongoDBMachineSpec, flattened from nested ComputeType arrays.
func getMongoDBMachineSpecList(ctx *cli.Context, region, zone string) []string {
	params := map[string]interface{}{
		"Action": "ListUMongoDBMachineSpec",
		"Region": region,
		"Zone":   zone,
	}
	payload, err := genericCall(ctx, "ListUMongoDBMachineSpec", params)
	if err != nil {
		return nil
	}
	dataSet, ok := payload["DataSet"].([]interface{})
	if !ok {
		return nil
	}
	var list []string
	for _, item := range dataSet {
		spec, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		ct, ok := spec["ComputeType"].([]interface{})
		if !ok {
			continue
		}
		for _, c := range ct {
			m, ok := c.(map[string]interface{})
			if !ok {
				continue
			}
			id, _ := m["MachineTypeId"].(string)
			desc, _ := m["Description"].(string)
			if id != "" {
				list = append(list, fmt.Sprintf("%s/%s", id, desc))
			}
		}
	}
	return list
}

// ---------------------------------------------------------------------------
// VPC / Subnet completion (copied per boundary rule, products must be self-contained)
// ---------------------------------------------------------------------------

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
	if err != nil {
		return nil
	}
	list := make([]string, 0, len(vpcInsList))
	for _, v := range vpcInsList {
		list = append(list, fmt.Sprintf("%s/%s", v.VPCId, v.Name))
	}
	return list
}

func getAllSubnets(ctx *cli.Context, vpcID, project, region string) ([]vpc.SubnetInfo, error) {
	client := cli.NewServiceClient(ctx, vpc.NewClient)
	req := client.NewDescribeSubnetRequest()
	req.ProjectId = &project
	req.Region = &region
	if vpcID != "" {
		req.VPCId = &vpcID
	}
	subnets := []vpc.SubnetInfo{}
	for limit, offset := 50, 0; ; offset += limit {
		req.Limit = &limit
		req.Offset = &offset
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
	list := make([]string, 0, len(subnets))
	for _, s := range subnets {
		list = append(list, fmt.Sprintf("%s/%s", s.SubnetId, s.SubnetName))
	}
	return list
}
