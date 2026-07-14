// Package ip ...
//
// @Brief  查询国内高防IP列表命令
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

// newList 构建 uddos mainland ip list 命令
//
// @Brief  构建国内高防 ip list 子命令，调用 GetBGPServiceIP
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
	var resourceID, bgpIP string
	var offset, limit int

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List mainland BGP high-protection IPs",
		Long:    "List BGP DDoS protection IP addresses for a mainland service instance",
		Example: "  ucloud uddos mainland ip list --resource-id ghp-xxxxx",
		Run: func(cmd *cobra.Command, args []string) {
			client := cli.NewServiceClient(ctx, uaccount.NewClient)
			params := map[string]interface{}{
				"Action":     "GetBGPServiceIP",
				"ResourceId": resourceID,
				"Offset":     offset,
				"Limit":      limit,
			}
			if bgpIP != "" {
				params["BgpIP"] = bgpIP
			}
			req := client.NewGenericRequest()
			if err := req.SetPayload(params); err != nil {
				ctx.HandleError(fmt.Errorf("set payload: %w", err))
				return
			}
			resp, err := client.GenericInvoke(req)
			if err != nil {
				ctx.HandleError(fmt.Errorf("GetBGPServiceIP: %w", err))
				return
			}
			payload := resp.GetPayload()
			gameIPInfo, _ := payload["GameIPInfo"].([]interface{})
			rows := make([]IPRow, 0, len(gameIPInfo))
			for _, item := range gameIPInfo {
				m, ok := item.(map[string]interface{})
				if !ok {
					continue
				}
				rows = append(rows, IPRow{
					DefenceIP:           strVal(m, "DefenceIP"),
					UserIP:              strVal(m, "UserIP"),
					LineType:            strVal(m, "LineType"),
					Status:              strVal(m, "Status"),
					Cname:               strVal(m, "Cname"),
					RuleCnt:             intVal(m, "RuleCnt"),
					DefenceDDosBaseFlow: intVal(m, "DefenceDDosBaseFlow"),
					DefenceDDosMaxFlow:  intVal(m, "DefenceDDosMaxFlow"),
					Remark:              strVal(m, "Remark"),
				})
			}
			ctx.PrintList(rows)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&resourceID, "resource-id", "", "Required. Service resource ID")
	flags.StringVar(&bgpIP, "bgp-ip", "", "Optional. Filter by BGP IP address")
	flags.IntVar(&offset, "offset", 0, "Optional. Page offset, default 0")
	flags.IntVar(&limit, "limit", 20, "Optional. Page size, default 20")
	cmd.MarkFlagRequired("resource-id")
	return cmd
}
