// Package rule ...
//
// @Brief  更新国内高防转发规则命令
//
// @File   update.go
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

// newUpdate 构建 uddos mainland rule update 命令
//
// @Brief  构建国内高防 rule update 子命令，调用 UpdateBGPServiceFwdRule
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
func newUpdate(ctx *cli.Context) *cobra.Command {
	var resourceID, sourceIP, bgpIP, loadBalance, fwdType, ruleID string
	var ruleIndex, bgpIPPort, sourceDetect int

	cmd := &cobra.Command{
		Use:     "update",
		Short:   "Update a BGP forwarding rule",
		Long:    "Update an existing forwarding rule in a mainland BGP DDoS protection service",
		Example: "  ucloud uddos mainland rule update --resource-id ghp-xxxxx --bgp-ip 1.2.3.4 --rule-index 0 --source-ip 10.0.0.2",
		Run: func(cmd *cobra.Command, args []string) {
			client := cli.NewServiceClient(ctx, uaccount.NewClient)
			params := map[string]interface{}{
				"Action":       "UpdateBGPServiceFwdRule",
				"ResourceId":   resourceID,
				"BgpIP":        bgpIP,
				"RuleIndex":    ruleIndex,
				"LoadBalance":  loadBalance,
				"FwdType":      fwdType,
				"BgpIPPort":    bgpIPPort,
				"SourceDetect": sourceDetect,
			}
			if cmd.Flags().Changed("source-ip") {
				params["SourceAddrArr"] = []string{sourceIP}
				params["SourcePortArr"] = []string{"0"}
				params["SourceToaIDArr"] = []string{"0"}
			}
			if cmd.Flags().Changed("rule-id") {
				params["RuleID"] = ruleID
			}
			req := client.NewGenericRequest()
			if err := req.SetPayload(params); err != nil {
				ctx.HandleError(fmt.Errorf("set payload: %w", err))
				return
			}
			_, err := client.GenericInvoke(req)
			if err != nil {
				ctx.HandleError(fmt.Errorf("UpdateBGPServiceFwdRule: %w", err))
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "rule[%d] updated for service[%s]\n", ruleIndex, resourceID)
			ctx.EmitResult(cli.OpResultRow{
				ResourceID: strconv.Itoa(ruleIndex),
				Action:     "update",
				Status:     "Updated",
			})
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&resourceID, "resource-id", "", "Required. Service resource ID")
	flags.StringVar(&bgpIP, "bgp-ip", "", "Required. BGP IP address of the rule")
	flags.IntVar(&ruleIndex, "rule-index", 0, "Required. Rule index to update")
	flags.StringVar(&sourceIP, "source-ip", "", "Optional. New origin server IP address")
	flags.StringVar(&ruleID, "rule-id", "", "Optional. Rule ID (alternative to rule-index)")
	flags.StringVar(&loadBalance, "load-balance", "No", "Optional. Enable load balance: Yes or No")
	flags.StringVar(&fwdType, "fwd-type", "IP", "Optional. Forwarding protocol: IP, TCP or UDP")
	flags.IntVar(&bgpIPPort, "bgp-ip-port", 0, "Optional. BGP IP port")
	flags.IntVar(&sourceDetect, "source-detect", 0, "Optional. Source detection: 0=disabled, 1=enabled")
	cmd.MarkFlagRequired("resource-id")
	cmd.MarkFlagRequired("bgp-ip")
	cmd.MarkFlagRequired("rule-index")
	return cmd
}
