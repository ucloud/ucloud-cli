// Package service ...
//
// @Brief  查询国内高防服务列表命令
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

// newList 构建 uddos mainland service list 命令
//
// @Brief  构建国内高防 service list 子命令，调用 DescribeHighProtectGameServiceInfo
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
		Short: "List mainland DDoS protection service instances",
		Long:  "List mainland China DDoS high-protection service instances via DescribeHighProtectGameServiceInfo.",
		Example: `  # List all mainland services
  ucloud uddos mainland service list

  # Filter by resource ID
  ucloud uddos mainland service list --resource-id ghp-xxxxx`,
		Run: func(cmd *cobra.Command, args []string) {
			client := cli.NewServiceClient(ctx, uaccount.NewClient)
			params := map[string]interface{}{
				"Action": "DescribeHighProtectGameServiceInfo",
				"Offset": offset,
				"Limit":  limit,
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
				ctx.HandleError(fmt.Errorf("DescribeHighProtectGameServiceInfo: %w", err))
				return
			}
			payload := resp.GetPayload()
			gameInfo, _ := payload["GameInfo"].([]interface{})
			rows := make([]ServiceRow, 0, len(gameInfo))
			for _, item := range gameInfo {
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
					Name:          strVal(m, "HighProtectGameServiceName"),
					DefenceStatus: strVal(m, "DefenceStatus"),
					ExpireTime:    expireTime,
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
