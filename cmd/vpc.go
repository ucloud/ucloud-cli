package cmd

import (
	"github.com/spf13/cobra"
	. "github.com/ucloud/ucloud-cli/util"
)

//NewCmdVpc ucloud vpc
func NewCmdVpc() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vpc",
		Short: "List vpc",
		Long:  "List vpc",
		Args:  cobra.NoArgs,
	}

	cmd.AddCommand(NewCmdVpcCreate())
	cmd.AddCommand(NewCmdVPCList())
	cmd.AddCommand(NewCmdVpcDelete())
	cmd.AddCommand(NewCmdVpcCreatePeer())
	cmd.AddCommand(NewCmdVpcListPeer())
	cmd.AddCommand(NewCmdVpcDeletePeer())
	cmd.AddCommand(NewCmdSubnetCreate())
	return cmd
}

/* type VPCRow struct {
	VPCName        string
	ResourceID     string
	Group          string
	NetworkSegment string
	SubnetCount    int
	CreationTime   string
}

type SubnetRow struct {
	SubnetName     string
	ResourceID     string
	Group          string
	AffiliatedVPC  string
	NetworkSegment string
	CreationTime   string
}
*/

type VPCIntercomRow struct {
	ProjectId string
	Network   []string
	DstRegion string
	Name      string
	VPCId     string
	Tag       string
}

//NewCreateVPCRequest ucloud vpc create
func NewCmdVpcCreate() *cobra.Command {
	var segments *[]string
	req := BizClient.NewCreateVPCRequest()
	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Create vpc network",
		Long:    "Create vpc network",
		Example: "ucloud vpc create --name xxx --segment xxx",
		Args:    cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			req.Network = *segments
			resp, err := BizClient.CreateVPC(req)
			if err != nil {
				HandleError(err)
				return
			}
			Cxt.Printf("VPC: %v created successfully!\n", resp.VPCId)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	req.Name = cmd.Flags().String("name", "", "Required. Name of the vpc network.")
	segments = cmd.Flags().StringSlice("segment", nil, "Required. The segment for private network.")
	req.Tag = cmd.Flags().String("Group", "Default", "Optional. Business group.")
	req.Remark = cmd.Flags().String("Remark", "Default", "Optional. The description of the vpc.")
	req.Region = cmd.Flags().String("region", ConfigInstance.Region, "Optional. Assign the region of the VPC")
	req.ProjectId = cmd.Flags().String("project-id", ConfigInstance.ProjectID, "Optional. Assign the project-id")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("segment")

	return cmd
}

//NewDeleteVPCRequest ucloud vpc delete
func NewCmdVpcDelete() *cobra.Command {
	req := BizClient.NewDeleteVPCRequest()
	cmd := &cobra.Command{
		Use:     "delete",
		Short:   "Delete vpc network",
		Long:    "Delete vpc network",
		Example: "ucloud vpc delete --vpc-id",
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := BizClient.DeleteVPC(req)
			if err != nil {
				HandleError(err)
			} else {
				Cxt.Printf("VPC [%s] was successfully deleted\n ", resp)
			}
		},
	}

	cmd.Flags().SortFlags = false

	req.VPCId = cmd.Flags().String("vpc-id", "", "Required. The vpc network you want to delete")
	req.Region = cmd.Flags().String("region", ConfigInstance.Region, "Optional. Clarify the region of the vpc")
	req.ProjectId = cmd.Flags().String("project-id", ConfigInstance.ProjectID, "Optional. The project id of the vpc")
	cmd.MarkFlagRequired("vpc-id")

	return cmd
}

//NewCreateVPCIntercomRequest ucloud vpc peer
func NewCmdVpcCreatePeer() *cobra.Command {
	req := BizClient.NewCreateVPCIntercomRequest()
	cmd := &cobra.Command{
		Use:     "create-intercome",
		Short:   "create intercome with other vpc",
		Long:    "create intercome with other vpc",
		Example: "ucloud vpc create-intercome --vpc-id --dstvpc-id --destregion",
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := BizClient.CreateVPCIntercom(req)
			if err != nil {
				HandleError(err)
				return
			}
			Cxt.Printf("The intercome [%s] has been establish", resp)
		},
	}

	cmd.Flags().SortFlags = false

	req.VPCId = cmd.Flags().String("vpc-id", "", "Required. The source vpc you want to establish the intercome")
	req.DstVPCId = cmd.Flags().String("dstvpc-id", "", "Required. The target vpc you want to establish the intercome")
	req.DstRegion = cmd.Flags().String("dstregion", "", "Required. If the intercome established across different regions")
	req.Region = cmd.Flags().String("region", ConfigInstance.Region, "Optioanl. The region of source vpc which will establish the intercome")
	req.ProjectId = cmd.Flags().String("project-id", ConfigInstance.ProjectID, "Optional. The project id of the source vpc")
	cmd.MarkFlagRequired("vpc-id")
	cmd.MarkFlagRequired("dstvpc-id")
	cmd.MarkFlagRequired("dstregion")

	return cmd
}

//NewDescribeVPCIntercomRequest
func NewCmdVpcListPeer() *cobra.Command {
	req := BizClient.NewDescribeVPCIntercomRequest()
	cmd := &cobra.Command{
		Use:     "list-intercome",
		Short:   "list intercome ",
		Long:    "list intercome",
		Example: "ucloud vpc list-intercome --vpc-id xx",
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := BizClient.DescribeVPCIntercom(req)
			if err != nil {
				HandleError(err)
				return
			}
			if global.json {
				PrintJSON(resp.DataSet)
			} else {
				list := make([]VPCIntercomRow, 0)
				for _, VPCIntercom := range resp.DataSet {
					row := VPCIntercomRow{}
					row.ProjectId = VPCIntercom.ProjectId
					row.Network = VPCIntercom.Network
					row.DstRegion = VPCIntercom.DstRegion
					row.Name = VPCIntercom.Name
					row.VPCId = VPCIntercom.VPCId
					row.Tag = VPCIntercom.Tag
				}
				PrintTable(list, []string{"ProjectId", "Network", "DstRegion", "Name", "VPCId", "Tag"})
			}

		},
	}
	req.VPCId = cmd.Flags().String("vpc-id", "", "Required. The vpc id which you wnat to describe the information")
	req.ProjectId = cmd.Flags().String("project-id", ConfigInstance.ProjectID, "Optional. The project id of source vpc")
	req.Region = cmd.Flags().String("region", ConfigInstance.Region, "Optional, The region of source vpc")
	cmd.MarkFlagRequired("vpc-id")

	return cmd
}

//NewDeleteVPCIntercomRequest
func NewCmdVpcDeletePeer() *cobra.Command {
	req := BizClient.NewDeleteVPCIntercomRequest()
	cmd := &cobra.Command{
		Use:     "delete-intercome",
		Short:   "delete the vpc intercome",
		Long:    "delete the vpc intercome",
		Example: "ucloud vpc delete-intercome --vpc-id xxx --dstvpc-id xxx",
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := BizClient.DeleteVPCIntercom(req)
			if err != nil {
				HandleError(err)
				return
			}
			Cxt.Printf("The intercome [%s] was deleted successfully", resp)
		},
	}

	cmd.Flags().SortFlags = false

	req.VPCId = cmd.Flags().String("vpc-id", "", "Required. The source vpc id from which you want to disconnect")
	req.DstVPCId = cmd.Flags().String("dstvpc-id", "", "Required. The target vpc which you want to disconnect with source vpc")
	req.ProjectId = cmd.Flags().String("project-id", ConfigInstance.ProjectID, "Optional. The project id of source vpc")
	req.Region = cmd.Flags().String("region", ConfigInstance.Region, "Optional. The region of source vpc from which you want to disconnect")
	cmd.MarkFlagRequired("vpc-id")
	cmd.MarkFlagRequired("dstvpc-id")
	return cmd
}

//NewCreateSubnetRequest  ucloud subnet create
func NewCmdSubnetCreate() *cobra.Command {
	req := BizClient.NewCreateSubnetRequest()
	cmd := &cobra.Command{
		Use:     "create-subnet",
		Short:   "Create subnet of vpc network",
		Long:    "Create subnet of vpc network",
		Example: "ucloud subnet create --vpc-id --segment",
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := BizClient.CreateSubnet(req)
			if err != nil {
				HandleError(err)
				return
			}
			Cxt.Printf("Subnet : %v created successfully!\n", resp.SubnetId)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	req.VPCId = cmd.Flags().String("vpc-id", "", "Required. Assign the VPC network of the subnet")
	req.Subnet = cmd.Flags().String("segment", "", "Required. Same as the vpc network")
	req.Region = cmd.Flags().String("region", ConfigInstance.Region, "Optional. The region of the subnet")
	req.ProjectId = cmd.Flags().String("project-id", ConfigInstance.ProjectID, "Optional. The project id of the subnet")
	req.Netmask = cmd.Flags().Int("netmask", 24, "Optional. The number of the IPs, default is 24")
	req.SubnetName = cmd.Flags().String("name", "Subnet", "Optional. The default is Subnet")
	req.Tag = cmd.Flags().String("Group", "Default", "Optional. Business group")
	cmd.MarkFlagRequired("vpc-id")
	cmd.MarkFlagRequired("segment")

	return cmd
}
