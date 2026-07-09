package uk8s

import (
	"slices"
	"strings"

	uk8ssdk "github.com/ucloud/ucloud-sdk-go/services/uk8s"
	vpcsdk "github.com/ucloud/ucloud-sdk-go/services/vpc"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

func derefStr(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

// listClusterIDs returns "ClusterId/Name" completion candidates for any
// cluster-id flag. Pass states=nil to include all clusters; otherwise filter
// against ClusterSet[].Status (CLUSTER_RUNNING etc. from status.go).
//
// Completion providers must never fail the shell: any error here is swallowed
// and the provider returns nil (no candidates).
func listClusterIDs(ctx *cli.Context, states []string, region, projectID string) []string {
	client := cli.NewServiceClient(ctx, uk8ssdk.NewClient)
	req := client.NewListUK8SClusterV2Request()
	req.Region = sdk.String(region)
	req.ProjectId = sdk.String(projectID)
	resp, err := client.ListUK8SClusterV2(req)
	if err != nil {
		return nil
	}
	out := make([]string, 0, len(resp.ClusterSet))
	for _, c := range resp.ClusterSet {
		if states != nil && !slices.Contains(states, c.Status) {
			continue
		}
		out = append(out, c.ClusterId+"/"+strings.ReplaceAll(c.ClusterName, " ", "-"))
	}
	return out
}

// listVPCIDs returns "VPCId/Name" candidates. Cross-product resource lookup
// uses the SDK service package directly (per §8 of the platform spec) — never
// import another products/<name>/ tree.
func listVPCIDs(ctx *cli.Context, projectID, region string) []string {
	client := cli.NewServiceClient(ctx, vpcsdk.NewClient)
	req := client.NewDescribeVPCRequest()
	req.ProjectId = sdk.String(projectID)
	req.Region = sdk.String(region)
	resp, err := client.DescribeVPC(req)
	if err != nil {
		return nil
	}
	out := make([]string, 0, len(resp.DataSet))
	for _, v := range resp.DataSet {
		out = append(out, v.VPCId+"/"+strings.ReplaceAll(v.Name, " ", "-"))
	}
	return out
}

// listSubnetIDs returns "SubnetId/Name" candidates for the chosen VPC.
func listSubnetIDs(ctx *cli.Context, vpcID, projectID, region string) []string {
	if vpcID == "" {
		return nil
	}
	client := cli.NewServiceClient(ctx, vpcsdk.NewClient)
	req := client.NewDescribeSubnetRequest()
	req.VPCId = sdk.String(vpcID)
	req.ProjectId = sdk.String(projectID)
	req.Region = sdk.String(region)
	resp, err := client.DescribeSubnet(req)
	if err != nil {
		return nil
	}
	out := make([]string, 0, len(resp.DataSet))
	for _, s := range resp.DataSet {
		out = append(out, s.SubnetId+"/"+strings.ReplaceAll(s.SubnetName, " ", "-"))
	}
	return out
}
