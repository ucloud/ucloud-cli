package uk8s

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	uk8ssdk "github.com/ucloud/ucloud-sdk-go/services/uk8s"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// Table rows are deliberately explicit, matching the UHost commands: table
// output is a stable, compact view, while JSON/YAML keep the full SDK response.
type clusterRow struct {
	ResourceID   string
	Name         string
	Version      string
	Type         string
	Status       string
	MasterCount  int
	NodeCount    int
	VPCID        string
	SubnetID     string
	APIServer    string
	CNIMode      string
	Runtime      string
	CreationTime string
}

type nodeGroupRow struct {
	ResourceID string
	Name       string
	NodeCount  int
	Zone       string
	SubnetID   string
	Config     string
	Image      string
	ChargeType string
	Tag        string
}

type imageRow struct {
	ResourceID string
	Name       string
	Kind       string
	Product    string
	OS         string
	SizeGB     int
	ZoneID     int
	Features   string
}

type nodeRow struct {
	ResourceID    string
	Name          string
	Role          string
	Status        string
	NodeGroup     string
	PrivateIP     string
	Config        string
	Image         string
	MachineType   string
	Zone          string
	InstanceID    string
	Unschedulable bool
	CreationTime  string
}

type versionRow struct {
	K8sVersion        string
	ContainerdVersion string
}

func clusterRows(clusters []uk8ssdk.ClusterSet) []clusterRow {
	rows := make([]clusterRow, 0, len(clusters))
	for _, cluster := range clusters {
		rows = append(rows, clusterRow{
			ResourceID: cluster.ClusterId, Name: cluster.ClusterName,
			Version: cluster.K8sVersion, Type: cluster.ClusterType, Status: cluster.Status,
			MasterCount: cluster.MasterCount, NodeCount: cluster.NodeCount,
			VPCID: cluster.VPCId, SubnetID: cluster.SubnetId, APIServer: cluster.ApiServer,
			CNIMode: cluster.CNIMode, Runtime: formatRuntime(cluster.RuntimeName, cluster.RuntimeVersion),
			CreationTime: formatTimestamp(cluster.CreateTime),
		})
	}
	return rows
}

func nodeGroupRows(groups []uk8ssdk.NodeGroupSet) []nodeGroupRow {
	rows := make([]nodeGroupRow, 0, len(groups))
	for _, group := range groups {
		rows = append(rows, nodeGroupRow{
			ResourceID: group.NodeGroupId, Name: group.NodeGroupName, NodeCount: len(group.NodeList),
			Zone: group.Zone, SubnetID: group.SubnetId,
			Config: fmt.Sprintf("cpu:%d memory:%dMB boot:%s:%dG", group.CPU, group.Mem, group.BootDiskType, group.BootDiskSize),
			Image:  fmt.Sprintf("%s|%s", group.ImageId, group.ImageName), ChargeType: group.ChargeType, Tag: group.Tag,
		})
	}
	return rows
}

func imageRows(response *uk8ssdk.DescribeUK8SImageResponse) []imageRow {
	rows := make([]imageRow, 0)
	appendImages := func(images []uk8ssdk.ImageInfo, kind, product string) {
		for _, image := range images {
			rows = append(rows, imageRow{
				ResourceID: image.ImageId, Name: image.ImageName, Kind: kind, Product: product,
				OS: image.OsName, SizeGB: image.ImageSize, ZoneID: image.ZoneId,
				Features: strings.Join(image.Features, ","),
			})
		}
	}
	appendImages(response.ImageSet, "Base", "UHost")
	appendImages(response.CustomImageSet, "Custom", "UHost")
	appendImages(response.PHostImageSet, "Base", "PHost")
	appendImages(response.CustomPHostImageSet, "Custom", "PHost")
	return rows
}

func nodeRows(nodes []uk8ssdk.NodeInfoV2) []nodeRow {
	rows := make([]nodeRow, 0, len(nodes))
	for _, node := range nodes {
		rows = append(rows, nodeRow{
			ResourceID: node.NodeId, Name: node.InstanceName, Role: node.NodeRole, Status: node.NodeStatus,
			NodeGroup: node.NodeGroupName, PrivateIP: nodePrivateIP(node),
			Config: fmt.Sprintf("cpu:%d memory:%dMB", node.CPU, node.Memory), Image: node.OsName,
			MachineType: node.MachineType, Zone: node.Zone, InstanceID: node.InstanceId,
			Unschedulable: node.Unschedulable, CreationTime: formatTimestamp(node.CreateTime),
		})
	}
	return rows
}

func clusterDescribeRows(cluster *uk8ssdk.DescribeUK8SClusterResponse) []cli.DescribeRow {
	return []cli.DescribeRow{
		{Attribute: "ResourceID", Content: cluster.ClusterId},
		{Attribute: "Name", Content: cluster.ClusterName},
		{Attribute: "Version", Content: cluster.Version},
		{Attribute: "Status", Content: cluster.Status},
		{Attribute: "Type", Content: cluster.ClusterType},
		{Attribute: "APIServer", Content: cluster.ApiServer},
		{Attribute: "ExternalAPIServer", Content: cluster.ExternalApiServer},
		{Attribute: "VPCID", Content: cluster.VPCId},
		{Attribute: "SubnetID", Content: cluster.SubnetId},
		{Attribute: "ServiceCIDR", Content: cluster.ServiceCIDR},
		{Attribute: "PodCIDR", Content: cluster.PodCIDR},
		{Attribute: "NodeCIDR", Content: cluster.NodeCIDR},
		{Attribute: "ClusterDomain", Content: cluster.ClusterDomain},
		{Attribute: "CNIMode", Content: cluster.CNIMode},
		{Attribute: "Runtime", Content: formatRuntime(cluster.RuntimeName, cluster.RuntimeVersion)},
		{Attribute: "MonitorType", Content: cluster.MonitorType},
		{Attribute: "MasterCount", Content: strconv.Itoa(cluster.MasterCount)},
		{Attribute: "NodeCount", Content: strconv.Itoa(cluster.NodeCount)},
		{Attribute: "MasterResourceStatus", Content: cluster.MasterResourceStatus},
		{Attribute: "ExternalUlb", Content: cluster.ExternalUlb},
		{Attribute: "InternalUlb", Content: cluster.InternalUlb},
		{Attribute: "LbClass", Content: cluster.LbClass},
		{Attribute: "DeleteProtection", Content: strconv.Itoa(cluster.DeleteProtection)},
		{Attribute: "EnableUserAuth", Content: strconv.FormatBool(cluster.EnableUserAuth)},
		{Attribute: "DedicatedPodSubnet", Content: strconv.FormatBool(cluster.DedicatedPodSubnet)},
		{Attribute: "CreationTime", Content: formatTimestamp(cluster.CreateTime)},
		{Attribute: "UpdateTime", Content: formatTimestamp(cluster.UpdateTime)},
		{Attribute: "KubeProxyMode", Content: cluster.KubeProxy.Mode},
		{Attribute: "PodSubnetIds", Content: strings.Join(cluster.PodSubnetIds, ",")},
		{Attribute: "PodSubnetSecGroups", Content: strings.Join(cluster.PodSubnetSecGroups, ",")},
		{Attribute: "CACert", Content: cluster.CACert},
		{Attribute: "EtcdCert", Content: cluster.EtcdCert},
		{Attribute: "EtcdKey", Content: cluster.EtcdKey},
	}
}

func nodePrivateIP(node uk8ssdk.NodeInfoV2) string {
	for _, ip := range node.IPSet {
		if ip.Type == "Private" {
			return ip.IP
		}
	}
	return ""
}

func formatTimestamp(value int) string {
	if value <= 0 {
		return ""
	}
	return time.Unix(int64(value), 0).Format(time.RFC3339)
}

func formatRuntime(name, version string) string {
	if name == "" {
		return version
	}
	if version == "" {
		return name
	}
	return name + ":" + version
}
