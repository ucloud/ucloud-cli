package cmd

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/vpc"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/base"
)

//NewCmdVpc ucloud vpc
func NewCmdVpc() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vpc",
		Short: "List and manipulate vpc instances",
		Long:  "List and manipulate vpc instances",
		Args:  cobra.NoArgs,
	}

	cmd.AddCommand(NewCmdVpcCreate())
	cmd.AddCommand(NewCmdVPCList())
	cmd.AddCommand(NewCmdVpcDelete())
	cmd.AddCommand(NewCmdVpcCreatePeer())
	cmd.AddCommand(NewCmdVpcListPeer())
	cmd.AddCommand(NewCmdVpcDeletePeer())
	return cmd
}

//VPCRow 表格行
type VPCRow struct {
	VPCName        string
	ResourceID     string
	Group          string
	NetworkSegment string
	SubnetCount    int
	CreationTime   string
}

//NewCmdVPCList ucloud vpc list
func NewCmdVPCList() *cobra.Command {
	vpcIDs := []string{}
	req := base.BizClient.NewDescribeVPCRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List vpc",
		Long:  "List vpc",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			for _, id := range vpcIDs {
				req.VPCIds = append(req.VPCIds, base.PickResourceID(id))
			}
			resp, err := base.BizClient.DescribeVPC(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			list := []VPCRow{}
			for _, vpc := range resp.DataSet {
				row := VPCRow{}
				row.VPCName = vpc.Name
				row.ResourceID = vpc.VPCId
				row.Group = vpc.Tag
				row.NetworkSegment = strings.Join(vpc.Network, ",")
				row.SubnetCount = vpc.SubnetCount
				row.CreationTime = base.FormatDate(vpc.CreateTime)
				list = append(list, row)
			}
			if global.json {
				base.PrintJSON(list)
			} else {
				base.PrintTableS(list)
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.Region = flags.String("region", base.ConfigIns.Region, "Optional. Region, see 'ucloud region'")
	req.ProjectId = flags.String("project-id", base.ConfigIns.ProjectID, "Optional. Project-id, see 'ucloud project list'")
	req.Tag = flags.String("group", "", "Optional. Group")
	flags.StringSliceVar(&vpcIDs, "vpc-id", []string{}, "Optional. Multiple values separated by commas")

	flags.SetFlagValuesFunc("vpc-id", func() []string {
		return getAllVPCIdNames(*req.ProjectId, *req.Region)
	})

	return cmd
}

//NewCmdVpcCreate ucloud vpc create
func NewCmdVpcCreate() *cobra.Command {
	var segments *[]string
	req := base.BizClient.NewCreateVPCRequest()
	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Create vpc network",
		Long:    "Create vpc network",
		Example: "ucloud vpc create --name xxx --segment 192.168.0.0/16",
		Args:    cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			req.Network = *segments
			resp, err := base.BizClient.CreateVPC(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			base.Cxt.Printf("vpc[%s] created\n", resp.VPCId)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	req.Name = cmd.Flags().String("name", "", "Required. Name of the vpc network.")
	segments = cmd.Flags().StringSlice("segment", nil, "Required. The segment for private network.")
	req.Tag = cmd.Flags().String("group", "", "Optional. Business group.")
	req.Remark = cmd.Flags().String("remark", "", "Optional. The description of the vpc.")
	req.Region = cmd.Flags().String("region", base.ConfigIns.Region, "Optional. Assign the region of the VPC")
	req.ProjectId = cmd.Flags().String("project-id", base.ConfigIns.ProjectID, "Optional. Assign the project-id")

	flags.SetFlagValuesFunc("vpc-id", func() []string {
		return getAllVPCIdNames(*req.ProjectId, *req.Region)
	})

	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("segment")

	return cmd
}

//NewCmdVpcDelete ucloud vpc delete
func NewCmdVpcDelete() *cobra.Command {
	idNames := []string{}
	req := base.BizClient.NewDeleteVPCRequest()
	cmd := &cobra.Command{
		Use:     "delete",
		Short:   "Delete vpc network",
		Long:    "Delete vpc network",
		Example: "ucloud vpc delete --vpc-id uvnet-xxx",
		Run: func(cmd *cobra.Command, args []string) {
			for _, idname := range idNames {
				req.VPCId = sdk.String(base.PickResourceID(idname))
				_, err := base.BizClient.DeleteVPC(req)
				if err != nil {
					base.HandleError(err)
					return
				}
				base.Cxt.Printf("vpc[%s] deleted\n", idname)
			}
		},
	}

	cmd.Flags().SortFlags = false

	cmd.Flags().StringSliceVar(&idNames, "vpc-id", nil, "Required. Resource ID of the vpc network to delete")
	req.Region = cmd.Flags().String("region", base.ConfigIns.Region, "Optional. Region of the vpc")
	req.ProjectId = cmd.Flags().String("project-id", base.ConfigIns.ProjectID, "Optional. Project id of the vpc")

	cmd.Flags().SetFlagValuesFunc("vpc-id", func() []string {
		return getAllVPCIdNames(*req.ProjectId, *req.Region)
	})

	cmd.MarkFlagRequired("vpc-id")

	return cmd
}

//NewCmdVpcCreatePeer ucloud vpc peer
func NewCmdVpcCreatePeer() *cobra.Command {
	req := base.BizClient.NewCreateVPCIntercomRequest()
	cmd := &cobra.Command{
		Use:     "create-intercome",
		Short:   "Create intercome with other vpc",
		Long:    "Create intercome with other vpc",
		Example: "ucloud vpc create-intercome --vpc-id xx --dst-vpc-id xx --dst-region xx",
		Run: func(cmd *cobra.Command, args []string) {
			req.ProjectId = sdk.String(base.PickResourceID(*req.ProjectId))
			req.DstProjectId = sdk.String(base.PickResourceID(*req.DstProjectId))
			req.VPCId = sdk.String(base.PickResourceID(*req.VPCId))
			req.DstVPCId = sdk.String(base.PickResourceID(*req.DstVPCId))
			_, err := base.BizClient.CreateVPCIntercom(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			base.Cxt.Printf("intercome [%s<-->%s] establish", *req.VPCId, *req.DstVPCId)
		},
	}

	cmd.Flags().SortFlags = false

	req.VPCId = cmd.Flags().String("vpc-id", "", "Required. The source vpc you want to establish the intercome")
	req.DstVPCId = cmd.Flags().String("dst-vpc-id", "", "Required. The target vpc you want to establish the intercome")
	req.DstRegion = cmd.Flags().String("dst-region", base.ConfigIns.Region, "Required. If the intercome established across different regions")
	req.Region = cmd.Flags().String("region", base.ConfigIns.Region, "Optioanl. The region of source vpc which will establish the intercome")
	req.ProjectId = cmd.Flags().String("project-id", base.ConfigIns.ProjectID, "Optional. The project id of the source vpc")
	req.DstProjectId = cmd.Flags().String("dst-project-id", base.ConfigIns.ProjectID, "Optional. The project id of the source vpc")

	cmd.MarkFlagRequired("vpc-id")
	cmd.MarkFlagRequired("dst-vpc-id")

	cmd.Flags().SetFlagValuesFunc("vpc-id", func() []string {
		return getAllVPCIdNames(*req.ProjectId, *req.Region)
	})
	cmd.Flags().SetFlagValuesFunc("dst-vpc-id", func() []string {
		return getAllVPCIdNames(*req.DstProjectId, *req.DstRegion)
	})
	cmd.Flags().SetFlagValuesFunc("region", getRegionList)
	cmd.Flags().SetFlagValuesFunc("dst-region", getRegionList)
	cmd.Flags().SetFlagValuesFunc("project-id", getProjectList)
	cmd.Flags().SetFlagValuesFunc("dst-project-id", getProjectList)

	return cmd
}

//VPCIntercomRow 表格行
type VPCIntercomRow struct {
	VPCName    string
	ResourceID string
	Segments   string
	ProjectID  string
	DstRegion  string
	Group      string
}

//NewCmdVpcListPeer ucloud vpc list-intercome
func NewCmdVpcListPeer() *cobra.Command {
	req := base.BizClient.NewDescribeVPCIntercomRequest()
	cmd := &cobra.Command{
		Use:     "list-intercome",
		Short:   "list intercome ",
		Long:    "list intercome",
		Example: "ucloud vpc list-intercome --vpc-id xx",
		Run: func(cmd *cobra.Command, args []string) {
			req.VPCId = sdk.String(base.PickResourceID(*req.VPCId))
			resp, err := base.BizClient.DescribeVPCIntercom(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			list := make([]VPCIntercomRow, 0)
			for _, VPCIntercom := range resp.DataSet {
				row := VPCIntercomRow{}
				row.ProjectID = VPCIntercom.ProjectId
				row.Segments = strings.Join(VPCIntercom.Network, ",")
				row.DstRegion = VPCIntercom.DstRegion
				row.VPCName = VPCIntercom.Name
				row.ResourceID = VPCIntercom.VPCId
				row.Group = VPCIntercom.Tag
				list = append(list, row)
			}
			if global.json {
				base.PrintJSON(list)
			} else {
				base.PrintTableS(list)
			}
		},
	}
	req.VPCId = cmd.Flags().String("vpc-id", "", "Required. The vpc id which you wnat to describe the information")
	req.ProjectId = cmd.Flags().String("project-id", base.ConfigIns.ProjectID, "Optional. The project id of source vpc")
	req.Region = cmd.Flags().String("region", base.ConfigIns.Region, "Optional, The region of source vpc")

	cmd.Flags().SetFlagValuesFunc("vpc-id", func() []string {
		return getAllVPCIdNames(*req.ProjectId, *req.Region)
	})
	cmd.Flags().SetFlagValuesFunc("region", getRegionList)
	cmd.Flags().SetFlagValuesFunc("project-id", getProjectList)

	cmd.MarkFlagRequired("vpc-id")

	return cmd
}

//NewCmdVpcDeletePeer ucloud vpc delete-intercome
func NewCmdVpcDeletePeer() *cobra.Command {
	req := base.BizClient.NewDeleteVPCIntercomRequest()
	cmd := &cobra.Command{
		Use:     "delete-intercome",
		Short:   "delete the vpc intercome",
		Long:    "delete the vpc intercome",
		Example: "ucloud vpc delete-intercome --vpc-id xxx --dst-vpc-id xxx",
		Run: func(cmd *cobra.Command, args []string) {
			req.VPCId = sdk.String(base.PickResourceID(*req.VPCId))
			req.DstVPCId = sdk.String(base.PickResourceID(*req.DstVPCId))
			_, err := base.BizClient.DeleteVPCIntercom(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			base.Cxt.Printf("intercome [%s<-->%s] deleted\n", *req.VPCId, *req.DstVPCId)
		},
	}

	cmd.Flags().SortFlags = false

	req.VPCId = cmd.Flags().String("vpc-id", "", "Required. Resource ID of source VPC to disconnect with destination VPC")
	req.DstVPCId = cmd.Flags().String("dst-vpc-id", "", "Required. Resource ID of destination VPC to disconnect with source VPC")
	req.ProjectId = cmd.Flags().String("project-id", base.ConfigIns.ProjectID, "Optional. The project id of source vpc")
	req.Region = cmd.Flags().String("region", base.ConfigIns.Region, "Optional. The region of source vpc to disconnect")
	req.DstRegion = cmd.Flags().String("dst-region", "", "Optional. The region of dest vpc to disconnect")

	cmd.MarkFlagRequired("vpc-id")
	cmd.MarkFlagRequired("dst-vpc-id")
	cmd.MarkFlagRequired("dst-region")

	cmd.Flags().SetFlagValuesFunc("vpc-id", func() []string {
		return getAllVPCIdNames(*req.ProjectId, *req.Region)
	})
	cmd.Flags().SetFlagValuesFunc("dst-region", getRegionList)

	return cmd
}

func getAllVPCIns(project, region string) ([]vpc.VPCInfo, error) {
	req := base.BizClient.NewDescribeVPCRequest()
	req.ProjectId = &project
	req.Region = &region
	resp, err := base.BizClient.DescribeVPC(req)
	if err != nil {
		return nil, err
	}
	return resp.DataSet, nil
}

func getAllVPCIdNames(project, region string) []string {
	vpcInsList, err := getAllVPCIns(project, region)
	list := []string{}
	if err != nil {
		return nil
	}
	for _, vpc := range vpcInsList {
		list = append(list, fmt.Sprintf("%s/%s", vpc.VPCId, vpc.Name))
	}
	return list
}

//NewCmdSubnet  ucloud subnet
func NewCmdSubnet() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "subnet",
		Short: "List, create and delete subnet",
		Long:  "List, create and delete subnet",
		Args:  cobra.NoArgs,
	}
	out := base.Cxt.GetWriter()
	cmd.AddCommand(NewCmdSubnetList())
	cmd.AddCommand(NewCmdSubnetCreate())
	cmd.AddCommand(NewCmdSubnetDelete(out))
	cmd.AddCommand(NewCmdSubnetListResource(out))

	return cmd
}

//SubnetRow 表格行
type SubnetRow struct {
	SubnetName     string
	ResourceID     string
	Group          string
	AffiliatedVPC  string
	NetworkSegment string
	CreationTime   string
}

//NewCmdSubnetList ucloud subnet list
func NewCmdSubnetList() *cobra.Command {
	req := base.BizClient.NewDescribeSubnetRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List subnet",
		Long:  `List subnet`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := base.BizClient.DescribeSubnet(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			list := make([]SubnetRow, 0)
			for _, sn := range resp.DataSet {
				row := SubnetRow{}
				row.SubnetName = sn.SubnetName
				row.ResourceID = sn.SubnetId
				row.Group = sn.Tag
				row.AffiliatedVPC = fmt.Sprintf("%s/%s", sn.VPCId, sn.VPCName)
				row.NetworkSegment = fmt.Sprintf("%s/%s", sn.Subnet, sn.Netmask)
				row.CreationTime = base.FormatDate(sn.CreateTime)
				list = append(list, row)
			}
			if global.json {
				base.PrintJSON(list)
			} else {
				base.PrintTableS(list)
			}
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	req.Region = flags.String("region", base.ConfigIns.Region, "Optional. Region, see 'ucloud region'")
	req.ProjectId = flags.String("project-id", base.ConfigIns.ProjectID, "Optional. Project-id, see 'ucloud project list'")
	flags.StringSliceVar(&req.SubnetIds, "subnet-id", []string{}, "Optional. Multiple values separated by commas")
	req.VPCId = flags.String("vpc-id", "", "Optional. Resource ID of VPC")
	req.Tag = flags.String("group", "", "Optional. Group")
	req.Offset = flags.Int("offset", 0, "Optional. Offset")
	req.Limit = flags.Int("limit", 50, "Optional. Limit")

	return cmd
}

//NewCmdSubnetCreate  ucloud subnet create
func NewCmdSubnetCreate() *cobra.Command {
	var segment *net.IPNet
	req := base.BizClient.NewCreateSubnetRequest()
	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Create subnet of vpc network",
		Long:    "Create subnet of vpc network",
		Example: "ucloud subnet create --vpc-id uvnet-vpcxid --name testName --segment 192.168.2.0/24",
		Run: func(cmd *cobra.Command, args []string) {
			ipMaskStrs := strings.SplitN(segment.String(), "/", 2)
			req.Subnet = sdk.String(ipMaskStrs[0])
			mask, err := strconv.Atoi(ipMaskStrs[1])
			if err != nil {
				base.HandleError(err)
				return
			}
			req.Netmask = sdk.Int(mask)
			req.VPCId = sdk.String(base.PickResourceID(*req.VPCId))
			resp, err := base.BizClient.CreateSubnet(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			base.Cxt.Printf("subnet[%s] created\n", resp.SubnetId)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	req.VPCId = cmd.Flags().String("vpc-id", "", "Required. Assign the VPC network of the subnet")
	segment = cmd.Flags().IPNet("segment", net.IPNet{}, "Required. Segment of subnet. For example '192.168.0.0/24'")
	req.SubnetName = cmd.Flags().String("name", "Subnet", "Optional. Name of subnet to create")
	req.Region = cmd.Flags().String("region", base.ConfigIns.Region, "Optional. The region of the subnet")
	req.ProjectId = cmd.Flags().String("project-id", base.ConfigIns.ProjectID, "Optional. The project id of the subnet")
	req.Tag = cmd.Flags().String("group", "", "Optional. Business group")
	req.Remark = cmd.Flags().String("remark", "", "Optional. Remark of subnet to create")

	cmd.Flags().SetFlagValuesFunc("vpc-id", func() []string {
		return getAllVPCIdNames(*req.ProjectId, *req.Region)
	})

	cmd.MarkFlagRequired("vpc-id")
	cmd.MarkFlagRequired("segment")

	return cmd
}

//NewCmdSubnetDelete ucloud subnet delete
func NewCmdSubnetDelete(out io.Writer) *cobra.Command {
	idNames := []string{}
	req := base.BizClient.NewDeleteSubnetRequest()
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete subnet",
		Long:  "Delete subnet",
		Run: func(c *cobra.Command, args []string) {
			req.ProjectId = sdk.String(base.PickResourceID(*req.ProjectId))
			for _, id := range idNames {
				req.SubnetId = sdk.String(base.PickResourceID(id))
				_, err := base.BizClient.DeleteSubnet(req)
				if err != nil {
					base.HandleError(err)
					continue
				}
				fmt.Fprintf(out, "subnet[%s] deleted\n", id)
			}
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&idNames, "subnet-id", nil, "Required. Resource ID of subent")
	bindRegion(req, flags)
	bindProjectID(req, flags)
	cmd.MarkFlagRequired("subnet-id")
	flags.SetFlagValuesFunc("subnet-id", func() []string {
		return getAllSubnetIDNames("", *req.ProjectId, *req.Region)
	})

	return cmd
}

//SubnetResourceRow 表格行
type SubnetResourceRow struct {
	ResourceName string
	ResourceID   string
	ResourceType string
	PrivateIP    string
}

//NewCmdSubnetListResource ucloud subnet list-resource
func NewCmdSubnetListResource(out io.Writer) *cobra.Command {
	req := base.BizClient.NewDescribeSubnetResourceRequest()
	cmd := &cobra.Command{
		Use:   "list-resource",
		Short: "List resources belong to subnet",
		Long:  "List resources belong to subnet",
		Run: func(c *cobra.Command, args []string) {
			req.SubnetId = sdk.String(base.PickResourceID(*req.SubnetId))
			resp, err := base.BizClient.DescribeSubnetResource(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			list := []SubnetResourceRow{}
			for _, r := range resp.DataSet {
				row := SubnetResourceRow{
					ResourceName: r.Name,
					ResourceID:   r.ResourceId,
					ResourceType: r.ResourceType,
					PrivateIP:    r.IP,
				}
				list = append(list, row)
			}
			base.PrintList(list, global.json)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.SubnetId = flags.String("subnet-id", "", "Required. Resource ID of subnet which resources to list belong to")
	req.ResourceType = flags.String("resource-type", "", "Optional. Resource type of resources to list. Accept values:'uhost','phost','ulb','uhadoophost','ufortresshost','unatgw','ukafka','umem','docker','udb','udw' and 'vip'")
	bindRegion(req, flags)
	bindProjectID(req, flags)
	bindLimit(req, flags)
	bindOffset(req, flags)
	cmd.MarkFlagRequired("subnet-id")
	flags.SetFlagValuesFunc("subnet-id", func() []string {
		return getAllSubnetIDNames("", *req.ProjectId, *req.Region)
	})
	flags.SetFlagValues("resource-type", "uhost", "phost", "ulb", "uhadoophost", "ufortresshost", "unatgw", "ukafka", "umem", "docker", "udb", "udw", "vip")

	return cmd
}

func getAllSubnets(vpcID, project, region string) ([]vpc.VPCSubnetInfoSet, error) {
	req := base.BizClient.NewDescribeSubnetRequest()
	req.ProjectId = sdk.String(base.PickResourceID(project))
	req.Region = sdk.String(region)
	if vpcID != "" {
		req.VPCId = sdk.String(base.PickResourceID(vpcID))
	}
	subnets := []vpc.VPCSubnetInfoSet{}
	for limit, offset := 50, 0; ; offset += limit {
		req.Limit = sdk.Int(limit)
		req.Offset = sdk.Int(offset)
		resp, err := base.BizClient.DescribeSubnet(req)
		if err != nil {
			base.HandleError(err)
			return nil, err
		}
		subnets = append(subnets, resp.DataSet...)
		if limit+offset >= resp.TotalCount {
			break
		}
	}
	return subnets, nil
}

func getAllSubnetIDNames(vpcID, project, region string) []string {
	subnets, err := getAllSubnets(vpcID, project, region)
	if err != nil {
		return nil
	}
	list := []string{}
	for _, s := range subnets {
		list = append(list, fmt.Sprintf("%s/%s", s.SubnetId, s.SubnetName))
	}
	return list
}
