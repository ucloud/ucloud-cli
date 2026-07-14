// Package ip ...
//
// @Brief  海外高防IP管理命令聚合
//
// @File   cmd.go
//
// @Author leas.li(cc)
//
// @Email  leas.li@ucloud.cn
//
// @Date   2026/07/11
//
// @CopyRights(C) UCloud All rights reserved.
package ip

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"
	"github.com/ucloud/ucloud-sdk-go/ucloud/response"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// NewCommand 构建 uddos overseas ip 命令组
//
// @Brief  构建海外高防 ip 命令组并挂载子命令
//
// @Param  ctx *cli.Context
//
// @Return *cobra.Command
//
// @Author leas.li(cc)
//
// @Email  leas.li@ucloud.cn
//
// @Date   2026/07/11
func NewCommand(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ip",
		Short: "Manage overseas DDoS protection IPs",
		Long:  "List, create, delete, bind and unbind overseas BGP DDoS protection IPs",
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(newList(ctx))
	cmd.AddCommand(newCreate(ctx))
	cmd.AddCommand(newDelete(ctx))
	// cmd.AddCommand(newBind(ctx))
	// cmd.AddCommand(newUnbind(ctx))
	return cmd
}

// napInvoker 定义 bind/unbind 所需的最小 SDK 客户端接口
type napInvoker interface {
	NewGenericRequest() request.GenericRequest
	GenericInvoke(request.GenericRequest) (response.GenericResponse, error)
}

// checkOverseasService 校验 resourceID 对应的高防服务是否为海外高防（NapType=2）
//
// @Brief  调用 DescribeNapServiceInfo 查询服务类型，非海外高防时返回错误
//
// @Param  client napInvoker
//
// @Param  resourceID string 高防服务资源 ID
//
// @Return error
//
// @Author leas.li(cc)
//
// @Email  leas.li@ucloud.cn
//
// @Date   2026/07/11
func checkOverseasService(client napInvoker, resourceID string) error {
	params := map[string]interface{}{
		"Action":     "DescribeNapServiceInfo",
		"ResourceId": resourceID,
		"NapType":    2,
		"Offset":     0,
		"Limit":      1,
	}
	req := client.NewGenericRequest()
	if err := req.SetPayload(params); err != nil {
		return fmt.Errorf("check service type set payload: %w", err)
	}
	resp, err := client.GenericInvoke(req)
	if err != nil {
		return fmt.Errorf("DescribeNapServiceInfo: %w", err)
	}
	payload := resp.GetPayload()
	serviceInfo, _ := payload["ServiceInfo"].([]interface{})
	if len(serviceInfo) == 0 {
		return fmt.Errorf("resource %q is not an overseas DDoS service; ip bind/unbind only supports overseas (NapType=2)", resourceID)
	}
	return nil
}

func strVal(m map[string]interface{}, key string) string {
	v, _ := m[key].(string)
	return v
}

func intVal(m map[string]interface{}, key string) int {
	v, _ := m[key].(float64)
	return int(v)
}
