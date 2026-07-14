// Package rule ...
//
// @Brief  创建国内高防转发规则命令
//
// @File   create.go
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

// newCreate 构建 uddos mainland rule create 命令
//
// @Brief  构建国内高防 rule create 子命令，调用 CreateBGPServiceFwdRule
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
func newCreate(ctx *cli.Context) *cobra.Command {
	var resourceID, sourceIP, bgpIP, loadBalance, fwdType, remark string
	var bgpIPPort, sourceDetect int

	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Create a BGP forwarding rule",
		Long:    "Create a forwarding rule for a mainland BGP DDoS protection service",
		Example: "  ucloud uddos mainland rule create --resource-id ghp-xxxxx --bgp-ip 103.216.x.x --source-ip 10.0.0.1",
		Run: func(cmd *cobra.Command, args []string) {
			client := cli.NewServiceClient(ctx, uaccount.NewClient)
			params := map[string]interface{}{
				"Action":       "CreateBGPServiceFwdRule",
				"ResourceId":   resourceID,
				"BgpIP":        bgpIP,
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
			if cmd.Flags().Changed("remark") {
				params["Remark"] = remark
			}
			req := client.NewGenericRequest()
			if err := req.SetPayload(params); err != nil {
				ctx.HandleError(fmt.Errorf("set payload: %w", err))
				return
			}
			resp, err := client.GenericInvoke(req)
			if err != nil {
				ctx.HandleError(fmt.Errorf("CreateBGPServiceFwdRule: %w", err))
				return
			}
			payload := resp.GetPayload()
			ruleIndex := intVal(payload, "RuleIndex")
			fmt.Fprintf(ctx.ProgressWriter(), "rule[%d] created for service[%s]\n", ruleIndex, resourceID)
			ctx.EmitResult(cli.OpResultRow{
				ResourceID: strconv.Itoa(ruleIndex),
				Action:     "create",
				Status:     "Created",
			})
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&resourceID, "resource-id", "", "Required. Service resource ID")
	flags.StringVar(&bgpIP, "bgp-ip", "", "Required. BGP IP address for this rule")
	flags.StringVar(&sourceIP, "source-ip", "", "Required. Origin server IP address")
	flags.StringVar(&loadBalance, "load-balance", "No", "Optional. Enable load balance: Yes or No, default No")
	flags.StringVar(&fwdType, "fwd-type", "IP", "Optional. Forwarding protocol: IP, TCP or UDP, default IP")
	flags.IntVar(&bgpIPPort, "bgp-ip-port", 0, "Optional. BGP IP port (0 for IP protocol)")
	flags.IntVar(&sourceDetect, "source-detect", 0, "Optional. Source detection: 0=disabled, 1=enabled, default 0")
	flags.StringVar(&remark, "remark", "", "Optional. Remark for this rule")
	cmd.MarkFlagRequired("resource-id")
	cmd.MarkFlagRequired("bgp-ip")
	cmd.MarkFlagRequired("source-ip")
	return cmd
}
