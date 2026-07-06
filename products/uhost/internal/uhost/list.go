package uhost

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	uhostsdk "github.com/ucloud/ucloud-sdk-go/services/uhost"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"
	sdkerror "github.com/ucloud/ucloud-sdk-go/ucloud/error"

	"github.com/ucloud/ucloud-cli/internal/common"
	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newList ucloud uhost list
func newList(ctx *cli.Context) *cobra.Command {
	var allRegion, pageOff, idOnly bool
	var uhostIds []string
	client := cli.NewServiceClient(ctx, uhostsdk.NewClient)
	req := client.NewDescribeUHostInstanceRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all UHost Instances",
		Long:  `List all UHost Instances`,
		Run: func(cmd *cobra.Command, args []string) {
			*req.VPCId = ctx.PickResourceID(*req.VPCId)
			*req.SubnetId = ctx.PickResourceID(*req.SubnetId)
			*req.IsolationGroup = ctx.PickResourceID(*req.IsolationGroup)
			for _, uhost := range uhostIds {
				req.UHostIds = append(req.UHostIds, ctx.PickResourceID(uhost))
			}

			uhosts, err := getAllUHosts(ctx, client, req, pageOff, allRegion)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if idOnly {
				listUhostID(ctx, uhosts)
			} else {
				listUhost(ctx, uhosts, allRegion)
			}
		},
	}
	cmd.Flags().SortFlags = false
	req.ProjectId = cmd.Flags().String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = cmd.Flags().String("region", ctx.DefaultRegion(), "Optional. Assign region.")
	req.Zone = cmd.Flags().String("zone", "", "Optional. Assign availability zone")
	req.Offset = cmd.Flags().Int("offset", 0, "Optional. Offset default 0")
	req.Limit = cmd.Flags().Int("limit", 50, "Optional. Limit default 50, max value 100")
	req.VPCId = cmd.Flags().String("vpc-id", "", "Optional. Resource ID of VPC. List uhost instances of the specified VPC")
	req.SubnetId = cmd.Flags().String("subnet-id", "", "Optional. Resource ID of Subnet. List uhost instances of the specified Subnet")
	req.IsolationGroup = cmd.Flags().String("isolation-group", "", "Optional. Resource ID of isolation group. List uhost instances of the specified isolation group")
	cmd.Flags().StringSliceVar(&uhostIds, "uhost-id", make([]string, 0), "Optional. Resource ID of uhost instances, multiple values separated by comma(without space)")
	cmd.Flags().BoolVar(&allRegion, "all-region", false, "Optional. Accpet values: true or false. List uhost instances of all regions when assigned true")
	cmd.Flags().BoolVar(&pageOff, "page-off", false, "Optional. Paging or not. If all-region is specified this flag will be true. Accept values: true or false. If assigned, the limit flag will be disabled and list all uhost instances")
	cmd.Flags().BoolVar(&idOnly, "uhost-id-only", false, "Optional. Just display resource id of uhost")
	ctx.BindGroup(cmd, req)

	command.SetFlagValues(cmd, "page-off", "true", "false")
	command.SetFlagValues(cmd, "uhost-id-only", "true", "false")
	command.SetCompletion(cmd, "project-id", ctx.ProjectList)
	command.SetCompletion(cmd, "region", ctx.RegionList)
	command.SetCompletion(cmd, "zone", func() []string {
		return ctx.ZoneList(req.GetRegion())
	})

	command.SetCompletion(cmd, "vpc-id", func() []string {
		return getAllVPCIdNames(ctx, *req.ProjectId, *req.Region)
	})
	command.SetCompletion(cmd, "subnet-id", func() []string {
		return getAllSubnetIDNames(ctx, *req.VPCId, *req.ProjectId, *req.Region)
	})
	command.SetCompletion(cmd, "isolation-group", func() []string {
		return getIsolationGroupList(ctx, *req.ProjectId, *req.Region)
	})
	command.SetCompletion(cmd, "uhost-id", func() []string {
		return getUhostList(ctx, nil, *req.ProjectId, *req.Region, *req.Zone)
	})

	return cmd
}

// listUhost renders the uhost slice via ctx.PrintList, selecting columns per
// output mode using the per-mode row structs (rows.go). AWS-style: --output
// selects only the format — table shows curated columns (uhostRowDefault, or
// uhostRowAllRegion with a trailing Zone under --all-region); json/yaml always
// emit the full uhostRow so no field is lost (e.g. DiskSet/VPC/Subnet).
func listUhost(ctx *cli.Context, uhosts []uhostsdk.UHostInstanceSet, listAllRegion bool) {
	list := make([]uhostRow, 0)
	for _, host := range uhosts {
		row := uhostRow{}
		row.UHostName = host.Name
		row.Remark = host.Remark
		row.ResourceID = host.UHostId
		row.Group = host.Tag
		for _, ip := range host.IPSet {
			if row.PublicIP != "" {
				row.PublicIP += " | "
			}
			if ip.Type == "Private" {
				row.PrivateIP = ip.IP
				row.VPC = ip.VPCId
				row.Subnet = ip.SubnetId
			} else {
				row.PublicIP += fmt.Sprintf("%s", ip.IP)
			}
		}
		cupCore := host.CPU
		memorySize := host.Memory / 1024
		diskSize := 0
		var disks []string
		for _, disk := range host.DiskSet {
			if disk.Type == "Data" || disk.Type == "Udisk" {
				diskSize += disk.Size
			}
			disks = append(disks, fmt.Sprintf("%s:%s:%dG", disk.Type, disk.DiskType, disk.Size))
		}
		row.Zone = host.Zone
		row.DiskSet = strings.Join(disks, "|")
		row.Config = fmt.Sprintf("cpu:%d memory:%dG disk:%dG", cupCore, memorySize, diskSize)
		row.Image = fmt.Sprintf("%s|%s", host.BasicImageId, host.BasicImageName)
		row.CreationTime = common.FormatDate(host.CreateTime)
		row.State = host.State
		row.Type = host.MachineType + "/" + host.HostType
		if host.HotplugFeature {
			row.Type += "/HotPlug"
		}
		list = append(list, row)
	}

	// JSON/YAML mode: print the full row set (matches cmd/uhost.go, which
	// marshalled the full UHostRow slice in --json mode). ctx.PrintList routes
	// json/yaml by format; for table mode we narrow to the per-mode struct.
	if ctx.Format() != cli.OutputTable {
		ctx.PrintList(list)
		return
	}

	if listAllRegion {
		rows := make([]uhostRowAllRegion, 0, len(list))
		for _, r := range list {
			rows = append(rows, uhostRowAllRegion{
				UHostName: r.UHostName, ResourceID: r.ResourceID, Group: r.Group,
				PrivateIP: r.PrivateIP, PublicIP: r.PublicIP, Config: r.Config,
				Image: r.Image, Type: r.Type, State: r.State,
				CreationTime: r.CreationTime, Zone: r.Zone,
			})
		}
		ctx.PrintList(rows)
		return
	}
	rows := make([]uhostRowDefault, 0, len(list))
	for _, r := range list {
		rows = append(rows, uhostRowDefault{
			UHostName: r.UHostName, ResourceID: r.ResourceID, Group: r.Group,
			PrivateIP: r.PrivateIP, PublicIP: r.PublicIP, Config: r.Config,
			Image: r.Image, Type: r.Type, State: r.State, CreationTime: r.CreationTime,
		})
	}
	ctx.PrintList(rows)
}

func listUhostID(ctx *cli.Context, uhosts []uhostsdk.UHostInstanceSet) {
	ids := make([]string, 0)
	for _, u := range uhosts {
		ids = append(ids, u.UHostId)
	}
	// The id list IS the result of --uhost-id-only, not narration: write it to
	// stdout (ctx.Out), never ProgressWriter — otherwise in non-TTY/json mode the
	// ids go to stderr and `ids=$(ucloud uhost list --uhost-id-only)` captures
	// nothing.
	fmt.Fprintln(ctx.Out(), strings.Join(ids, ","))
}

func fetchUHosts(client *uhostsdk.UHostClient, req *uhostsdk.DescribeUHostInstanceRequest) ([]uhostsdk.UHostInstanceSet, int, error) {
	resp, err := client.DescribeUHostInstance(req)
	if err != nil {
		return nil, 0, err
	}
	return resp.UHostSet, resp.TotalCount, nil
}

func fetchUHostsPageOff(client *uhostsdk.UHostClient, req *uhostsdk.DescribeUHostInstanceRequest) ([]uhostsdk.UHostInstanceSet, error) {
	_req := *req
	result := make([]uhostsdk.UHostInstanceSet, 0)
	for limit, offset := 50, 0; ; offset += limit {
		_req.Offset = sdk.Int(offset)
		_req.Limit = sdk.Int(limit)
		uhosts, total, err := fetchUHosts(client, &_req)
		if err != nil {
			return nil, err
		}
		result = append(result, uhosts...)
		if offset+limit >= total {
			break
		}
	}
	return result, nil
}

func getAllUHosts(ctx *cli.Context, client *uhostsdk.UHostClient, req *uhostsdk.DescribeUHostInstanceRequest, pageOff bool, allRegion bool) ([]uhostsdk.UHostInstanceSet, error) {
	if allRegion {
		result := make([]uhostsdk.UHostInstanceSet, 0)
		regions, err := ctx.AllRegions()
		if err != nil {
			return nil, err
		}
		for _, region := range regions {
			_req := *req
			_req.Region = sdk.String(region)
			//如果要获取所有region的主机，则不分页
			uhosts, err := fetchUHostsPageOff(client, &_req)
			// Has no permission in current region for UHost
			if e, ok := err.(sdkerror.Error); ok && e.Code() == _RetCodeRegionNoPermission {
				continue
			}
			if err != nil {
				return nil, err
			}
			result = append(result, uhosts...)
		}
		return result, nil
	}

	if pageOff {
		_req := *req
		uhosts, err := fetchUHostsPageOff(client, &_req)
		if err != nil {
			return nil, err
		}
		return uhosts, nil
	}

	uhosts, _, err := fetchUHosts(client, req)
	if err != nil {
		return nil, err
	}
	return uhosts, nil
}

// _RetCodeRegionNoPermission is the SDK RetCode returned when the account has no
// permission for UHost in a region; the --all-region path skips such regions.
// Verbatim from cmd/uhost.go.
const _RetCodeRegionNoPermission = 230
