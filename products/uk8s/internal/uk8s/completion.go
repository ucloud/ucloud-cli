package uk8s

import (
	"slices"
	"strings"

	uhostsdk "github.com/ucloud/ucloud-sdk-go/services/uhost"
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

func listUK8SImageIDs(ctx *cli.Context, projectID, region, zone string) []string {
	client := cli.NewServiceClient(ctx, uk8ssdk.NewClient)
	req := client.NewDescribeUK8SImageRequest()
	req.ProjectId = sdk.String(projectID)
	req.Region = sdk.String(region)
	req.Zone = sdk.String(zone)
	resp, err := client.DescribeUK8SImage(req)
	if err != nil {
		return nil
	}
	out := make([]string, 0, len(resp.ImageSet))
	for _, image := range resp.ImageSet {
		out = append(out, image.ImageId+"/"+strings.ReplaceAll(image.ImageName, " ", "-"))
	}
	return out
}

func listUK8SVersions(ctx *cli.Context, projectID, region string) []string {
	client := cli.NewServiceClient(ctx, uk8ssdk.NewClient)
	req := client.NewGetUK8SVersionsRequest()
	req.ProjectId = sdk.String(projectID)
	req.Region = sdk.String(region)
	req.Kind = sdk.String(defaultUK8SKind)
	resp, err := client.GetUK8SVersions(req)
	if err != nil {
		return nil
	}
	out := make([]string, 0, len(resp.Data))
	for _, version := range resp.Data {
		out = append(out, version.K8sVersion)
	}
	return out
}

func listIsolationGroupIDs(ctx *cli.Context, projectID, region string) []string {
	client := cli.NewServiceClient(ctx, uhostsdk.NewClient)
	req := client.NewDescribeIsolationGroupRequest()
	req.ProjectId = sdk.String(projectID)
	req.Region = sdk.String(region)
	req.Limit = sdk.Int(100)
	resp, err := client.DescribeIsolationGroup(req)
	if err != nil {
		return nil
	}
	out := make([]string, 0, len(resp.IsolationGroupSet))
	for _, group := range resp.IsolationGroupSet {
		out = append(out, group.GroupId+"/"+strings.ReplaceAll(group.GroupName, " ", "-"))
	}
	return out
}

func listNodeGroupIDs(ctx *cli.Context, clusterID, projectID, region string) []string {
	if clusterID == "" {
		return nil
	}
	client := cli.NewServiceClient(ctx, uk8ssdk.NewClient)
	req := client.NewListUK8SNodeGroupRequest()
	req.ClusterId = sdk.String(ctx.PickResourceID(clusterID))
	req.ProjectId = sdk.String(projectID)
	req.Region = sdk.String(region)
	resp, err := client.ListUK8SNodeGroup(req)
	if err != nil {
		return nil
	}
	out := make([]string, 0, len(resp.NodeGroupList))
	for _, group := range resp.NodeGroupList {
		out = append(out, group.NodeGroupId+"/"+strings.ReplaceAll(group.NodeGroupName, " ", "-"))
	}
	return out
}

func listNodeIDs(ctx *cli.Context, clusterID, projectID, region string) []string {
	if clusterID == "" {
		return nil
	}
	client := cli.NewServiceClient(ctx, uk8ssdk.NewClient)
	req := client.NewListUK8SClusterNodeV2Request()
	req.ClusterId = sdk.String(ctx.PickResourceID(clusterID))
	req.ProjectId = sdk.String(projectID)
	req.Region = sdk.String(region)
	resp, err := client.ListUK8SClusterNodeV2(req)
	if err != nil {
		return nil
	}
	out := make([]string, 0, len(resp.NodeSet))
	for _, node := range resp.NodeSet {
		if strings.EqualFold(node.NodeRole, "master") {
			continue
		}
		out = append(out, node.NodeId+"/"+strings.ReplaceAll(node.InstanceName, " ", "-"))
	}
	return out
}
