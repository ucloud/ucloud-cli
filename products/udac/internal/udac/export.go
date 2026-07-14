package udac

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ucloud/ucloud-sdk-go/services/uaccount"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// actionForExportType 根据实例类型返回从 UDAC 移除对应的 Action 名。
func actionForExportType(instanceType string) (string, error) {
	switch instanceType {
	case "mysql":
		return "RemoveUDACMySQLInstances", nil
	case "mongodb":
		return "RemoveUDACUMongoDBClusters", nil
	default:
		return "", fmt.Errorf("unsupported instance type: %s, supported: %v", instanceType, SupportedTypes)
	}
}

// exportInstances 调用 UDAC 移除 MySQL API，所有实例共用同一个 zone。
// 走 SDK 默认 FormEncoder（扁平化格式），MySQL 后端接受。
func exportInstances(ctx *cli.Context, action, projectID, zone string, instanceIDs []string) (map[string]interface{}, error) {
	instanceInfoSet := make([]interface{}, 0, len(instanceIDs))
	for _, id := range instanceIDs {
		instanceInfoSet = append(instanceInfoSet, map[string]interface{}{
			"ID":   id,
			"Zone": zone,
		})
	}

	params := map[string]interface{}{
		"Action":          action,
		"ProjectId":       projectID,
		"InstanceInfoSet": instanceInfoSet,
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
	return resp.GetPayload(), nil
}

func exportMongoDBClusters(ctx *cli.Context, action, projectID, region string, clusterIDs []string) (map[string]interface{}, error) {
	mongoDBClusterSet := make([]interface{}, 0, len(clusterIDs))
	for _, id := range clusterIDs {
		mongoDBClusterSet = append(mongoDBClusterSet, map[string]interface{}{
			"ClusterId": id,
			"Region":    region,
		})
	}

	params := map[string]interface{}{
		"Action":            action,
		"ProjectId":         projectID,
		"Region":            region,
		"MongoDBClusterSet": mongoDBClusterSet,
	}

	client := cli.NewServiceClient(ctx, uaccount.NewClient)
	req := client.NewGenericRequest()
	if err := req.SetPayload(params); err != nil {
		return nil, fmt.Errorf("set payload: %w", err)
	}
	req.SetEncoder(request.NewJSONEncoder(client.GetConfig(), client.GetCredential()))
	resp, err := client.GenericInvoke(req)
	if err != nil {
		return nil, err
	}
	return resp.GetPayload(), nil
}

// newExport implements `ucloud udac export`
// MySQL 必填：--udb-id, --type=mysql, --zone, --project-id
// MongoDB 必填：--udb-id, --type=mongodb, --region, --project-id
// export 是从 UDAC 移除实例（不是导出数据），语义同 import 的反向操作。
func newExport(ctx *cli.Context) *cobra.Command {
	var instanceIDs []string
	var instanceType string
	var common request.CommonBase

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export database instances from UDAC",
		Long: `Export existing database instances from the Database Autonomous Center (UDAC).

You must specify the instance type via --type. Supported types: mysql, mongodb.

Required flags:
  mysql:    --udb-id, --type=mysql, --project-id (--zone falls back to config default)
  mongodb:  --udb-id, --type=mongodb, --project-id (--region falls back to config default)

--project-id, --region, --zone fall back to config defaults (default-project-id,
default-region, default-zone).

This is a synchronous operation: the command returns after the export API responds.`,
		Run: func(c *cobra.Command, args []string) {
			// 1. 开头单独校验 udb-id 和 type（早失败）
			if len(instanceIDs) == 0 {
				ctx.HandleError(fmt.Errorf("required flag(s) not set: %s", resourceIDFlag))
				return
			}
			if instanceType == "" {
				ctx.HandleError(fmt.Errorf("required flag(s) not set: type"))
				return
			}

			// 2. 类型校验 + Action 选择
			action, err := actionForExportType(instanceType)
			if err != nil {
				ctx.HandleError(err)
				return
			}

			// 3. 从 common 取绑定值
			projectID := common.GetProjectId()
			region := common.GetRegion()
			zone := common.GetZone()

			// 4. 类型相关必填校验（配置默认值兜底，空时报错）
			var missing []string
			if projectID == "" {
				missing = append(missing, "project-id")
			}
			if instanceType == "mysql" && zone == "" {
				missing = append(missing, "zone")
			}
			if instanceType == "mongodb" && region == "" {
				missing = append(missing, "region")
			}
			if len(missing) > 0 {
				ctx.HandleError(fmt.Errorf("required flag(s) not set: %s", strings.Join(missing, ", ")))
				return
			}

			// 5. 归一化 instance ID（支持 "udb-xxx/instance-name" 格式）
			for i, id := range instanceIDs {
				instanceIDs[i] = ctx.PickResourceID(id)
			}

			// 6. 按类型分流调用 API
			var payload map[string]interface{}
			if instanceType == "mongodb" {
				payload, err = exportMongoDBClusters(ctx, action, projectID, region, instanceIDs)
			} else {
				payload, err = exportInstances(ctx, action, projectID, zone, instanceIDs)
			}
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if len(payload) == 0 {
				ctx.HandleError(fmt.Errorf("empty response from server"))
				return
			}

			// 7. 输出
			w := ctx.ProgressWriter()
			for _, id := range instanceIDs {
				if instanceType == "mongodb" {
					fmt.Fprintf(w, "%s[%s] exported successfully (type: %s, region: %s)\n", productName, id, instanceType, region)
				} else {
					fmt.Fprintf(w, "%s[%s] exported successfully (type: %s, zone: %s)\n", productName, id, instanceType, zone)
				}
			}
			ctx.EmitResult(cli.OpResultRow{
				ResourceID: instanceIDs[0],
				Action:     "export",
				Status:     "Exported",
			})
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&instanceIDs, resourceIDFlag, nil, "Required. Instance ID(s) to export. Repeatable.")
	flags.StringVar(&instanceType, typeFlag, "", "Required. Instance type: mysql, mongodb.")

	// 公共参数绑定：project-id/region/zone 都用配置默认值兜底
	ctx.BindRegion(cmd, &common)
	ctx.BindZone(cmd, &common)
	ctx.BindProjectID(cmd, &common)

	command.SetFlagValues(cmd, typeFlag, SupportedTypes...)

	return cmd
}
