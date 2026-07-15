// Package service ...
//
// @Brief  查询海外高防服务列表命令
//
// @File   list.go
//
// @Author leas.li(cc)
//
// @Email  leas.li@ucloud.cn
//
// @Date   2026/07/11
//
// @CopyRights(C) UCloud All rights reserved.
package service

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/ucloud/ucloud-sdk-go/services/uaccount"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newList 构建 uddos overseas service list 命令
//
// @Brief  构建海外高防 service list 子命令，调用 DescribeNapServiceInfo（NapType=2）
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
func newList(ctx *cli.Context) *cobra.Command {
	var resourceID string
	var offset, limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List overseas DDoS protection service instances",
		Long:  "List overseas DDoS high-protection service instances via DescribeNapServiceInfo (NapType=2).",
		Example: `  # List all overseas services
  ucloud uddos overseas service list

  # Filter by resource ID
  ucloud uddos overseas service list --resource-id nap-xxxxx`,
		Run: func(cmd *cobra.Command, args []string) {
			client := cli.NewServiceClient(ctx, uaccount.NewClient)
			params := map[string]interface{}{
				"Action":  "DescribeNapServiceInfo",
				"NapType": 2,
				"Offset":  offset,
				"Limit":   limit,
			}
			if resourceID != "" {
				params["ResourceId"] = resourceID
			}
			req := client.NewGenericRequest()
			if err := req.SetPayload(params); err != nil {
				ctx.HandleError(fmt.Errorf("set payload: %w", err))
				return
			}
			resp, err := client.GenericInvoke(req)
			if err != nil {
				ctx.HandleError(fmt.Errorf("DescribeNapServiceInfo: %w", err))
				return
			}
			payload := resp.GetPayload()
			serviceInfo, _ := payload["ServiceInfo"].([]interface{})
			rows := make([]ServiceRow, 0, len(serviceInfo))
			for _, item := range serviceInfo {
				m, ok := item.(map[string]interface{})
				if !ok {
					continue
				}
				expireTime := ""
				if ts := intVal(m, "ExpiredTime"); ts > 0 {
					expireTime = time.Unix(int64(ts), 0).Format("2006-01-02 15:04:05")
				}
				rows = append(rows, ServiceRow{
					ResourceID:    strVal(m, "ResourceId"),
					Name:          strVal(m, "Name"),
					DefenceStatus: strVal(m, "DefenceStatus"),
					ExpireTime:    expireTime,
					Remark:        strVal(m, "Remark"),
				})
			}
			ctx.PrintList(rows)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&resourceID, "resource-id", "", "Optional. Filter by resource ID")
	flags.IntVar(&offset, "offset", 0, "Optional. Page offset, default 0")
	flags.IntVar(&limit, "limit", 20, "Optional. Page size, default 20")
	return cmd
}
