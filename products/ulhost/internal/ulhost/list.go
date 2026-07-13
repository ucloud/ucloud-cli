package ulhost

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	ucompsharesdk "github.com/ucloud/ucloud-sdk-go/services/ucompshare"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/internal/common"
	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newList ucloud ulhost list
func newList(ctx *cli.Context) *cobra.Command {
	var allRegion, pageOff, idOnly bool
	var ulhostIds []string
	client := cli.NewServiceClient(ctx, ucompsharesdk.NewClient)
	req := client.NewDescribeULHostInstanceRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all ULHost Instances",
		Long:  `List all ULHost Instances`,
		Run: func(cmd *cobra.Command, args []string) {
			for _, ulhost := range ulhostIds {
				req.ULHostIds = append(req.ULHostIds, ctx.PickResourceID(ulhost))
			}

			ulhosts, err := getAllULHosts(ctx, client, req, pageOff, allRegion)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if idOnly {
				listULHostID(ctx, ulhosts)
			} else {
				listULHost(ctx, ulhosts, allRegion)
			}
		},
	}
	cmd.Flags().SortFlags = false
	req.ProjectId = cmd.Flags().String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = cmd.Flags().String("region", ctx.DefaultRegion(), "Optional. Assign region.")
	req.Offset = cmd.Flags().Int("offset", 0, "Optional. Offset default 0")
	req.Limit = cmd.Flags().Int("limit", 50, "Optional. Limit default 50, max value 100")
	cmd.Flags().StringSliceVar(&ulhostIds, "ulhost-id", make([]string, 0), "Optional. Resource ID of ulhost instances, multiple values separated by comma(without space)")
	cmd.Flags().BoolVar(&allRegion, "all-region", false, "Optional. Accept values: true or false. List ulhost instances of all regions when assigned true")
	cmd.Flags().BoolVar(&pageOff, "page-off", false, "Optional. Paging or not. If all-region is specified this flag will be true. Accept values: true or false. If assigned, the limit flag will be disabled and list all ulhost instances")
	cmd.Flags().BoolVar(&idOnly, "ulhost-id-only", false, "Optional. Just display resource id of ulhost")
	// NOTE: Unlike uhost, the ucompshare DescribeULHostInstanceRequest does not have
	// a Tag field, so ctx.BindGroup is not applicable here. Group filtering is not
	// supported by the ulhost describe API.

	command.SetFlagValues(cmd, "page-off", "true", "false")
	command.SetFlagValues(cmd, "ulhost-id-only", "true", "false")
	command.SetCompletion(cmd, "project-id", ctx.ProjectList)
	command.SetCompletion(cmd, "region", ctx.RegionList)
	command.SetCompletion(cmd, "ulhost-id", func() []string {
		return getULHostList(ctx, nil, *req.ProjectId, *req.Region)
	})

	return cmd
}

// listULHost renders the ulhost slice via ctx.PrintList, selecting columns per
// output mode using the per-mode row structs (rows.go).
func listULHost(ctx *cli.Context, ulhosts []ucompsharesdk.ULHostInstanceSet, listAllRegion bool) {
	list := make([]ulhostRow, 0)
	for _, host := range ulhosts {
		row := ulhostRow{}
		row.Name = host.Name
		row.Remark = host.Remark
		row.ResourceID = host.ULHostId
		row.Group = host.Tag
		for _, ip := range host.IPSet {
			if ip.Type == "Private" {
				row.PrivateIP = ip.IP
			} else {
				if row.PublicIP != "" {
					row.PublicIP += " | "
				}
				row.PublicIP += ip.IP
			}
		}
		memorySize := host.Memory / 1024
		var disks []string
		for _, disk := range host.DiskSet {
			disks = append(disks, fmt.Sprintf("%s:%s:%dG", disk.Type, disk.DiskType, disk.Size))
		}
		row.Zone = host.Zone
		row.DiskSet = strings.Join(disks, "|")
		row.Config = fmt.Sprintf("cpu:%d memory:%dG", host.CPU, memorySize)
		row.Image = fmt.Sprintf("%s|%s", host.ImageId, host.ImageName)
		row.CreationTime = common.FormatDate(host.CreateTime)
		row.State = host.State
		row.ChargeType = host.ChargeType
		row.AutoRenew = host.AutoRenew
		row.ExpireTime = common.FormatDate(host.ExpireTime)
		list = append(list, row)
	}

	// JSON/YAML mode: print the full row set.
	if ctx.Format() != cli.OutputTable {
		ctx.PrintList(list)
		return
	}

	if listAllRegion {
		rows := make([]ulhostRowAllRegion, 0, len(list))
		for _, r := range list {
			rows = append(rows, ulhostRowAllRegion{
				Name: r.Name, ResourceID: r.ResourceID, Group: r.Group,
				PublicIP: r.PublicIP, Config: r.Config,
				Image: r.Image, State: r.State, ChargeType: r.ChargeType,
				CreationTime: r.CreationTime, Zone: r.Zone,
			})
		}
		ctx.PrintList(rows)
		return
	}
	rows := make([]ulhostRowDefault, 0, len(list))
	for _, r := range list {
		rows = append(rows, ulhostRowDefault{
			Name: r.Name, ResourceID: r.ResourceID, Group: r.Group,
			PublicIP: r.PublicIP, Config: r.Config,
			Image: r.Image, State: r.State, ChargeType: r.ChargeType,
			CreationTime: r.CreationTime,
		})
	}
	ctx.PrintList(rows)
}

func listULHostID(ctx *cli.Context, ulhosts []ucompsharesdk.ULHostInstanceSet) {
	ids := make([]string, 0)
	for _, h := range ulhosts {
		ids = append(ids, h.ULHostId)
	}
	fmt.Fprintln(ctx.Out(), strings.Join(ids, ","))
}

func fetchULHosts(client *ucompsharesdk.UCompShareClient, req *ucompsharesdk.DescribeULHostInstanceRequest) ([]ucompsharesdk.ULHostInstanceSet, error) {
	resp, err := client.DescribeULHostInstance(req)
	if err != nil {
		return nil, err
	}
	return resp.ULHostInstanceSets, nil
}

func fetchULHostsPageOff(client *ucompsharesdk.UCompShareClient, req *ucompsharesdk.DescribeULHostInstanceRequest) ([]ucompsharesdk.ULHostInstanceSet, error) {
	_req := *req
	result := make([]ucompsharesdk.ULHostInstanceSet, 0)
	for limit, offset := 50, 0; ; offset += limit {
		_req.Offset = sdk.Int(offset)
		_req.Limit = sdk.Int(limit)
		ulhosts, err := fetchULHosts(client, &_req)
		if err != nil {
			return nil, err
		}
		result = append(result, ulhosts...)
		// The ucompshare SDK does not return TotalCount, so we stop when
		// fewer results than the limit are returned.
		if len(ulhosts) < limit {
			break
		}
	}
	return result, nil
}

func getAllULHosts(ctx *cli.Context, client *ucompsharesdk.UCompShareClient, req *ucompsharesdk.DescribeULHostInstanceRequest, pageOff bool, allRegion bool) ([]ucompsharesdk.ULHostInstanceSet, error) {
	if allRegion {
		result := make([]ucompsharesdk.ULHostInstanceSet, 0)
		regions, err := ctx.AllRegions()
		if err != nil {
			return nil, err
		}
		for _, region := range regions {
			_req := *req
			_req.Region = sdk.String(region)
			ulhosts, err := fetchULHostsPageOff(client, &_req)
			if err != nil {
				continue
			}
			result = append(result, ulhosts...)
		}
		return result, nil
	}

	if pageOff {
		_req := *req
		ulhosts, err := fetchULHostsPageOff(client, &_req)
		if err != nil {
			return nil, err
		}
		return ulhosts, nil
	}

	ulhosts, err := fetchULHosts(client, req)
	if err != nil {
		return nil, err
	}
	return ulhosts, nil
}
