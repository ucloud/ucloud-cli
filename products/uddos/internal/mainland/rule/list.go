// Package rule ...
//
// @Brief  查询国内高防转发规则列表命令
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
package rule

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/ucloud/ucloud-sdk-go/services/uaccount"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newList 构建 uddos mainland rule list 命令
//
// @Brief  构建国内高防 rule list 子命令，调用 GetNapFwdRule
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
	var ruleIndex, offset, limit int

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List BGP forwarding rules",
		Long:    "List forwarding rules for a mainland BGP DDoS protection service",
		Example: "  ucloud uddos mainland rule list --resource-id ghp-xxxxx",
		Run: func(cmd *cobra.Command, args []string) {
			client := cli.NewServiceClient(ctx, uaccount.NewClient)
			params := map[string]interface{}{
				"Action":     "GetBGPServiceFwdRule",
				"ResourceId": resourceID,
				"Offset":     offset,
				"Limit":      limit,
			}
			if cmd.Flags().Changed("rule-index") {
				params["RuleIndex"] = ruleIndex
			}
			if bgpIP != "" {
				params["NapIP"] = bgpIP
			}
			req := client.NewGenericRequest()
			if err := req.SetPayload(params); err != nil {
				ctx.HandleError(fmt.Errorf("set payload: %w", err))
				return
			}
			resp, err := client.GenericInvoke(req)
			if err != nil {
				ctx.HandleError(fmt.Errorf("GetBGPServiceFwdRule: %w", err))
				return
			}
			payload := resp.GetPayload()
			ruleInfo, _ := payload["RuleInfo"].([]interface{})
			rows := make([]RuleRow, 0, len(ruleInfo))
			for _, item := range ruleInfo {
				m, ok := item.(map[string]interface{})
				if !ok {
					continue
				}
				rows = append(rows, RuleRow{
					RuleIndex:    strconv.Itoa(intVal(m, "RuleIndex")),
					RuleID:       strVal(m, "RuleID"),
					BgpIP:        strVal(m, "BgpIP"),
					SourceIP:     strVal(m, "SourceIPInfo"),
					FwdType:      strVal(m, "FwdType"),
					BgpIPPort:    strconv.Itoa(intVal(m, "BgpIPPort")),
					LoadBalance:  strVal(m, "LoadBalance"),
					SourceDetect: strconv.Itoa(intVal(m, "SourceDetect")),
					Remark:       strVal(m, "Remark"),
				})
			}
			ctx.PrintList(rows)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&resourceID, "resource-id", "", "Required. Service resource ID")
	flags.StringVar(&bgpIP, "bgp-ip", "", "Optional. Filter by BGP IP address")
	flags.IntVar(&ruleIndex, "rule-index", 0, "Optional. Filter by rule index")
	flags.IntVar(&offset, "offset", 0, "Optional. Page offset, default 0")
	flags.IntVar(&limit, "limit", 32, "Optional. Page size, default 32")
	cmd.MarkFlagRequired("resource-id")
	return cmd
}
