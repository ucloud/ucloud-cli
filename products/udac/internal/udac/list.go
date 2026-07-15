package udac

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/ucloud/ucloud-sdk-go/services/uaccount"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// actionForType 根据实例类型返回对应的 UDAC Action 名。
func actionForType(instanceType string) (string, error) {
	switch instanceType {
	case "mysql":
		return "ListUDACMySQLInstance", nil
	case "mongodb":
		return "ListUDACUMongoDBClusters", nil
	default:
		return "", fmt.Errorf("unsupported instance type: %s, supported: %v", instanceType, SupportedTypes)
	}
}

// regionFromZone 从 zone 推导 region。
// UCloud zone 命名规则：{region}-{suffix}，例如 cn-bj2-02 → cn-bj2、hk-02 → hk。
func regionFromZone(zone string) string {
	if i := strings.LastIndex(zone, "-"); i > 0 {
		return zone[:i]
	}
	return zone
}

// newInstanceRow 从 API 返回的 map 构造一行展示数据。
// overrideRegion 非空时用用户传入值，否则优先读 API 返回的 Region，再兜底从 Zone 推导。
// MongoDB 响应字段：ClusterId/Name/Region/State/JoinTime（无 Zone）。
// MySQL 响应字段：ID 或 InstanceId/Name/Zone/State/Status/JoinTime（无 Region）。
func newInstanceRow(m map[string]interface{}, instanceType, overrideRegion string) (importedInstanceRow, bool) {
	id := firstString(m, "ClusterId", "InstanceId", "ID")
	if id == "" {
		return importedInstanceRow{}, false
	}
	zoneVal := getString(m, "Zone")
	regionVal := getString(m, "Region")
	if regionVal == "" {
		regionVal = overrideRegion
	}
	if regionVal == "" {
		regionVal = regionFromZone(zoneVal)
	}
	status := firstString(m, "State", "Status")
	joinTime := getInt64(m, "JoinTime")
	importTime := ""
	if joinTime > 0 {
		importTime = time.Unix(joinTime, 0).Format(time.RFC3339)
	}
	return importedInstanceRow{
		ResourceID: id,
		InstanceID: id,
		Name:       getString(m, "Name"),
		Type:       instanceType,
		Status:     status,
		ImportTime: importTime,
		Region:     regionVal,
		Zone:       zoneVal,
	}, true
}

// firstString 按顺序尝试多个 key，返回第一个非空字符串。
func firstString(m map[string]interface{}, keys ...string) string {
	for _, k := range keys {
		if v := getString(m, k); v != "" {
			return v
		}
	}
	return ""
}

// fetchUDACInstances 调用 UDAC list API，返回账号下全部实例。
func fetchUDACInstances(ctx *cli.Context, action, projectID string) ([]map[string]interface{}, error) {
	params := map[string]interface{}{
		"Action":    action,
		"ProjectId": projectID,
	}

	client := cli.NewServiceClient(ctx, uaccount.NewClient)
	req := client.NewGenericRequest()
	if err := req.SetPayload(params); err != nil {
		return nil, fmt.Errorf("set payload: %w", err)
	}
	resp, err := client.GenericInvoke(req)
	if err != nil {
		return nil, err
	}
	payload := resp.GetPayload()

	var raw []interface{}
	for _, key := range []string{"InstanceInfoSet", "Instances", "DataSet"} {
		if val, ok := payload[key].([]interface{}); ok {
			raw = val
			break
		}
	}
	out := make([]map[string]interface{}, 0, len(raw))
	for _, item := range raw {
		if m, ok := item.(map[string]interface{}); ok {
			out = append(out, m)
		}
	}
	return out, nil
}

// newList implements `ucloud udac list`
// --project-id 必填（配置默认值兜底）；其他可选。
func newList(ctx *cli.Context) *cobra.Command {
	var instanceID, instanceType, statusFilter string
	var allRegions bool
	var common request.CommonBase

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List imported database instances in UDAC",
		Long: `List database instances that have been imported into the Database Autonomous Center (UDAC).

Required flag: --project-id (falls back to default-project-id from config if set).

Optional filters:
  --type         Instance type: mysql, mongodb
  --region       Filter by region. Defaults to config's default-region.
  --zone         Filter by zone. If omitted, list across all zones in the region.
  --udb-id   	 List only the specified instance.
  --status       Filter by status (e.g., Running, Failed).
  --all-regions  List instances across all regions (ignore --region and config default).

When both --region and --zone are specified, the zone must belong to the region.`,
		Run: func(c *cobra.Command, args []string) {
			// 1. 必填校验：project-id（配置默认值兜底）
			projectID := common.GetProjectId()
			if projectID == "" {
				ctx.HandleError(fmt.Errorf("required flag(s) not set: project-id"))
				return
			}

			// 2. 类型校验 + Action 选择
			action, err := actionForType(instanceType)
			if err != nil {
				ctx.HandleError(err)
				return
			}

			// 3. 确定 region 过滤值：--all-regions 优先级最高
			if allRegions && c.Flags().Changed("region") {
				ctx.HandleError(fmt.Errorf("--all-regions and --region are mutually exclusive"))
				return
			}
			region := common.GetRegion()
			zone := common.GetZone()
			if allRegions {
				region = ""
			}

			// 4. region/zone 一致性校验
			if region != "" && zone != "" && !strings.HasPrefix(zone, region+"-") {
				ctx.HandleError(fmt.Errorf("zone %s does not belong to region %s", zone, region))
				return
			}

			// mongodb 响应无 Zone 字段，--zone 过滤会静默清空
			if instanceType == "mongodb" && zone != "" {
				ctx.HandleError(fmt.Errorf("--zone is not supported for mongodb, use --region instead"))
				return
			}

			// 5. 拉取全部实例（API 只带 ProjectId，不传 Region/Zone）
			instances, err := fetchUDACInstances(ctx, action, projectID)
			if err != nil {
				ctx.HandleError(err)
				return
			}

			// 6. 客户端过滤
			instanceID = ctx.PickResourceID(instanceID)
			rows := make([]importedInstanceRow, 0, len(instances))
			for _, m := range instances {
				if instanceID != "" && firstString(m, "ClusterId", "InstanceId", "ID") != instanceID {
					continue
				}
				zoneVal := getString(m, "Zone")
				if zone != "" && zoneVal != zone {
					continue
				}
				if region != "" {
					// 优先用 API 返回的 Region，其次从 Zone 推导
					instanceRegion := getString(m, "Region")
					if instanceRegion == "" {
						instanceRegion = regionFromZone(zoneVal)
					}
					if instanceRegion != region {
						continue
					}
				}
				if statusFilter != "" && firstString(m, "State", "Status") != statusFilter {
					continue
				}
				if row, ok := newInstanceRow(m, instanceType, region); ok {
					rows = append(rows, row)
				}
			}

			// 7. 输出（空列表正常输出空表，不报错）
			ctx.PrintList(rows)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringVar(&instanceID, resourceIDFlag, "", "Optional. List only the specified instance.")
	flags.StringVar(&instanceType, typeFlag, "mysql", "Optional. Instance type: mysql, mongodb.")
	flags.StringVar(&statusFilter, "status", "", "Optional. Filter by status (e.g., Running, Failed).")
	flags.BoolVar(&allRegions, "all-regions", false, "Optional. List instances across all regions (ignore --region and config default).")

	// 公共参数绑定：region/zone/project-id 都用配置默认值兜底
	ctx.BindRegion(cmd, &common)
	ctx.BindZoneEmpty(cmd, &common)
	ctx.BindProjectID(cmd, &common)

	command.SetFlagValues(cmd, typeFlag, SupportedTypes...)
	command.SetCompletion(cmd, resourceIDFlag, func() []string {
		return listImportedInstanceIDs(ctx, instanceType, common.GetRegion(), common.GetZone(), common.GetProjectId())
	})

	return cmd
}

func listImportedInstanceIDs(ctx *cli.Context, instanceType, region, zone, projectID string) []string {
	action, err := actionForType(instanceType)
	if err != nil {
		return nil
	}
	if instanceType == "mongodb" && zone != "" {
		return nil
	}
	instances, err := fetchUDACInstances(ctx, action, projectID)
	if err != nil {
		return nil
	}
	out := make([]string, 0, len(instances))
	for _, m := range instances {
		zoneVal := getString(m, "Zone")
		if zone != "" && zoneVal != zone {
			continue
		}
		if region != "" {
			instanceRegion := getString(m, "Region")
			if instanceRegion == "" {
				instanceRegion = regionFromZone(zoneVal)
			}
			if instanceRegion != region {
				continue
			}
		}
		id := firstString(m, "ClusterId", "InstanceId", "ID")
		if id == "" {
			continue
		}
		if name := getString(m, "Name"); name != "" {
			out = append(out, id+"/"+name)
		} else {
			out = append(out, id)
		}
	}
	return out
}

// getString 从 map 中安全获取字符串值
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

// getInt64 从 map 中安全获取 int64 值（兼容 float64/int64/int）
func getInt64(m map[string]interface{}, key string) int64 {
	if val, ok := m[key].(float64); ok {
		return int64(val)
	}
	if val, ok := m[key].(int64); ok {
		return int64(val)
	}
	if val, ok := m[key].(int); ok {
		return int64(val)
	}
	return 0
}
