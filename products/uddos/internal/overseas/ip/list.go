// Package ip ...
//
// @Brief  查询海外高防IP列表命令
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
package ip

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/ucloud/ucloud-sdk-go/services/uaccount"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newList 构建 uddos overseas ip list 命令
//
// @Brief  构建海外高防 ip list 子命令，调用 GetBGPServiceIP
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
	var resourceID, napIP string
	var offset, limit int

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List overseas BGP high-protection IPs",
		Long:    "List BGP DDoS protection IP addresses for an overseas service instance (Passthrough mode)",
		Example: "  ucloud uddos overseas ip list --resource-id nap-xxxxx",
		Run: func(cmd *cobra.Command, args []string) {
			client := cli.NewServiceClient(ctx, uaccount.NewClient)
			params := map[string]interface{}{
				"Action":     "DescribePassthroughNapIP",
				"ResourceId": resourceID,
				"Offset":     offset,
				"Limit":      limit,
			}
			if napIP != "" {
				params["NapIp"] = napIP
			}
			req := client.NewGenericRequest()
			if err := req.SetPayload(params); err != nil {
				ctx.HandleError(fmt.Errorf("set payload: %w", err))
				return
			}
			resp, err := client.GenericInvoke(req)
			if err != nil {
				ctx.HandleError(fmt.Errorf("DescribePassthroughNapIP: %w", err))
				return
			}
			payload := resp.GetPayload()
			ipInfo, _ := payload["IPInfo"].([]interface{})
			rows := make([]IPRow, 0, len(ipInfo))
			for _, item := range ipInfo {
				m, ok := item.(map[string]interface{})
				if !ok {
					continue
				}
				eipIP := ""
				if addrs, ok := m["EIPAddr"].([]interface{}); ok && len(addrs) > 0 {
					if first, ok := addrs[0].(map[string]interface{}); ok {
						eipIP = strVal(first, "IP")
					}
				}
				rows = append(rows, IPRow{
					EIPIP:     eipIP,
					EIPID:     strVal(m, "EIPId"),
					Status:    strVal(m, "Status"),
					EIPRegion: strVal(m, "EIPRegion"),
					Tag:       strVal(m, "Tag"),
					Remark:    strVal(m, "Remark"),
				})
			}
			ctx.PrintList(rows)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&resourceID, "resource-id", "", "Required. Service resource ID")
	flags.StringVar(&napIP, "nap-ip", "", "Optional. Filter by NAP IP address")
	flags.IntVar(&offset, "offset", 0, "Optional. Page offset, default 0")
	flags.IntVar(&limit, "limit", 20, "Optional. Page size, default 20")
	cmd.MarkFlagRequired("resource-id")
	return cmd
}
