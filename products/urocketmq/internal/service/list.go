package service

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/urocketmq"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"
	sdkerror "github.com/ucloud/ucloud-sdk-go/ucloud/error"

	"github.com/ucloud/ucloud-cli/internal/common"
	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newList ucloud urocketmq service list
func newList(ctx *cli.Context) *cobra.Command {
	var allRegion, idOnly bool
	client := cli.NewServiceClient(ctx, urocketmq.NewClient)
	req := client.NewListURocketMQServiceRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all URocketMQ instances",
		Long:  "List all URocketMQ instances",
		Run: func(cmd *cobra.Command, args []string) {
			services, err := getAllServices(ctx, client, req, allRegion)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if idOnly {
				listServiceID(ctx, services)
			} else {
				listService(ctx, services, allRegion)
			}
		},
	}
	cmd.Flags().SortFlags = false

	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)
	req.Limit = cmd.Flags().Int("limit", 20, "Optional. Limit default 20, max value 1000")
	req.Offset = cmd.Flags().Int("offset", 0, "Optional. Offset default 0")
	cmd.Flags().BoolVar(&allRegion, "all-region", false, "Optional. Accept values: true or false. List URocketMQ instances of all regions when assigned true")
	cmd.Flags().BoolVar(&idOnly, "id-only", false, "Optional. Just display resource id of URocketMQ service")

	command.SetFlagValues(cmd, "all-region", "true", "false")
	command.SetFlagValues(cmd, "id-only", "true", "false")

	return cmd
}

// getAllServices handles --all-region cross-region aggregation; single region fetches one page by user limit/offset.
func getAllServices(ctx *cli.Context, client *urocketmq.URocketMQClient, req *urocketmq.ListURocketMQServiceRequest, allRegion bool) ([]urocketmq.ServiceBaseInfo, error) {
	if allRegion {
		result := make([]urocketmq.ServiceBaseInfo, 0)
		regions, err := ctx.AllRegions()
		if err != nil {
			return nil, err
		}
		for _, region := range regions {
			_req := *req
			_req.Region = sdk.String(region)
			// --all-region does not paginate, fetches all per region
			services, err := fetchServicesPageOff(client, &_req)
			// Some accounts lack URocketMQ permissions in the current region; skip per platform convention RetCode 230
			if e, ok := err.(sdkerror.Error); ok && e.Code() == _RetCodeRegionNoPermission {
				continue
			}
			if err != nil {
				return nil, err
			}
			result = append(result, services...)
		}
		return result, nil
	}

	resp, err := client.ListURocketMQService(req)
	if err != nil {
		return nil, err
	}
	return resp.ServiceList, nil
}

// fetchServicesPageOff paginates all URocketMQ instances in the specified region. SDK response has no
// TotalCount, so uses last page item count < pageSize as termination condition.
func fetchServicesPageOff(client *urocketmq.URocketMQClient, req *urocketmq.ListURocketMQServiceRequest) ([]urocketmq.ServiceBaseInfo, error) {
	_req := *req
	result := make([]urocketmq.ServiceBaseInfo, 0)
	for limit, offset := 100, 0; ; offset += limit {
		_req.Offset = sdk.Int(offset)
		_req.Limit = sdk.Int(limit)
		resp, err := client.ListURocketMQService(&_req)
		if err != nil {
			return nil, err
		}
		result = append(result, resp.ServiceList...)
		if len(resp.ServiceList) < limit {
			break
		}
	}
	return result, nil
}

// listService renders the service list. json/yaml emits full-field serviceRow; table mode uses curated
// columns (serviceRowDefault, serviceRowAllRegion for --all-region with Region appended).
func listService(ctx *cli.Context, services []urocketmq.ServiceBaseInfo, listAllRegion bool) {
	list := make([]serviceRow, 0, len(services))
	for _, s := range services {
		list = append(list, toServiceRow(s))
	}

	// JSON/YAML mode: emits full-field rows (like uhost, --json always marshals full fields).
	if ctx.Format() != cli.OutputTable {
		ctx.PrintList(list)
		return
	}

	if listAllRegion {
		rows := make([]serviceRowAllRegion, 0, len(list))
		for _, r := range list {
			rows = append(rows, serviceRowAllRegion{
				Name: r.Name, ServiceId: r.ServiceId, State: r.State,
				Config: formatServiceConfig(r.Tps, r.Storage), Address: r.Address,
				CreateTime: common.FormatDate(r.CreateTime), ExpireTime: common.FormatDate(r.ExpireTime),
				Region: r.Region,
			})
		}
		ctx.PrintList(rows)
		return
	}

	rows := make([]serviceRowDefault, 0, len(list))
	for _, r := range list {
		rows = append(rows, serviceRowDefault{
			Name: r.Name, ServiceId: r.ServiceId, State: r.State,
			Config: formatServiceConfig(r.Tps, r.Storage), Address: r.Address,
			CreateTime: common.FormatDate(r.CreateTime), ExpireTime: common.FormatDate(r.ExpireTime),
		})
	}
	ctx.PrintList(rows)
}

// toServiceRow maps SDK ServiceBaseInfo to a full-field row.
func toServiceRow(s urocketmq.ServiceBaseInfo) serviceRow {
	return serviceRow{
		ServiceId:       s.ServiceId,
		Name:            s.Name,
		State:           s.State,
		Tps:             s.Tps,
		Storage:         s.Storage,
		TopicLimit:      s.TopicLimit,
		Address:         s.Address,
		AddressExtranet: s.AddressExtranet,
		VpcId:           s.VpcId,
		SubnetId:        s.SubnetId,
		ChargeType:      s.ChargeType,
		CreateTime:      s.CreateTime,
		ExpireTime:      s.ExpireTime,
		Remark:          s.Remark,
		Tag:             s.Tag,
		Edition:         s.Edition,
		Mode:            s.Mode,
		AutoRenew:       s.AutoRenew,
		IsExpire:        s.IsExpire,
		Quantity:        s.Quantity,
		Region:          s.Region,
	}
}

// formatServiceConfig concatenates the Config column in table mode: Tps + Storage.
func formatServiceConfig(tps, storage int) string {
	return fmt.Sprintf("tps:%d storage:%dG", tps, storage)
}

// listServiceID outputs only the ServiceId list to ctx.Out() (not ProgressWriter) for script capture.
// Corresponds to uhost listUhostID.
func listServiceID(ctx *cli.Context, services []urocketmq.ServiceBaseInfo) {
	ids := make([]string, 0, len(services))
	for _, s := range services {
		ids = append(ids, s.ServiceId)
	}
	fmt.Fprintln(ctx.Out(), strings.Join(ids, ","))
}

// _RetCodeRegionNoPermission is the SDK RetCode when account lacks permission in the current region;
// --all-region path skips that region. Follows uhost (cmd/uhost.go) platform convention.
const _RetCodeRegionNoPermission = 230
